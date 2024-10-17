package languagepb

import (
	"fmt"

	structpb "google.golang.org/protobuf/types/known/structpb"

	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/slices"
)

// ErrorsFromProto converts a protobuf ErrorList to a []builderrors.Error.
func ErrorsFromProto(e *ErrorList) []builderrors.Error {
	if e == nil {
		return []builderrors.Error{}
	}
	return slices.Map(e.Errors, errorFromProto)
}

func ErrorsToProto(errs []builderrors.Error) *ErrorList {
	return &ErrorList{Errors: slices.Map(errs, errorToProto)}
}

func levelFromProto(level Error_ErrorLevel) builderrors.ErrorLevel {
	switch level {
	case Error_INFO:
		return builderrors.INFO
	case Error_WARN:
		return builderrors.WARN
	case Error_ERROR:
		return builderrors.ERROR
	}
	panic(fmt.Sprintf("unhandled ErrorLevel %v", level))
}

func levelToProto(level builderrors.ErrorLevel) Error_ErrorLevel {
	switch level {
	case builderrors.INFO:
		return Error_INFO
	case builderrors.WARN:
		return Error_WARN
	case builderrors.ERROR:
		return Error_ERROR
	}
	panic(fmt.Sprintf("unhandled ErrorLevel %v", level))
}

func errorFromProto(e *Error) builderrors.Error {
	return builderrors.Error{
		Pos:   PosFromProto(e.Pos),
		Msg:   e.Msg,
		Level: levelFromProto(e.Level),
	}
}

func errorToProto(e builderrors.Error) *Error {
	return &Error{
		Msg: e.Msg,
		Pos: &Position{
			Filename:    e.Pos.Filename,
			StartColumn: int64(e.Pos.StartColumn),
			EndColumn:   int64(e.Pos.EndColumn),
			Line:        int64(e.Pos.Line),
		},
		Level: levelToProto(e.Level),
	}
}

func PosFromProto(pos *Position) builderrors.Position {
	if pos == nil {
		return builderrors.Position{}
	}
	return builderrors.Position{
		Line:        int(pos.Line),
		StartColumn: int(pos.StartColumn),
		EndColumn:   int(pos.EndColumn),
		Filename:    pos.Filename,
	}
}

func LogLevelFromProto(level LogMessage_LogLevel) log.Level {
	switch level {
	case LogMessage_INFO:
		return log.Info
	case LogMessage_DEBUG:
		return log.Debug
	case LogMessage_WARN:
		return log.Warn
	case LogMessage_ERROR:
		return log.Error
	default:
		panic(fmt.Sprintf("unhandled log level %v", level))
	}
}

func LogLevelToProto(level log.Level) LogMessage_LogLevel {
	switch level {
	case log.Info:
		return LogMessage_INFO
	case log.Debug:
		return LogMessage_DEBUG
	case log.Warn:
		return LogMessage_WARN
	case log.Error:
		return LogMessage_ERROR
	default:
		panic(fmt.Sprintf("unhandled log level %v", level))
	}
}

// ModuleConfigToProto converts a moduleconfig.AbsModuleConfig to a protobuf ModuleConfig.
//
// Absolute configs are used because relative paths may change resolve differently between parties.
func ModuleConfigToProto(config moduleconfig.AbsModuleConfig) (*ModuleConfig, error) {
	proto := &ModuleConfig{
		Name:      config.Module,
		Dir:       config.Dir,
		DeployDir: config.DeployDir,
		Watch:     config.Watch,
		Language:  config.Language,
	}
	if config.Build != "" {
		proto.Build = &config.Build
	}
	if config.GeneratedSchemaDir != "" {
		proto.GeneratedSchemaDir = &config.GeneratedSchemaDir
	}

	langConfigProto, err := structpb.NewStruct(config.LanguageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal language config: %w", err)
	}
	proto.LanguageConfig = langConfigProto
	return proto, nil

}

// ModuleConfigFromProto converts a protobuf ModuleConfig to a moduleconfig.AbsModuleConfig.
func ModuleConfigFromProto(proto *ModuleConfig) moduleconfig.AbsModuleConfig {
	config := moduleconfig.AbsModuleConfig{
		Module:             proto.Name,
		Dir:                proto.Dir,
		DeployDir:          proto.DeployDir,
		Watch:              proto.Watch,
		Language:           proto.Language,
		Build:              proto.GetBuild(),
		GeneratedSchemaDir: proto.GetGeneratedSchemaDir(),
	}
	if proto.LanguageConfig != nil {
		config.LanguageConfig = proto.LanguageConfig.AsMap()
	}
	return config
}

func ProjectConfigToProto(projConfig projectconfig.Config) *ProjectConfig {
	return &ProjectConfig{
		Dir:    projConfig.Path,
		Name:   projConfig.Name,
		NoGit:  projConfig.NoGit,
		Hermit: projConfig.Hermit,
	}
}

func ProjectConfigFromProto(proto *ProjectConfig) projectconfig.Config {
	return projectconfig.Config{
		Path:   proto.Dir,
		Name:   proto.Name,
		NoGit:  proto.NoGit,
		Hermit: proto.Hermit,
	}
}
