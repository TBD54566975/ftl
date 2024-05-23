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

func TestReflectSchemaType(t *testing.T) {
	AllowAnyPackageForTesting = true
	t.Cleanup(func() { AllowAnyPackageForTesting = false })

	Register(WithSumType[MySumType](Variant1{}, Variant2{}))

	v := AllTypesToReflect{SumType: &Variant1{}}

	tests := []struct {
		name     string
		value    any
		expected schema.Type
	}{
		{"Data", &v, &schema.Ref{Module: "reflection", Name: "AllTypesToReflect"}},
		{"SumType", &v.SumType, &schema.Ref{Module: "reflection", Name: "MySumType"}},
		{"Enum", &v.Enum, &schema.Ref{Module: "reflection", Name: "Enum"}},
		{"Int", &v.Int, &schema.Int{}},
		{"String", &v.String, &schema.String{}},
		{"Float", &v.Float, &schema.Float{}},
		{"Any", &v.Any, &schema.Any{}},
		{"Array", &v.Array, &schema.Array{Element: &schema.Int{}}},
		{"Map", &v.Map, &schema.Map{Key: &schema.String{}, Value: &schema.Int{}}},
		{"Bool", &v.Bool, &schema.Bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := reflectSchemaType(reflect.TypeOf(tt.value).Elem())
			assert.Equal(t, tt.expected, st)
		})
	}

	t.Run("InvalidType", func(t *testing.T) {
		var invalid uint
		assert.Panics(t, func() {
			reflectSchemaType(reflect.TypeOf(&invalid).Elem())
		})
	})
}
