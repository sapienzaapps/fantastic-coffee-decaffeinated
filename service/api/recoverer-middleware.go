package api

import (
	"net/http"
	"runtime/debug"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/sirupsen/logrus"
)

// recovererMiddleware recovers from panics and replies with HTTP 500.
//
// It is adapted from chi's middleware.Recoverer: the original prints the stack
// trace directly to stderr, while this version logs it through the
// request-specific logger, so that panics share the same format and destination
// as the rest of the application logs.
func (rt *_router) recovererMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// We don't recover http.ErrAbortHandler: the response must be
				// aborted, and it is not an error worth logging.
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				logger := reqcontext.FromContext(r.Context()).Logger
				if logger == nil {
					logger = rt.baseLogger
				}
				logger.WithFields(logrus.Fields{
					"panic": rvr,
					"stack": string(debug.Stack()),
				}).Error("panic recovered")

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
