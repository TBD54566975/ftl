package failing

import (
	"context"

	"ftl/failing/child"
	ps "ftl/pubsub"

	"github.com/block/ftl/go-runtime/ftl"
)

type DifferentDeclDupl = ftl.Config[string]

// duplicated in child.go
type FtlConfigEndpoint = ftl.Config[string]

// duplicated in child.go
type FtlSecretEndpoint = ftl.Secret[string]

type DuplDbConfig struct {
	ftl.DefaultPostgresDatabaseConfig
}

func (DuplDbConfig) Name() string { return "testdb" }

type Request struct {
	BadParam error
}
type Response struct {
	AnotherBadParam uint64
}

//ftl:export
func WrongDirective(ctx context.Context, req Request) (Response, error) {
	return Response{}, nil
}

//ftl:verb
func TooManyParams(ctx context.Context, req Request, req2 Request) (Response, error) {
	return Response{}, nil
}

//ftl:verb
func WrongParamOrder(first Request, second string) (Response, error) {
	return Response{}, nil
}

//ftl:verb
func UnitSecondParam(ctx context.Context, unit ftl.Unit) (Response, error) {
	return Response{}, nil
}

//ftl:verb
func NoParams() (Response, error) {
	return Response{}, nil
}

//ftl:verb
func TooManyReturn(ctx context.Context, req Request) (Response, Response, error) {
	return "", Response{}, nil
}

//ftl:verb
func NoReturn(ctx context.Context, req Request) {
}

//ftl:verb
func NoError(ctx context.Context, req Request) Response {
	return Response{}
}

//ftl:verb
func WrongResponse(ctx context.Context, req Request) (string, ftl.Unit) {
	return "", ftl.Unit{}
}

// Duplicate
//
//ftl:verb
func WrongResponse(ctx context.Context, req Request) (string, ftl.Unit) {
	return "", ftl.Unit{}
}

//ftl:verb
type BadStruct struct {
	unexported string
}

//ftl:enum
type TypeEnum interface{ typeEnum() }

//ftl:enum
type BadValueEnum int

func (BadValueEnum) typeEnum() {}

const (
	A1 BadValueEnum = iota
)

//ftl:enum
type BadValueEnumOrderDoesntMatter int

const (
	A2 BadValueEnumOrderDoesntMatter = iota
)

func (BadValueEnumOrderDoesntMatter) typeEnum() {}

//ftl:enum export
type ExportedTypeEnum interface{ exportedTypeEnum() }

//ftl:data
type PrivateData struct{}

func (PrivateData) exportedTypeEnum() {}

// type sa9d8s!aksd = ftl.Config[string]

// There can be only one
//
//ftl:typealias
//ftl:enum
type TypeAliasAndEnum int

type NonFTLAlias int

//ftl:data
type DataUsingNonFTLAlias struct {
	Value NonFTLAlias
}

//ftl:enum
type TypeEnum1 interface{ typeEnum1() }

//ftl:enum
type TypeEnum2 interface{ typeEnum2() }

//ftl:typealias
type ConformsToTwoTypeEnums string

func (ConformsToTwoTypeEnums) typeEnum1() {}
func (ConformsToTwoTypeEnums) typeEnum2() {}

//ftl:enum
type TypeEnum3 interface{ ExportedMethod() }

//ftl:enum
type NoMethodsTypeEnum interface{}

//ftl:cron * * * * *
func GoodCron(ctx context.Context) error {
	return nil
}

//ftl:cron 0 0 0 0 0
func BadCron(ctx context.Context) error {
	return nil
}

//ftl:cron 4d
func DaysNotPossible(ctx context.Context) error {
	return nil
}

//ftl:verb
func BadPublish(ctx context.Context, topic ps.PublicBroadcast) error {
	topic.Publish(ctx, ps.PayinEvent{Name: "Test"})
	return ps.Broadcast(ctx)
}

//ftl:data
type UnexportedFieldStruct struct {
	unexported string
}

//ftl:data
type BadChildField struct {
	Child child.BadChildStruct
}

//ftl:data
type Redeclared struct {
}

//ftl:enum
type EnumVariantConflictParent int

const (
	SameVariant EnumVariantConflictParent = iota
)

//ftl:verb
func Empty(ctx context.Context) error {
	return nil
}

//ftl:verb
func CallsEmptyDirectly(ctx context.Context) error {
	return Empty(ctx)
}
