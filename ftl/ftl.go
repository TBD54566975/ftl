package ftl

import (
	"context"
	"sync"

	"github.com/alecthomas/errors"
	"google.golang.org/grpc"

	ftlv1 "github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/common/socket"
)

type FTL struct {
	srv    *ftlServer
	socket socket.Socket
}

func New(socket socket.Socket) *FTL {
	return &FTL{socket: socket}
}

// Serve starts serving the FTL service.
func (b *FTL) Serve() error {
	l, err := b.socket.Listen()
	if err != nil {
		return errors.WithStack(err)
	}
	srv := grpc.NewServer()
	ftlv1.RegisterFTLServiceServer(srv, b.srv)
	return errors.WithStack(srv.Serve(l))
}

type driveContext struct {
	client ftlv1.DriveServiceClient
}

var _ ftlv1.FTLServiceServer = (*ftlServer)(nil)

type ftlServer struct {
	lock   sync.Mutex
	drives []driveContext
}

func (f *ftlServer) RegisterDrive(ctx context.Context, request *ftlv1.RegisterDriveRequest) (*ftlv1.RegisterDriveResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (f *ftlServer) Ping(ctx context.Context, request *ftlv1.PingRequest) (*ftlv1.PingResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (f *ftlServer) Call(ctx context.Context, request *ftlv1.CallRequest) (*ftlv1.CallResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (f *ftlServer) List(ctx context.Context, request *ftlv1.ListRequest) (*ftlv1.ListResponse, error) {
	// TODO implement me
	panic("implement me")
}
