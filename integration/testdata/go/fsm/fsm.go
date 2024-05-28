package fsm

import (
	"context"
	"fmt"
	"os"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type One struct {
	Instance string
}

type Two struct {
	Instance string
}

var fsm = ftl.FSM("fsm",
	ftl.Start(Start),
	ftl.Transition(Start, Middle),
	ftl.Transition(Middle, End),
)

//ftl:verb
func Start(ctx context.Context, in One) error {
	appendLog("start %s", in.Instance)
	return nil
}

//ftl:verb
func Middle(ctx context.Context, in One) error {
	appendLog("middle %s", in.Instance)
	return nil
}

//ftl:verb
func End(ctx context.Context, in One) error {
	appendLog("end %s", in.Instance)
	return nil
}

//ftl:verb
func SendOne(ctx context.Context, in One) error {
	return fsm.Send(ctx, in.Instance, in)
}

//ftl:verb
func SendTwo(ctx context.Context, in Two) error {
	return fsm.Send(ctx, in.Instance, in)
}

func appendLog(msg string, args ...interface{}) {
	dest, ok := os.LookupEnv("FSM_LOG_FILE")
	if !ok {
		panic("FSM_LOG_FILE not set")
	}
	w, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, msg+"\n", args...)
	err = w.Close()
	if err != nil {
		panic(err)
	}
}
