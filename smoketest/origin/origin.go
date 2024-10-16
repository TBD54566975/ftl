package origin

import (
	"context"
	"time"

	"ftl/builtin"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var nonce = ftl.Config[string]("nonce")

//ftl:export
var AgentBroadcast = ftl.Topic[Agent]("agentBroadcast")

type Agent struct {
	ID            int                   `json:"id"`
	Alias         string                `json:"alias"`
	LicenseToKill bool                  `json:"license_to_kill"`
	HiredAt       time.Time             `json:"hired_at"`
	BriefedAt     ftl.Option[time.Time] `json:"briefed_at"`
}

type PostAgentResponse struct {
	ID int `json:"id"`
}

type PostAgentErrorResponse string

//ftl:ingress POST /ingress/agent
func PostAgent(ctx context.Context, req builtin.HttpRequest[Agent, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[PostAgentResponse, PostAgentErrorResponse], error) {
	agent := Agent{
		ID:            req.Body.ID,
		Alias:         req.Body.Alias,
		LicenseToKill: req.Body.LicenseToKill,
		HiredAt:       req.Body.HiredAt,
	}
	err := AgentBroadcast.Publish(ctx, agent)
	if err != nil {
		return builtin.HttpResponse[PostAgentResponse, PostAgentErrorResponse]{
			Status: 500,
			Body:   ftl.None[PostAgentResponse](),
		}, err
	}
	return builtin.HttpResponse[PostAgentResponse, PostAgentErrorResponse]{
		Status: 201,
		Body:   ftl.Some(PostAgentResponse{ID: agent.ID}),
	}, nil
}

// Exported verb

type GetNonceRequest struct{}
type GetNonceResponse struct{ Nonce string }

//ftl:verb export
func GetNonce(ctx context.Context, req GetNonceRequest) (GetNonceResponse, error) {
	return GetNonceResponse{Nonce: nonce.Get(ctx)}, nil
}
