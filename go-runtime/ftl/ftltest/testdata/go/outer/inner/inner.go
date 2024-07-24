package inner

import (
	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

//ftl:data export
type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

//ftl:data export
type EchoResponse struct {
	Message string `json:"message"`
}
