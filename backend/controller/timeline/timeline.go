package timeline

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/timeline/dal"
	"github.com/TBD54566975/ftl/backend/libdal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

type Service struct {
	dal *dal.DAL
}

func New(ctx context.Context, conn libdal.Connection, encryption *encryption.Service) *Service {
	return &Service{dal: dal.New(conn, encryption)}
}

type Log struct {
	DeploymentKey model.DeploymentKey
	RequestKey    optional.Option[model.RequestKey]
	Msg           *ftlv1.StreamDeploymentLogsRequest
}

func (s *Service) RecordLog(ctx context.Context, log *Log) error {
	err := s.dal.InsertLogEvent(ctx, &dal.LogEvent{
		RequestKey:    log.RequestKey,
		DeploymentKey: log.DeploymentKey,
		Time:          log.Msg.TimeStamp.AsTime(),
		Level:         log.Msg.LogLevel,
		Attributes:    log.Msg.Attributes,
		Message:       log.Msg.Message,
		Error:         optional.Ptr(log.Msg.Error),
	})
	if err != nil {
		return fmt.Errorf("failed to insert log event: %w", err)
	}
	return nil
}

type Call struct {
	DeploymentKey    model.DeploymentKey
	RequestKey       model.RequestKey
	ParentRequestKey optional.Option[model.RequestKey]
	StartTime        time.Time
	DestVerb         *schema.Ref
	Callers          []*schema.Ref
	Request          *ftlv1.CallRequest
	Response         either.Either[*ftlv1.CallResponse, error]
}

func (s *Service) RecordCall(ctx context.Context, call *Call) {
	logger := log.FromContext(ctx)
	var sourceVerb optional.Option[schema.Ref]
	if len(call.Callers) > 0 {
		sourceVerb = optional.Some(*call.Callers[0])
	}

	var errorStr optional.Option[string]
	var stack optional.Option[string]
	var responseBody []byte

	switch response := call.Response.(type) {
	case either.Left[*ftlv1.CallResponse, error]:
		resp := response.Get()
		responseBody = resp.GetBody()
		if callError := resp.GetError(); callError != nil {
			errorStr = optional.Some(callError.Message)
			stack = optional.Ptr(callError.Stack)
		}
	case either.Right[*ftlv1.CallResponse, error]:
		callError := response.Get()
		errorStr = optional.Some(callError.Error())
	}

	err := s.dal.InsertCallEvent(ctx, &dal.CallEvent{
		Time:             call.StartTime,
		DeploymentKey:    call.DeploymentKey,
		RequestKey:       optional.Some(call.RequestKey),
		ParentRequestKey: call.ParentRequestKey,
		Duration:         time.Since(call.StartTime),
		SourceVerb:       sourceVerb,
		DestVerb:         *call.DestVerb,
		Request:          call.Request.GetBody(),
		Response:         responseBody,
		Error:            errorStr,
		Stack:            stack,
	})
	if err != nil {
		logger.Errorf(err, "failed to record call")
	}
}

func (s *Service) InsertLogEvent(ctx context.Context, log *dal.LogEvent) error {
	err := s.dal.InsertLogEvent(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to insert log event: %w", err)
	}
	return nil
}

func (s *Service) QueryTimeline(ctx context.Context, limit int, filters ...dal.TimelineFilter) ([]dal.TimelineEvent, error) {
	events, err := s.dal.QueryTimeline(ctx, limit, filters...)
	if err != nil {
		return nil, fmt.Errorf("failed to query timeline: %w", err)
	}
	return events, nil
}

func (s *Service) DeleteOldEvents(ctx context.Context, eventType dal.EventType, age time.Duration) (int64, error) {
	deleted, err := s.dal.DeleteOldEvents(ctx, eventType, age)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old events: %w", err)
	}
	return deleted, nil
}
