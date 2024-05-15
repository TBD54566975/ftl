package reflection

import (
	"reflect"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTypeRegistry(t *testing.T) {
	allowAnyPackageForTesting = true
	defer func() { allowAnyPackageForTesting = false }()
	tr := NewTypeRegistry(WithSumType[MySumType](Variant1{}, Variant2{}))

	svariant, ok := tr.GetVariantByType(reflect.TypeFor[MySumType](), reflect.TypeFor[Variant1]()).Get()
	assert.True(t, ok)
	assert.Equal(t, "Variant1", svariant)

	variant, ok := tr.GetVariantByName(reflect.TypeFor[MySumType](), "Variant1").Get()
	assert.True(t, ok)
	assert.Equal(t, reflect.TypeFor[Variant1](), variant)

	ok = tr.IsSumTypeDiscriminator(reflect.TypeFor[MySumType]())
	assert.True(t, ok)

	discriminator, ok := tr.GetDiscriminatorByVariant(reflect.TypeFor[Variant1]()).Get()
	assert.True(t, ok)
	assert.Equal(t, reflect.TypeFor[MySumType](), discriminator)
}
