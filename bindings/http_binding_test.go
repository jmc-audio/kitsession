package bindings

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"

	kitlog "github.com/go-kit/kit/log"
)

func TestNoop(t *testing.T) {
}

var (
	statsdAddress           = "127.0.0.1:8125"
	statsdReportingInterval = int64(1)
	rateLimitFillInterval   = int64(60)
	rateLimitCapacity       = int64(1024)
	rateLimitQuantum        = int64(1)
)

type httpBindingsTestCase struct {
	name                    string
	httpRequestMethod       string
	httpRequestURL          string
	httpRequestBody         []byte
	expectedResponseCode    int
	expectedResponseBody    []byte
	expectedResponseHeaders map[string]string
	mockSetup               func()
}

// Implementation of Reader interface that returns an error any time the Read function is called
type FailingReader struct{}

// Read implementation that simulates a read error by always returning an error
func (f *FailingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("Simulated read error")
}

type HTTPBindingsSuite struct {
	suite.Suite
	errc     chan error
	logger   kitlog.Logger
	endpoint Servicer
}

func TestHTTPBindingsSuite(t *testing.T) {
	suite.Run(t, new(HTTPBindingsSuite))
}

func (s *HTTPBindingsSuite) SetupTest() {
	s.resetMocks()
}

func (s *HTTPBindingsSuite) SetupSuite() {
	s.errc = make(chan error)
	s.logger = kitlog.NewLogfmtLogger(os.Stderr)
}

func (s *HTTPBindingsSuite) resetContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "errc", s.errc)
	ctx = context.WithValue(ctx, "logger", s.logger)
	sessions := make(map[string]context.Context)
	mtx := &sync.Mutex{}

	ctx = context.WithValue(ctx, "sessions", &sessions)
	ctx = context.WithValue(ctx, "session.mtx", mtx)
	ctx = context.WithValue(ctx, "session.ttl", 5*time.Second)
	ctx = context.WithValue(ctx, "session.refresh", true)
	return ctx
}

func (s *HTTPBindingsSuite) resetMocks() {
}

func (s *HTTPBindingsSuite) assertMockExpectations() {
}

func (s *HTTPBindingsSuite) TestApplicationEndpoints() {
	tests := []*httpBindingsTestCase{
		{
			name:                 "1",
			httpRequestMethod:    "GET",
			httpRequestURL:       "http://seed/id/1",
			httpRequestBody:      []byte{},
			expectedResponseBody: []byte("{\"Status\":\"OK\"}\n"),
			expectedResponseCode: http.StatusOK,
			mockSetup:            func() {},
		}, {
			name:                 "2",
			httpRequestMethod:    "GET",
			httpRequestURL:       "http://seed/id/2",
			httpRequestBody:      []byte{},
			expectedResponseBody: []byte("{\"Status\":\"OK\"}\n"),
			expectedResponseCode: http.StatusOK,
			mockSetup:            func() {},
		},
	}

	for _, t := range tests {
		// GIVEN

		s.resetMocks()
		t.mockSetup()
		ctx := s.resetContext()

		s.endpoint = NewEndpoint(ctx)

		router := createRouter(ctx, s.endpoint)

		req, err := http.NewRequest(t.httpRequestMethod, t.httpRequestURL, bytes.NewReader(t.httpRequestBody))

		s.Nil(err, fmt.Sprintf("%s: Error building HTTP request", t.name))
		s.NotNil(req, fmt.Sprintf("%s: Unable to build HTTP request", t.name))

		// WHEN
		response := httptest.NewRecorder()

		router.ServeHTTP(response, req)

		// THEN
		s.Equal(t.expectedResponseCode, response.Code, fmt.Sprintf("%s: Response code mismatch", t.name))
		s.Equal(t.expectedResponseBody, response.Body.Bytes(),
			fmt.Sprintf("%s: Response body mismatch\n%s\n", t.name, spew.Sdump(response.Body)))

		for h := range t.expectedResponseHeaders {
			s.Equal(t.expectedResponseHeaders[h], response.Header().Get(h), fmt.Sprintf("%s: Response header mismatch", t.name))
		}
		s.assertMockExpectations()
	}
}
