package status

import "context"

var _ StatusManager = &noopStatusManager{}
var _ StatusLine = &noopStatusLine{}

type noopStatusManager struct{}

func (r *noopStatusManager) Close() {
}

func (r *noopStatusManager) SetModuleState(module string, state BuildState) {

}

func (r *noopStatusManager) IntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, statusKeyInstance, r)
}

type noopStatusLine struct{}

func (n noopStatusLine) SetMessage(message string) {
}

func (n noopStatusLine) Close() {
}

func (r *noopStatusManager) NewStatus(message string) StatusLine {
	return &noopStatusLine{}
}
