package reflection

import (
	"reflect"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTypeRegistry(t *testing.T) {
	AllowAnyPackageForTesting = true
	defer func() { AllowAnyPackageForTesting = false }()
	ResetTypeRegistry()
	Register(WithSumType[MySumType](Variant1{}, Variant2{}))

	svariant, ok := GetVariantByType(reflect.TypeFor[MySumType](), reflect.TypeFor[Variant1]()).Get()
	assert.True(t, ok)
	assert.Equal(t, "Variant1", svariant)

	variant, ok := GetVariantByName(reflect.TypeFor[MySumType](), "Variant1").Get()
	assert.True(t, ok)
	assert.Equal(t, reflect.TypeFor[Variant1](), variant)

	ok = IsSumTypeDiscriminator(reflect.TypeFor[MySumType]())
	assert.True(t, ok)

	discriminator, ok := GetDiscriminatorByVariant(reflect.TypeFor[Variant1]()).Get()
	assert.True(t, ok)
	assert.Equal(t, reflect.TypeFor[MySumType](), discriminator)

	ResetTypeRegistry()
	_, ok = GetVariantByType(reflect.TypeFor[MySumType](), reflect.TypeFor[Variant1]()).Get()
	assert.False(t, ok) // test ResetTypeRegistry()
}
