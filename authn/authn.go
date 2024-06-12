package authn

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/user"
	"strings"
	"sync"
	"time"

	"github.com/zalando/go-keyring"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

// GetAuthenticationHeaders returns authentication headers for the given endpoint.
//
// "authenticators" are authenticator executables to use for each endpoint. The key is the URL of the endpoint, the
// value is the name/path of the authenticator executable. The authenticator executable will be called with the URL as
// the first argument, and output a list of headers to stdout to use for authentication.
//
// If the endpoint is already authenticated, the existing credentials will be returned. Additionally, credentials will
// be cached across runs in the keyring.
func GetAuthenticationHeaders(ctx context.Context, endpoint *url.URL, authenticators map[string]string) (http.Header, error) {
	logger := log.FromContext(ctx).Scope(endpoint.Hostname())

	// Next, try the authenticator.
	logger.Tracef("Trying authenticator")
	authenticator, ok := authenticators[endpoint.Hostname()]
	if !ok {
		logger.Tracef("No authenticator found for %s in %s", endpoint, authenticators)
		return nil, nil
	}

	endpoint = &url.URL{
		Scheme: endpoint.Scheme,
		Host:   endpoint.Host,
		User:   endpoint.User,
	}

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	// First, check if we have credentials in the keyring and that they work.
	keyringKey := "ftl+" + endpoint.String()
	logger.Tracef("Trying keyring key %s", keyringKey)
	creds, err := keyring.Get(keyringKey, usr.Name)
	if errors.Is(err, keyring.ErrNotFound) {
		logger.Tracef("No credentials found in keyring")
	} else if err != nil {
		if !strings.Contains(err.Error(), `exec: "dbus-launch": executable file not found in $PATH`) {
			logger.Tracef("Failed to get credentials from keyring: %s", err)
		}
	} else {
		logger.Tracef("Credentials found in keyring: %s", creds)
		if headers, err := checkAuth(ctx, logger, endpoint, creds); err != nil {
			return nil, err
		} else if headers != nil {
			return headers, nil
		}
	}

	cmd := exec.Command(ctx, log.Error, ".", authenticator, endpoint.String())
	out := &strings.Builder{}
	cmd.Stdout = out
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("authenticator %s failed: %w", authenticator, err)
	}

	creds = out.String()
	headers, err := checkAuth(ctx, logger, endpoint, creds)
	if err != nil {
		return nil, err
	} else if headers == nil {
		return nil, nil
	}

	logger.Tracef("Authenticator %s succeeded", authenticator)
	w := &strings.Builder{}
	for name, values := range headers {
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\r\n", name, value)
		}
	}
	err = keyring.Set(keyringKey, usr.Name, w.String())
	if err != nil {
		logger.Tracef("Failed to save credentials to keyring: %s", err)
	}
	return headers, nil
}

// Check credentials and return authenticating headers if we're able to successfully authenticate.
func checkAuth(ctx context.Context, logger *log.Logger, endpoint *url.URL, creds string) (http.Header, error) {
	// Parse the headers
	headers := http.Header{}
	buf := bufio.NewScanner(strings.NewReader(creds))
	logger.Tracef("Parsing credentials")
	for buf.Scan() {
		line := buf.Text()
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header %q", line)
		}
		headers[name] = append(headers[name], strings.TrimSpace(value))
	}
	if buf.Err() != nil {
		return nil, buf.Err()
	}

	// Issue a HEAD request with the headers to verify we get a 200 back.
	client := &http.Client{
		Timeout: time.Second * 5,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse //errtrace:skip(the HTTP client doesn't handle wrapping errors)
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	logger.Tracef("Authentication probe: %s %s", req.Method, req.URL)
	for header, values := range headers {
		for _, value := range values {
			req.Header.Add(header, value)
		}
	}
	logger.Tracef("Authenticating with headers %s", headers)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:gosec
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Tracef("Endpoint returned %d for authenticated request", resp.StatusCode)
		logger.Tracef("Response headers: %s", resp.Header)
		logger.Tracef("Response body: %s", body)
		return nil, nil
	}
	logger.Tracef("Successfully authenticated with %s", headers)
	return headers, nil
}

// Transport returns a transport that will authenticate requests to the given endpoints.
func Transport(next http.RoundTripper, authenticators map[string]string) http.RoundTripper {
	if len(authenticators) == 0 {
		return next
	}
	return &authnTransport{
		authenticators: authenticators,
		credentials:    map[string]http.Header{},
		next:           next,
	}
}

type authnTransport struct {
	lock           sync.RWMutex
	authenticators map[string]string
	credentials    map[string]http.Header
	next           http.RoundTripper
}

func (a *authnTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	a.lock.RLock()
	creds, ok := a.credentials[r.URL.Hostname()]
	a.lock.RUnlock()
	if !ok {
		var err error
		creds, err = GetAuthenticationHeaders(r.Context(), r.URL, a.authenticators)
		if err != nil {
			return nil, fmt.Errorf("failed to get authentication headers for %s: %w", r.URL.Hostname(), err)
		}
		a.lock.Lock()
		a.credentials[r.URL.Hostname()] = creds
		a.lock.Unlock()
	}
	for header, values := range creds {
		for _, value := range values {
			r.Header.Add(header, value)
		}
	}
	resp, err := a.next.RoundTrip(r)
	return resp, err
}
