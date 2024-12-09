package timeline

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

type Ingress struct {
	DeploymentKey   model.DeploymentKey
	RequestKey      model.RequestKey
	StartTime       time.Time
	Verb            *schema.Ref
	RequestMethod   string
	RequestPath     string
	RequestHeaders  http.Header
	ResponseStatus  int
	ResponseHeaders http.Header
	RequestBody     []byte
	ResponseBody    []byte
	Error           optional.Option[string]
}

var _ Event = Ingress{}

func (Ingress) clientEvent() {}
func (i Ingress) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	requestKey := i.RequestKey.String()

	requestBody := i.RequestBody
	if len(requestBody) == 0 {
		requestBody = []byte("{}")
	}

	responseBody := i.ResponseBody
	if len(responseBody) == 0 {
		responseBody = []byte("{}")
	}

	reqHeaderBytes, err := json.Marshal(i.RequestHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request header: %w", err)
	}

	respHeaderBytes, err := json.Marshal(i.ResponseHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response header: %w", err)
	}

	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_Ingress{
			Ingress: &timelinepb.IngressEvent{
				DeploymentKey:  i.DeploymentKey.String(),
				RequestKey:     &requestKey,
				Timestamp:      timestamppb.New(i.StartTime),
				VerbRef:        i.Verb.ToProto().(*schemapb.Ref), //nolint:forcetypeassert
				Method:         i.RequestMethod,
				Path:           i.RequestPath,
				StatusCode:     int32(i.ResponseStatus),
				Duration:       durationpb.New(time.Since(i.StartTime)),
				Request:        string(requestBody),
				RequestHeader:  string(reqHeaderBytes),
				Response:       string(responseBody),
				ResponseHeader: string(respHeaderBytes),
				Error:          i.Error.Ptr(),
			},
		},
	}, nil
}
