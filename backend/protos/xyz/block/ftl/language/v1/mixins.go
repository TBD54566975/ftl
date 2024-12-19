package languagepb

import (
	"fmt"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/projectconfig"
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
	case Error_ERROR_LEVEL_INFO:
		return builderrors.INFO
	case Error_ERROR_LEVEL_WARN:
		return builderrors.WARN
	case Error_ERROR_LEVEL_ERROR:
		return builderrors.ERROR
	}
	panic(fmt.Sprintf("unhandled ErrorLevel %v", level))
}

func levelToProto(level builderrors.ErrorLevel) Error_ErrorLevel {
	switch level {
	case builderrors.INFO:
		return Error_ERROR_LEVEL_INFO
	case builderrors.WARN:
		return Error_ERROR_LEVEL_WARN
	case builderrors.ERROR:
		return Error_ERROR_LEVEL_ERROR
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
	var pos *Position
	if bpos, ok := e.Pos.Get(); ok {
		pos = &Position{
			Filename:    bpos.Filename,
			StartColumn: int64(bpos.StartColumn),
			EndColumn:   int64(bpos.EndColumn),
			Line:        int64(bpos.Line),
		}
	}
	return &Error{
		Msg:   e.Msg,
		Pos:   pos,
		Level: levelToProto(e.Level),
	}
}

func PosFromProto(pos *Position) optional.Option[builderrors.Position] {
	if pos == nil {
		return optional.None[builderrors.Position]()
	}
	return optional.Some(builderrors.Position{
		Line:        int(pos.Line),
		StartColumn: int(pos.StartColumn),
		EndColumn:   int(pos.EndColumn),
		Filename:    pos.Filename,
	})
}

// ModuleConfigToProto converts a moduleconfig.AbsModuleConfig to a protobuf ModuleConfig.
//
// Absolute configs are used because relative paths may change resolve differently between parties.
func ModuleConfigToProto(config moduleconfig.AbsModuleConfig) (*ModuleConfig, error) {
	proto := &ModuleConfig{
		Name:      config.Module,
		Dir:       config.Dir,
		DeployDir: config.DeployDir,
		BuildLock: config.BuildLock,
		Watch:     config.Watch,
		Language:  config.Language,
	}
	if config.Build != "" {
		proto.Build = &config.Build
	}
	if config.DevModeBuild != "" {
		proto.DevModeBuild = &config.DevModeBuild
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
		Module:                proto.Name,
		Dir:                   proto.Dir,
		DeployDir:             proto.DeployDir,
		Watch:                 proto.Watch,
		Language:              proto.Language,
		Build:                 proto.GetBuild(),
		DevModeBuild:          proto.GetDevModeBuild(),
		BuildLock:             proto.BuildLock,
		SQLMigrationDirectory: proto.GetSqlMigrationDir(),
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
