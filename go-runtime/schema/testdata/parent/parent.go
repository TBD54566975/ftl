package parent

import (
	"context"
	// Import the FTL SDK.
	"ftl/parent/child"
)

//ftl:verb export
func Verb(ctx context.Context) (child.ChildStruct, error) {
	return child.ChildStruct{}, nil
}
