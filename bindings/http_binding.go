package bindings

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/net/context"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/jmc-audio/kitsession/session"

	"github.com/gorilla/handlers"

	"github.com/gorilla/mux"
)

type Servicer interface {
	Run(context.Context, interface{}) (interface{}, error)
}

type Request struct {
	params map[string]string
}

type Response struct {
	Status string
}

func (r *Request) SessionID() string {
	if id, ok := r.params["id"]; ok {
		return id
	}
	return ""
}

func decodeRequest(r *http.Request) (response interface{}, err error) {
	var (
		param, value string
		ok           bool
	)
	urlParams := mux.Vars(r)

	if param, ok = urlParams["param"]; !ok {
		return nil, errors.New("No param in request")
	}
	if value, ok = urlParams["value"]; !ok {
		return nil, errors.New("No param in request")
	}
	return &Request{map[string]string{param: value}}, nil
}

func encodeResponse(w http.ResponseWriter, i interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(i.(*Response))
}

type Endpoint struct{}

func NewEndpoint(ctx context.Context) Servicer {
	return &Endpoint{}
}

func (h *Endpoint) Run(ctx context.Context, i interface{}) (interface{}, error) {
	//	log.Logger(ctx).Debug().Log("ctx", spew.Sdump(ctx))
	return &Response{"OK"}, nil
}

func StartHTTPListener(root context.Context) {
	go func() {
		ctx, cancel := context.WithCancel(root)
		defer cancel()

		errc := ctx.Value("errc").(chan error)

		sessions := make(map[string]context.Context)
		mtx := &sync.Mutex{}

		ctx = context.WithValue(ctx, "sessions", &sessions)
		ctx = context.WithValue(ctx, "session.mtx", mtx)
		ctx = context.WithValue(ctx, "session.ttl", 5*time.Second)
		ctx = context.WithValue(ctx, "session.refresh", true)

		router := createRouter(ctx, NewEndpoint(ctx))
		errc <- http.ListenAndServe(":6502", handlers.CombinedLoggingHandler(os.Stderr, router))
	}()
}

func createRouter(ctx context.Context, endpoint Servicer) *mux.Router {
	router := mux.NewRouter()

	router.Handle("/{param}/{value}",
		kithttp.NewServer(
			ctx,
			session.WithSession()(endpoint.Run),
			decodeRequest,
			encodeResponse,
		)).Methods("GET")
	return router
}
