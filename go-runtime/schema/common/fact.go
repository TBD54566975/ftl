package common

import (
	"go/types"
	"reflect"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
)

// SchemaFact is a fact that associates a schema node with a Go object.
type SchemaFact interface {
	analysis.Fact
	Add(v SchemaFactValue)
	Get() []SchemaFactValue
}

// DefaultFact should be used as the base type for all schema facts. Each
// Analyzer needs a uniuqe Fact type that is otherwise identical, and this type
// simply reduces that boilerplate.
//
// Usage:
//
//	type Fact = common.DefaultFact[struct{}]
type DefaultFact[T any] struct {
	value []SchemaFactValue
}

func (*DefaultFact[T]) AFact() {}
func (t *DefaultFact[T]) Add(v SchemaFactValue) {
	if t.value == nil {
		t.value = []SchemaFactValue{}
	}
	t.value = append(t.value, v)
}
func (t *DefaultFact[T]) Get() []SchemaFactValue { return t.value }

// SchemaFactValue is the value of a SchemaFact.
type SchemaFactValue interface {
	schemaFactValue()
}

// ExtractedDecl is a fact for associating an object with an extracted schema decl.
type ExtractedDecl struct {
	Decl schema.Decl
}

func (*ExtractedDecl) schemaFactValue() {}

// MaybeTypeEnum is a fact for marking an object as a possible type enum discriminator.
type MaybeTypeEnum struct {
	Enum *schema.Enum
}

func (*MaybeTypeEnum) schemaFactValue() {}

// MaybeTypeEnumVariant is a fact for marking an object as a possible type enum variant.
type MaybeTypeEnumVariant struct {
	GetValue func(pass *analysis.Pass) optional.Option[*schema.TypeValue]
	// the parent enum
	Parent types.Object
	// this variant
	Variant *schema.EnumVariant
}

func (*MaybeTypeEnumVariant) schemaFactValue() {}

// MaybeValueEnumVariant is a fact for marking an object as a possible value enum variant.
type MaybeValueEnumVariant struct {
	// this variant
	Variant *schema.EnumVariant
	// type of the variant
	Type types.Object
}

func (*MaybeValueEnumVariant) schemaFactValue() {}

// ExtractedMetadata is a fact for associating an object with extracted schema metadata.
type ExtractedMetadata struct {
	Type       schema.Decl
	IsExported bool
	Metadata   []schema.Metadata
	Comments   []string
}

func (*ExtractedMetadata) schemaFactValue() {}

// NeedsExtraction is a fact for marking a type that needs to be extracted by another extractor.
type NeedsExtraction struct{}

func (*NeedsExtraction) schemaFactValue() {}

// FailedExtraction is a fact for marking a type that failed to be extracted by another extractor.
type FailedExtraction struct{}

func (*FailedExtraction) schemaFactValue() {}

// ExternalType is a fact for marking an external type.
type ExternalType struct{}

func (*ExternalType) schemaFactValue() {}

// FunctionCall is a fact for marking an outbound function call on a function.
type FunctionCall struct {
	// The function being called.
	Callee types.Object
	// Position where the call takes place.
	Position schema.Position
}

func (*FunctionCall) schemaFactValue() {}

// VerbCall is a fact for marking a call to an FTL verb on a function.
type VerbCall struct {
	// The verb being called.
	VerbRef *schema.Ref
}

func (*VerbCall) schemaFactValue() {}

// IncludeNativeName marks a node that needs to be added to the native names map provided in the extraction result.
type IncludeNativeName struct {
	// The schema node associated with this native name.
	Node schema.Node
}

func (*IncludeNativeName) schemaFactValue() {}

// IncludeTopicMapper marks a node as the partition mapper type for a topic.
type IncludeTopicMapper struct {
	Topic *schema.Topic

	// The object for the partition mapper type.
	MapperObject types.Object
	// The object for the associated type for the partition mapper.
	AssociatedObject optional.Option[types.Object]
}

func (*IncludeTopicMapper) schemaFactValue() {}

type DatabaseConfigMethod int

const (
	DatabaseConfigMethodName DatabaseConfigMethod = iota
)

type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeMySQL    DatabaseType = "mysql"
)

// DatabaseConfig marks a database node with an extracted configuration value.
type DatabaseConfig struct {
	Type   DatabaseType
	Method DatabaseConfigMethod
	Value  any
}

func (*DatabaseConfig) schemaFactValue() {}

// VerbResourceParamOrder is a fact for marking the order of resource parameters used by a verb, where the signature
// is (context.Context, <request>, <resource1>, <resource2>, ...).
//
// This is used in the generated code that registers verb resources. The order of parameters is important because it
// will determine the order in which resources are passed to the verb when the call is constructed via reflection at
// runtime.
type VerbResourceParamOrder struct {
	Resources []VerbResourceParam
}

