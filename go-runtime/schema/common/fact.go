package common

import (
	"go/types"
	"reflect"

	"github.com/alecthomas/types/optional"
	sets "github.com/deckarep/golang-set/v2"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/golang-tools/go/analysis"
)

// SchemaFact is a fact that associates a schema node with a Go object.
type SchemaFact interface {
	analysis.Fact
	Set(v SchemaFactValue)
	Get() SchemaFactValue
}

// DefaultFact should be used as the base type for all schema facts. Each
// Analyzer needs a uniuqe Fact type that is otherwise identical, and this type
// simply reduces that boilerplate.
//
// Usage:
//
//	type Fact = common.DefaultFact[struct{}]
type DefaultFact[T any] struct {
	value SchemaFactValue
}

func (*DefaultFact[T]) AFact()                  {}
func (t *DefaultFact[T]) Set(v SchemaFactValue) { t.value = v }
func (t *DefaultFact[T]) Get() SchemaFactValue  { return t.value }

// SchemaFactValue is the value of a SchemaFact.
type SchemaFactValue interface {
	schemaFactValue()
}

// ExtractedDecl is a fact for associating an object with an extracted schema decl.
type ExtractedDecl struct {
	Decl schema.Decl
	// ShouldInclude is true if the object should be included in the schema.
	// We extract all objects by default, but some objects may not actually be referenced in the schema.
	ShouldInclude bool
	// Refs is a list of objects that the object references.
	Refs sets.Set[types.Object]
}

func (*ExtractedDecl) schemaFactValue() {}

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

// MarkSchemaDecl marks the given object as having been extracted to the given schema node.
func MarkSchemaDecl(pass *analysis.Pass, obj types.Object, decl schema.Decl) {
	fact := newFact(pass)
	fact.Set(&ExtractedDecl{Decl: decl, Refs: sets.NewSet[types.Object]()})
	pass.ExportObjectFact(obj, fact)
}

// markSchemaDeclIncluded marks the given decl as included in the schema.
func markSchemaDeclIncluded(pass *analysis.Pass, obj types.Object) {
	for _, f := range GetFactsForObject[*ExtractedDecl](pass, obj) {
		f.ShouldInclude = true
	}
}

// MarkFailedExtraction marks the given object as having failed extraction.
func MarkFailedExtraction(pass *analysis.Pass, obj types.Object) {
	fact := newFact(pass)
	fact.Set(&FailedExtraction{})
	pass.ExportObjectFact(obj, fact)
}

func MarkMetadata(pass *analysis.Pass, obj types.Object, md *ExtractedMetadata) {
	fact := newFact(pass)
	fact.Set(md)
	pass.ExportObjectFact(obj, fact)
}

// markNeedsExtraction marks the given object as needing extraction.
func markNeedsExtraction(pass *analysis.Pass, obj types.Object) {
	fact := newFact(pass)
	fact.Set(&NeedsExtraction{})
	pass.ExportObjectFact(obj, fact)
}

// MergeAllFacts merges schema facts inclusive of all available results and the present pass facts.
//
// If multiple facts are present for the same object, the facts will be prioritized by type:
// 1. ExtractedDecl
// 2. FailedExtraction
// 4. NeedsExtraction
//
// ExtractedMetadata facts are ignored.
func MergeAllFacts(pass *analysis.Pass) map[types.Object]SchemaFact {
	facts := make(map[types.Object]SchemaFact)
	for _, fact := range allFactsForPass(pass) {
		f, ok := fact.Fact.(SchemaFact)
		if !ok {
			continue
		}

		// skip metadata facts
		if _, ok = f.Get().(*ExtractedMetadata); ok {
			continue
		}

		// prioritize facts by type
		//
		// e.g. if one extractor marked an object as needing extraction and another extractor marked it with the
		// completed extraction, we should prioritize the completed extraction.
		prioritize := func(f SchemaFact) int {
			switch f.Get().(type) {
			case *ExtractedDecl:
				return 1
			case *FailedExtraction:
				return 2
			case *NeedsExtraction:
				return 3
			default:
				return 4
			}
		}
		existing, ok := facts[fact.Object]
		if !ok || prioritize(f) < prioritize(existing) {
			facts[fact.Object] = f
		}
	}
	return facts
}

func GetFact[T SchemaFactValue](facts []SchemaFact) optional.Option[T] {
	for _, fact := range facts {
		if f, ok := fact.Get().(T); ok {
			return optional.Some(f)
		}
	}
	return optional.None[T]()
}

// GetFactsForObject returns all facts marked on the object.
func GetFactsForObject[T SchemaFactValue](pass *analysis.Pass, obj types.Object) []T {
	var facts []T
	for _, fact := range allFactsForPass(pass) {
		if fact.Object != obj {
			continue
		}
		sf, ok := fact.Fact.(SchemaFact)
		if !ok {
			continue
		}
		if f, ok := sf.Get().(T); ok {
			facts = append(facts, f)
		}
	}
	return facts
}

func GetFacts[T SchemaFactValue](pass *analysis.Pass) map[types.Object]T {
	facts := make(map[types.Object]T)
	for _, fact := range allFactsForPass(pass) {
		sf, ok := fact.Fact.(SchemaFact)
		if !ok {
			continue
		}
		if f, ok := sf.Get().(T); ok {
			facts[fact.Object] = f
		}
	}
	return facts
}

// GetFactForObject returns the first fact of the provided type marked on the object.
func GetFactForObject[T SchemaFactValue](pass *analysis.Pass, obj types.Object) optional.Option[T] {
	for _, fact := range allFactsForPass(pass) {
		if fact.Object != obj {
			continue
		}
		sf, ok := fact.Fact.(SchemaFact)
		if !ok {
			continue
		}
		if f, ok := sf.Get().(T); ok {
			return optional.Some(f)
		}
	}
	return optional.None[T]()
}

func allFactsForPass(pass *analysis.Pass) []analysis.ObjectFact {
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

func newFact(pass *analysis.Pass) SchemaFact {
	factType := reflect.TypeOf(pass.Analyzer.FactTypes[0]).Elem()
	return reflect.New(factType).Interface().(SchemaFact) //nolint:forcetypeassert
}
