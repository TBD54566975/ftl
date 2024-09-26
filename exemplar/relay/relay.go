package relay

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"ftl/origin"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var logFile = ftl.Config[string]("log_file")
var db = ftl.PostgresDatabase("exemplardb")

// PubSub

var _ = ftl.Subscription(origin.AgentBroadcast, "agentConsumer")

//ftl:subscribe agentConsumer
func ConsumeAgentBroadcast(ctx context.Context, agent origin.Agent) error {
	ftl.LoggerFromContext(ctx).Infof("Received agent %v", agent.Id)
	mission.Send(ctx, strconv.Itoa(agent.Id), agent)
	return nil
}

// FSM

// //ftl:typealias
// type Agent origin.Agent

var mission = ftl.FSM(
	"mission",
	ftl.Start(Briefed),
	ftl.Transition(Briefed, Deployed),
	ftl.Transition(Deployed, Succeeded),
	ftl.Transition(Deployed, Terminated),
)

type deployment struct {
	Agent  origin.Agent
	Target string
}

type missionSuccess struct {
	AgentID   int
	SuccessAt time.Time
}

type agentTerminated struct {
	AgentID      int
	TerminatedAt time.Time
}

//ftl:verb
func Briefed(ctx context.Context, agent origin.Agent) error {
	briefedAt := time.Now()
	ftl.LoggerFromContext(ctx).Infof("Briefed agent %v at %s", agent.Id, briefedAt)
	agent.BriefedAt = ftl.Some(briefedAt)
	d := deployment{
		Agent:  agent,
		Target: "villain",
	}
	return ftl.FSMNext(ctx, d)
}

//ftl:verb
func Deployed(ctx context.Context, d deployment) error {
	// TODO: Insert deployed agent into database, update via Succeeded/Terminated
	// _, err := ftl.Call(ctx, InsertDeployedAgent, InsertDeployedAgentRequest{
	// 	AgentID:       d.Agent.Id,
	// 	Alias:         d.Agent.Alias,
	// 	LicenseToKill: d.Agent.LicenseToKill,
	// 	BriefedAt:     d.Agent.BriefedAt.MustGet(),
	// 	DeployedAt:    time.Now(),
	// })
	// if err != nil {
	// 	return fmt.Errorf("failed to call agent insertion verb: %w", err)
	// }

	ftl.LoggerFromContext(ctx).Infof("Deployed agent %v to %s", d.Agent.Id, d.Target)
	appendLog(ctx, "deployed %d", d.Agent.Id)
	return nil
}

//ftl:verb
func Succeeded(ctx context.Context, s missionSuccess) error {
	fmt.Printf("Agent %d succeeded at %s\n", s.AgentID, s.SuccessAt)
	appendLog(ctx, "succeeded %d", s.AgentID)
	return nil
}

//ftl:verb
func Terminated(ctx context.Context, t agentTerminated) error {
	fmt.Printf("Agent %d terminated at %s\n", t.AgentID, t.TerminatedAt)
	appendLog(ctx, "terminated %d", t.AgentID)
	return nil
}

// Exported verbs

type MissionResultRequest struct {
	AgentID    int
	Successful bool
}

type MissionResultResponse struct{}

//ftl:verb export
func MissionResult(ctx context.Context, req MissionResultRequest) (MissionResultResponse, error) {
	fmt.Printf("Mission result for agent %v: %t\n", req.AgentID, req.Successful)
	agentID := req.AgentID
	var event any
	if req.Successful {
		event = missionSuccess{
			AgentID:   int(agentID),
			SuccessAt: time.Now(),
		}
	} else {
		event = agentTerminated{
			AgentID:      int(agentID),
			TerminatedAt: time.Now(),
		}
	}
	fmt.Printf("Sending event %v\n", event)
	err := mission.Send(ctx, strconv.Itoa(int(agentID)), event)
	if err != nil {
		return MissionResultResponse{}, err
	}
	return MissionResultResponse{}, nil
}

type GetLogFileRequest struct{}
type GetLogFileResponse struct {
	Path string
}

//ftl:verb export
func GetLogFile(ctx context.Context, req GetLogFileRequest) (GetLogFileResponse, error) {
	return GetLogFileResponse{Path: logFile.Get(ctx)}, nil
}

// DB

type InsertDeployedAgentRequest struct {
	AgentID       int
	Alias         string
	LicenseToKill bool
	BriefedAt     time.Time
	DeployedAt    time.Time
}

type InsertDeployedAgentResponse struct{}

//ftl:verb
func InsertDeployedAgent(ctx context.Context, req InsertDeployedAgentRequest) (InsertDeployedAgentResponse, error) {
	err := setupDatabase(ctx)
	if err != nil {
		return InsertDeployedAgentResponse{}, fmt.Errorf("failed to setup database: %w", err)
	}
	_, err = db.Get(ctx).Exec(
		"INSERT INTO deployed_agents (agent_id, alias, license_to_kill, briefed_at, deployed_at) VALUES ($1, $2, $3, $4);",
		req.AgentID, req.Alias, req.LicenseToKill, req.BriefedAt, req.DeployedAt,
	)
	if err != nil {
		return InsertDeployedAgentResponse{}, fmt.Errorf("failed to insert deployed agent: %w", err)
	}
	return InsertDeployedAgentResponse{}, nil
}

type UpdateDeployedAgentRequest struct {
	AgentID      int
	SuccessAt    ftl.Option[time.Time]
	TerminatedAt ftl.Option[time.Time]
}

type UpdateDeployedAgentResponse struct{}

//ftl:verb
func UpdateDeployedAgent(ctx context.Context, req UpdateDeployedAgentRequest) (UpdateDeployedAgentResponse, error) {
	err := setupDatabase(ctx)
	if err != nil {
		return UpdateDeployedAgentResponse{}, fmt.Errorf("failed to setup database: %w", err)
	}
	if successAt, ok := req.SuccessAt.Get(); ok {
		_, err = db.Get(ctx).Exec("UPDATE deployed_agents SET success_at = $1 WHERE agent_id = $2;", successAt, req.AgentID)
		if err != nil {
			return UpdateDeployedAgentResponse{}, fmt.Errorf("failed to update deployed agent: %w", err)
		}
	}
	if termAt, ok := req.TerminatedAt.Get(); ok {
		_, err = db.Get(ctx).Exec("UPDATE deployed_agents SET terminated_at = $1 WHERE agent_id = $2;", termAt, req.AgentID)
		if err != nil {
			return UpdateDeployedAgentResponse{}, fmt.Errorf("failed to update deployed agent: %w", err)
		}
	}
	return UpdateDeployedAgentResponse{}, nil
}

// Helpers

func appendLog(ctx context.Context, msg string, args ...interface{}) {
	path := logFile.Get(ctx)
	if path == "" {
		panic("log_file config not set")
	}
	w, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, msg+"\n", args...)
	err = w.Close()
	if err != nil {
		panic(err)
	}
}

func setupDatabase(ctx context.Context) error {
	_, err := db.Get(ctx).Exec(`CREATE TABLE IF NOT EXISTS deployed_agents (
		agent_id INT PRIMARY KEY,
		alias TEXT,
		license_to_kill BOOLEAN,
		deployed_at TIMESTAMPTZ NOT NULL,
		success_at TIMESTAMPTZ,
		terminated_at TIMESTAMPTZ
	);`)
	return err
}