type VerbResourceParam struct {
	Ref  *schema.Ref
	Type schema.Metadata
}

func (*VerbResourceParamOrder) schemaFactValue() {}

// MarkSchemaDecl marks the given object as having been extracted to the given schema decl.
func MarkSchemaDecl(pass *analysis.Pass, obj types.Object, decl schema.Decl) {
	fact := newFact(pass, obj)
	fact.Add(&ExtractedDecl{Decl: decl})
	pass.ExportObjectFact(obj, fact)
}

// MarkFailedExtraction marks the given object as having failed extraction.
func MarkFailedExtraction(pass *analysis.Pass, obj types.Object) {
	fact := newFact(pass, obj)
	fact.Add(&FailedExtraction{})
	pass.ExportObjectFact(obj, fact)
}

func MarkMetadata(pass *analysis.Pass, obj types.Object, md *ExtractedMetadata) {
	fact := newFact(pass, obj)
	fact.Add(md)
	pass.ExportObjectFact(obj, fact)
}

// MarkNeedsExtraction marks the given object as needing extraction.
func MarkNeedsExtraction(pass *analysis.Pass, obj types.Object) {
	fact := newFact(pass, obj)
	fact.Add(&NeedsExtraction{})
	pass.ExportObjectFact(obj, fact)
}

// MarkMaybeTypeEnumVariant marks the given object as a possible type enum variant.
func MarkMaybeTypeEnumVariant(pass *analysis.Pass, obj types.Object, variant *schema.EnumVariant,
	parent types.Object, valueFunc func(pass *analysis.Pass) optional.Option[*schema.TypeValue]) {
	fact := newFact(pass, obj)
	fact.Add(&MaybeTypeEnumVariant{Parent: parent, Variant: variant, GetValue: valueFunc})
	pass.ExportObjectFact(obj, fact)
}

// MarkMaybeValueEnumVariant marks the given object as a possible value enum variant.
func MarkMaybeValueEnumVariant(pass *analysis.Pass, obj types.Object, variant *schema.EnumVariant, typ types.Object) {
	fact := newFact(pass, obj)
	fact.Add(&MaybeValueEnumVariant{Variant: variant, Type: typ})
	pass.ExportObjectFact(obj, fact)
}

// MarkMaybeTypeEnum marks the given object as a possible type enum discriminator.
func MarkMaybeTypeEnum(pass *analysis.Pass, obj types.Object, enum *schema.Enum) {
	fact := newFact(pass, obj)
	fact.Add(&MaybeTypeEnum{Enum: enum})
	pass.ExportObjectFact(obj, fact)
}

// MarkFunctionCall marks the given object as having an outbound function call.
func MarkFunctionCall(pass *analysis.Pass, obj types.Object, callee types.Object, pos schema.Position) {
	fact := newFact(pass, obj)
	fact.Add(&FunctionCall{Callee: callee, Position: pos})
	pass.ExportObjectFact(obj, fact)
}

// MarkIncludeNativeName marks the given object as needing to be added to the native names map.
func MarkIncludeNativeName(pass *analysis.Pass, obj types.Object, node schema.Node) {
	fact := newFact(pass, obj)
	fact.Add(&IncludeNativeName{Node: node})
	pass.ExportObjectFact(obj, fact)
}

// MarkDatabaseConfig marks the given database object with an extracted config value.
func MarkDatabaseConfig(pass *analysis.Pass, obj types.Object, dbType DatabaseType,
	method DatabaseConfigMethod, value any) {
	fact := newFact(pass, obj)
	fact.Add(&DatabaseConfig{Type: dbType, Method: method, Value: value})
	pass.ExportObjectFact(obj, fact)
}

// MarkVerbResourceParamOrder marks the given verb object with the order of its parameters.
func MarkVerbResourceParamOrder(pass *analysis.Pass, obj types.Object, resources []VerbResourceParam) {
	fact := newFact(pass, obj)
	fact.Add(&VerbResourceParamOrder{Resources: resources})
	pass.ExportObjectFact(obj, fact)
}

// MarkTopicMapper marks the given object as the partition mapper for a topic.
func MarkTopicMapper(pass *analysis.Pass, mapperObj types.Object, associatedObj optional.Option[types.Object], topic *schema.Topic) {
	fact := newFact(pass, mapperObj)
	fact.Add(&IncludeTopicMapper{
		Topic:            topic,
		MapperObject:     mapperObj,
		AssociatedObject: associatedObj,
	})
	pass.ExportObjectFact(mapperObj, fact)
}

