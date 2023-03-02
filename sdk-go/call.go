package sdkgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	"github.com/alecthomas/errors"
)

// Call a Verb through the Agent.
func Call[Req, Resp any](ctx context.Context, verb func(ctx context.Context, req Req) (Resp, error), req Req) (resp Resp, err error) {
	callerPc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(callerPc).Name()
	callee := runtime.FuncForPC(reflect.ValueOf(verb).Pointer()).Name()
	reqData, err := json.Marshal(req)
	if err != nil {
		return resp, errors.Wrap(err, "failed to marshal request")
	}
	hreq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("http://127.0.0.1:8080/%s", callee),
		bytes.NewReader(reqData),
	)
	if err != nil {
		return resp, errors.Wrap(err, "failed to create request")
	}
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("User-Agent", "FTL")
	hreq.Header.Set("X-FTL-Caller", caller)
	hresp, err := http.DefaultClient.Do(hreq)
	if err != nil {
		return resp, errors.Wrap(err, "failed to send request")
	}
	defer hresp.Body.Close() //nolint:gosec
	if hresp.StatusCode != http.StatusOK {
		return resp, errors.Errorf("verb failed: %s", hresp.Status)
	}
	dec := json.NewDecoder(hresp.Body)
	err = dec.Decode(&resp)
	if err != nil {
		return resp, errors.Wrap(err, "failed to decode response")
	}
	return resp, nil
}
