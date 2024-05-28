package reflection

import (
	"reflect"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/schema"
)

type MySumType interface{ sealed() }

type Variant1 struct{ Field1 string }

func (Variant1) sealed() {}

type Variant2 struct{ Field2 int }

func (Variant2) sealed() {}

type Enum int

const (
	Enum1 Enum = iota
)

type AllTypesToReflect struct {
	SumType MySumType
	Enum    Enum
	Bool    bool
	Int     int
	Float   float64
	String  string
	Any     any
	Array   []int
	Map     map[string]int
}

func TestReflectTypeFromValue(t *testing.T) {
	AllowAnyPackageForTesting = true
	t.Cleanup(func() { AllowAnyPackageForTesting = false })

	Register(SumType[MySumType](Variant1{}, Variant2{}))

	v := AllTypesToReflect{SumType: Variant1{}}

	tests := []struct {
		name     string
		value    schema.Type
		expected schema.Type
	}{
		{"Data", TypeFromValue(&v), &schema.Ref{Module: "reflection", Name: "AllTypesToReflect"}},
		{"SumType", TypeFromValue(&v.SumType), &schema.Ref{Module: "reflection", Name: "MySumType"}},
		{"Enum", TypeFromValue(&v.Enum), &schema.Ref{Module: "reflection", Name: "Enum"}},
		{"Int", TypeFromValue(&v.Int), &schema.Int{}},
		{"String", TypeFromValue(&v.String), &schema.String{}},
		{"Float", TypeFromValue(&v.Float), &schema.Float{}},
		{"Any", TypeFromValue(&v.Any), &schema.Any{}},
		{"Array", TypeFromValue(&v.Array), &schema.Array{Element: &schema.Int{}}},
		{"Map", TypeFromValue(&v.Map), &schema.Map{Key: &schema.String{}, Value: &schema.Int{}}},
		{"Bool", TypeFromValue(&v.Bool), &schema.Bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.value)
		})
	}

	t.Run("InvalidType", func(t *testing.T) {
		var invalid uint
		assert.Panics(t, func() {
			ReflectTypeToSchemaType(reflect.TypeOf(&invalid).Elem())
		})
	})
}