// GetAllFactsExtractionStatus merges schema facts inclusive of all available results and the present pass facts.
// For a given object, it provides the current extraction status.
//
// If multiple extraction facts are present for the same object, the facts will be prioritized by type:
// 1. ExtractedDecl
// 2. FailedExtraction
// 3. NeedsExtraction
//
// All other fact types are ignored.
func GetAllFactsExtractionStatus(pass *analysis.Pass) map[types.Object]SchemaFactValue {
	facts := make(map[types.Object]SchemaFactValue)
	for _, fact := range allFacts(pass) {
		sf, ok := fact.Fact.(SchemaFact)
		if !ok {
			continue
		}

		// prioritize facts by type
		//
		// e.g. if one extractor marked an object as needing extraction and another extractor marked it with the
		// completed extraction, we should prioritize the completed extraction.
		prioritize := func(v SchemaFactValue) int {
			switch v.(type) {
			case *NeedsExtraction:
				return 1
			case *FailedExtraction:
				return 2
			case *ExtractedDecl:
				return 3
			default:
				return -1
			}
		}
		for _, f := range sf.Get() {
			newPriority := prioritize(f)
			if newPriority == -1 {
				continue
			}

			existing, ok := facts[fact.Object]
			existingPriority := prioritize(existing)
			if !ok || newPriority > existingPriority {
				facts[fact.Object] = f
			}
		}
	}
	return facts
}

// GetAllFactsOfType returns all facts of the provided type marked on objects, across the current pass and results from
// prior passes. If multiple of the same fact type are marked on a single object, the first fact is returned.
func GetAllFactsOfType[T SchemaFactValue](pass *analysis.Pass) map[types.Object][]T {
	return getFactsScoped[T](allFacts(pass))
}

// GetCurrentPassFacts returns all facts of the provided type marked on objects during the current pass.
// If multiple of the same fact type are marked on a single object, the first fact is returned.
func GetCurrentPassFacts[T SchemaFactValue](pass *analysis.Pass) map[types.Object][]T {
	return getFactsScoped[T](pass.AllObjectFacts())
}

func getFactsScoped[T SchemaFactValue](scope []analysis.ObjectFact) map[types.Object][]T {
	facts := make(map[types.Object][]T)
	for _, fact := range scope {
		sf, ok := fact.Fact.(SchemaFact)
		if !ok {
			continue
		}

		for _, f := range sf.Get() {
			if t, ok := f.(T); ok {
				if _, exists := facts[fact.Object]; !exists {
					facts[fact.Object] = []T{t}
				}
				facts[fact.Object] = append(facts[fact.Object], t)
			}
		}
	}
	return facts
}

// GetFactForObject returns the first fact of the provided type marked on the object.
func GetFactForObject[T SchemaFactValue](pass *analysis.Pass, obj types.Object) optional.Option[T] {
	for _, fact := range allFacts(pass) {
		if fact.Object != obj {
			continue
		}
		sf, ok := fact.Fact.(SchemaFact)
		if !ok {
			continue
		}
		for _, f := range sf.Get() {
			if f, ok := f.(T); ok {
				return optional.Some(f)
			}
		}
	}
	return optional.None[T]()
}

// GetFactsForObject returns all facts of the provided type marked on the object.
func GetFactsForObject[T SchemaFactValue](pass *analysis.Pass, obj types.Object) []T {
	facts := []T{}
	for _, fact := range allFacts(pass) {
		if fact.Object != obj {
			continue
		}
		sf, ok := fact.Fact.(SchemaFact)
		if !ok {
			continue
		}
		for _, f := range sf.Get() {
			if f, ok := f.(T); ok {
				facts = append(facts, f)
			}
		}
	}
	return facts
}

func GetAllFacts(pass *analysis.Pass) map[types.Object][]SchemaFactValue {
	facts := make(map[types.Object][]SchemaFactValue)
	for _, fact := range allFacts(pass) {
		sf, ok := fact.Fact.(SchemaFact)
		if !ok {
			continue
		}
		facts[fact.Object] = sf.Get()
	}
	return facts
}

func allFacts(pass *analysis.Pass) []analysis.ObjectFact {
	var all []analysis.ObjectFact
	all = append(all, pass.AllObjectFacts()...)
	for _, result := range pass.ResultOf {
		r, ok := result.(ExtractorResult)
		if !ok {
			continue
		}
		all = append(all, r.Facts...)
	}
	return all
}

func newFact(pass *analysis.Pass, obj types.Object) SchemaFact {
	existing := optional.None[SchemaFact]()
	for _, fact := range pass.AllObjectFacts() {
		if fact.Object != obj {
			continue
		}
		if sf, ok := fact.Fact.(SchemaFact); ok {
			existing = optional.Some(sf)
		}
	}

	fact, ok := existing.Get()
	if !ok {
		factType := reflect.TypeOf(pass.Analyzer.FactTypes[0]).Elem()
		fact = reflect.New(factType).Interface().(SchemaFact) //nolint:forcetypeassert
	}
	return fact
}
