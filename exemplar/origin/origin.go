package origin

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"

	"ftl/builtin"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var nonce = ftl.Config[string]("nonce")

//ftl:export
var AgentBroadcast = ftl.Topic[Agent]("agentBroadcast")

type Agent struct {
	ID            int
	Alias         string
	LicenseToKill bool
	HiredAt       time.Time
	BriefedAt     ftl.Option[time.Time]
}

type PostAgentRequest struct {
	Alias         string                `json:"alias"`
	LicenseToKill bool                  `json:"license_to_kill"`
	HiredAt       time.Time             `json:"hired_at"`
	BriefedAt     ftl.Option[time.Time] `json:"briefed_at"`
}

type PostAgentResponse struct {
	ID int `json:"id"`
}

type PostAgentErrorResponse string

//ftl:ingress POST /http/agent
func PostAgent(ctx context.Context, req builtin.HttpRequest[PostAgentRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[PostAgentResponse, PostAgentErrorResponse], error) {
	agent := Agent{
		ID:            generateRandomID(),
		Alias:         req.Body.Alias,
		LicenseToKill: req.Body.LicenseToKill,
		HiredAt:       req.Body.HiredAt,
	}
	AgentBroadcast.Publish(ctx, agent)
	return builtin.HttpResponse[PostAgentResponse, PostAgentErrorResponse]{
		Status:  201,
		Headers: map[string][]string{"Post": {"Header from FTL"}},
		Body:    ftl.Some(PostAgentResponse{ID: agent.ID}),
	}, nil
}

// Exported verb

type GetNonceRequest struct{}
type GetNonceResponse struct{ Nonce string }

//ftl:verb export
func GetNonce(ctx context.Context, req GetNonceRequest) (GetNonceResponse, error) {
	return GetNonceResponse{Nonce: nonce.Get(ctx)}, nil
}

// Helpers

func generateRandomID() int {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		panic(err)
	}
	return int(n.Int64())
}
