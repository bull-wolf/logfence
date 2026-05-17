package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yourorg/logfence/internal/redactor"
)

// DispatchMiddleware annotates the request header with a resolved output name
// based on the log entry's field value, using a Dispatcher.
type DispatchMiddleware struct {
	dispatcher *redactor.Dispatcher
	next       http.Handler
	// HeaderName is the HTTP header set on the proxied request.
	headerName string
}

// NewDispatchMiddleware creates a DispatchMiddleware with default config.
func NewDispatchMiddleware(next http.Handler) *DispatchMiddleware {
	return &DispatchMiddleware{
		dispatcher: redactor.NewDispatcher(redactor.DefaultDispatcherConfig()),
		next:       next,
		headerName: "X-Logfence-Output",
	}
}

// NewDispatchMiddlewareWithConfig creates a DispatchMiddleware with a custom config.
func NewDispatchMiddlewareWithConfig(cfg redactor.DispatcherConfig, headerName string, next http.Handler) *DispatchMiddleware {
	if headerName == "" {
		headerName = "X-Logfence-Output"
	}
	return &DispatchMiddleware{
		dispatcher: redactor.NewDispatcher(cfg),
		next:       next,
		headerName: headerName,
	}
}

func (m *DispatchMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		m.next.ServeHTTP(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		m.next.ServeHTTP(w, r)
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	output := m.dispatcher.Dispatch(body)
	r.Header.Set(m.headerName, output)

	m.next.ServeHTTP(w, r)
}
