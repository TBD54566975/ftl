package module

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var configGlobal = ftl.Config[string]("ftlEndpoint")
var configGithub = ftl.Config[string]("githubAccessToken")
var configMissing = ftl.Config[string]("missingConfig")

var secretEncrypt = ftl.Secret[string]("encryptionKey")
var secretMissing = ftl.Secret[string]("missingSecret")

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:export
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
