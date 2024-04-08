package reflect

import (
	"testing"
)

type structOfPointers struct {
	intPtr    *int
	floatPtr  *float64
	structPtr *structOfPointers
}

func TestCopyStructOfPointers(t *testing.T) {
	one := 1
	two := 2
	half := 0.5
	quarter := 0.25

	for _, tt := range []struct {
		name  string
		input structOfPointers
	}{
		{
			"StructWithinStruct",
			structOfPointers{
				intPtr:   &one,
				floatPtr: &half,
				structPtr: &structOfPointers{
					intPtr:    &two,
					floatPtr:  &quarter,
					structPtr: nil,
				},
			},
		},
		{
			"StructWithinStructWithinStruct",
			structOfPointers{
				intPtr:   &one,
				floatPtr: &half,
				structPtr: &structOfPointers{
					intPtr:   &two,
					floatPtr: &quarter,
					structPtr: &structOfPointers{
						intPtr:    &two,
						floatPtr:  &quarter,
						structPtr: nil,
					},
				},
			},
		},
		{
			"StructWithNils",
			structOfPointers{
				intPtr:    nil,
				floatPtr:  nil,
				structPtr: nil,
			},
		},
		{
			"StructWithStructWithNils",
			structOfPointers{
				intPtr:   &one,
				floatPtr: &half,
				structPtr: &structOfPointers{
					intPtr:    nil,
					floatPtr:  nil,
					structPtr: nil,
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output := DeepCopy(tt.input)
			testEqualityOfStruct(t, &tt.input, &output)
		})
	}
}

func testEqualityOfStruct(t *testing.T, expected, actual *structOfPointers) {
	if expected == nil || actual == nil {
		if expected != actual {
			t.Errorf("struct point does not match nil struct pointer. expected %v, got %v", expected, actual)
		}
		return
	}

	testEqualityOfInt(t, expected.intPtr, actual.intPtr)
	testEqualityOfFloat64(t, expected.floatPtr, actual.floatPtr)
	testEqualityOfStruct(t, expected.structPtr, actual.structPtr)
}

func testEqualityOfInt(t *testing.T, expected, actual *int) {
	if expected == nil || actual == nil {
		if expected != actual {
			t.Errorf("int point does not match nil int pointer. expected %v, got %v", expected, actual)
		}
		return
	}
}

func testEqualityOfFloat64(t *testing.T, expected, actual *float64) {
	if expected == nil || actual == nil {
		if expected != actual {
			t.Errorf("float point does not match nil float pointer. expected %v, got %v", expected, actual)
		}
		return
	}
}
