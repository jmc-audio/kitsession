package session

import (
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/jmc-audio/kitsession/log"
	"golang.org/x/net/context"
)

type Session interface {
	SessionID() string
}

func WithSession() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, i interface{}) (interface{}, error) {
			// Note that production systems *must* check these type assertions
			// Use the Go-ish idiom, e.g., 
			// if val, ok := ctx.Value("key").(string); ok { // safe in this scope ...}
			sessions := *(ctx.Value("sessions").(*map[string]context.Context))
			mtx := ctx.Value("session.mtx").(*sync.Mutex)
			ttl := ctx.Value("session.ttl").(time.Duration)
			refresh := ctx.Value("session.refresh").(bool)

			var session_ctx context.Context
			var ok bool
			var session Session

			if session, ok = i.(Session); ok {
				id := session.SessionID()
				log.Logger(ctx).Debug().
					Log("message", "have session id", "session_id", id)

				mtx.Lock()
				defer mtx.Unlock()

				if session_ctx, ok = sessions[id]; ok {
					log.Logger(ctx).Debug().
						Log("message", "have session context", "session_id", id)

					if expires, ok := session_ctx.Value("expires").(time.Time); ok {
						log.Logger(ctx).Debug().Log("message", "have session expiry", "session_id", id, "expires", expires.UTC().Format(time.RFC3339))
						if expires.Before(time.Now()) {
							log.Logger(ctx).Debug().Log("message", "session expired", "session_id", id, "expires", expires.UTC().Format(time.RFC3339))
							delete(sessions, id)
						} else {
							if refresh {
								log.Logger(ctx).Debug().Log("message", "touching session", "session_id", id)
								session_ctx = context.WithValue(session_ctx, "expires", time.Now().Add(ttl))
								sessions[id] = session_ctx
								return next(session_ctx, i)
							}
						}
					}
				}

				log.Logger(ctx).Debug().Log("message", "init session", "session_id", id)
				session_ctx = context.WithValue(ctx, "session_id", id)
				session_ctx = context.WithValue(session_ctx, "expires", time.Now().Add(ttl))
				sessions[id] = session_ctx

				return next(session_ctx, i)
			}

			return next(ctx, i)
		}
	}
}
