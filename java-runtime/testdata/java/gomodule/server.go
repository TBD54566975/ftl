package gomodule

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type TestObject struct {
	IntField    int
	FloatField  float64
	StringField string
	BytesField  []byte
	BoolField   bool
	TimeField   time.Time
	ArrayField  []string
	MapField    map[string]string
}

type TestObjectOptionalFields struct {
	IntField    ftl.Option[int]
	FloatField  ftl.Option[float64]
	StringField ftl.Option[string]
	BytesField  ftl.Option[[]byte]
	BoolField   ftl.Option[bool]
	TimeField   ftl.Option[time.Time]
	ArrayField  ftl.Option[[]string]
	MapField    ftl.Option[map[string]string]
}

// Test different signatures

//ftl:verb export
func SourceVerb(ctx context.Context) (string, error) {
	return "Source Verb", nil
}

//ftl:verb export
func SinkVerb(ctx context.Context, req string) error {
	return nil
}

//ftl:verb export
func EmptyVerb(ctx context.Context) error {
	return nil
}

//ftl:verb export
func ErrorEmptyVerb(ctx context.Context) error {
	return fmt.Errorf("verb failed")
}

// Test different param and return types

//ftl:verb export
func IntVerb(ctx context.Context, val int) (int, error) {
	return val, nil
}

//ftl:verb export
func FloatVerb(ctx context.Context, val float64) (float64, error) {
	return val, nil
}

//ftl:verb export
func StringVerb(ctx context.Context, val string) (string, error) {
	return val, nil
}

//ftl:verb export
func BytesVerb(ctx context.Context, val []byte) ([]byte, error) {
	return val, nil
}

//ftl:verb export
func BoolVerb(ctx context.Context, val bool) (bool, error) {
	return val, nil
}

//ftl:verb export
func StringArrayVerb(ctx context.Context, val []string) ([]string, error) {
	return val, nil
}

//ftl:verb export
func StringMapVerb(ctx context.Context, val map[string]string) (map[string]string, error) {
	return val, nil
}

//ftl:verb export
func TimeVerb(ctx context.Context, val time.Time) (time.Time, error) {
	return val, nil
}

//ftl:verb export
func TestObjectVerb(ctx context.Context, val TestObject) (TestObject, error) {
	return val, nil
}

//ftl:verb export
func TestObjectOptionalFieldsVerb(ctx context.Context, val TestObjectOptionalFields) (TestObjectOptionalFields, error) {
	return val, nil
}

// Now optional versions of all of the above

//ftl:verb export
func OptionalIntVerb(ctx context.Context, val ftl.Option[int]) (ftl.Option[int], error) {
	return val, nil
}

//ftl:verb export
func OptionalFloatVerb(ctx context.Context, val ftl.Option[float64]) (ftl.Option[float64], error) {
	return val, nil
}

//ftl:verb export
func OptionalStringVerb(ctx context.Context, val ftl.Option[string]) (ftl.Option[string], error) {
	return val, nil
}

//ftl:verb export
func OptionalBytesVerb(ctx context.Context, val ftl.Option[[]byte]) (ftl.Option[[]byte], error) {
	return val, nil
}

//ftl:verb export
func OptionalBoolVerb(ctx context.Context, val ftl.Option[bool]) (ftl.Option[bool], error) {
	return val, nil
}

//ftl:verb export
func OptionalStringArrayVerb(ctx context.Context, val ftl.Option[[]string]) (ftl.Option[[]string], error) {
	return val, nil
}

//ftl:verb export
func OptionalStringMapVerb(ctx context.Context, val ftl.Option[map[string]string]) (ftl.Option[map[string]string], error) {
	return val, nil
}

//ftl:verb export
func OptionalTimeVerb(ctx context.Context, val ftl.Option[time.Time]) (ftl.Option[time.Time], error) {
	return val, nil
}

//ftl:verb export
func OptionalTestObjectVerb(ctx context.Context, val ftl.Option[TestObject]) (ftl.Option[TestObject], error) {
	return val, nil
}

//ftl:verb export
func OptionalTestObjectOptionalFieldsVerb(ctx context.Context, val ftl.Option[TestObjectOptionalFields]) (ftl.Option[TestObjectOptionalFields], error) {
	return val, nil
}
