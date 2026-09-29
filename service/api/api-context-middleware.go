package api

import (
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// contextMiddleware adds a reqcontext.RequestContext instance related to the
// request. It is installed on the whole router (see New in api.go), so every
// handler can retrieve it with reqcontext.FromContext(r.Context()).
func (rt *_router) contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(r.Context())

		// Create a request-specific logger
		ctx := reqcontext.RequestContext{
			ID: reqID,
			Logger: rt.baseLogger.WithFields(logrus.Fields{
				"reqid":     reqID,
				"remote-ip": r.RemoteAddr,
			}),
		}

		// Store the context in the request and call the next handler in the chain
		next.ServeHTTP(w, r.WithContext(reqcontext.WithContext(r.Context(), ctx)))
	})
}
