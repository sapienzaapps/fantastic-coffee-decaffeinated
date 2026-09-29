package api

import (
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
)

// getContextReply is an example of HTTP endpoint that returns "Hello World!" as a plain text. It shows how to retrieve
// the request-specific reqcontext.RequestContext (with the request ID and the logger) from the request context.
func (rt *_router) getContextReply(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())
	ctx.Logger.Debug("request context retrieved")

	w.Header().Set("content-type", "text/plain")
	_, _ = w.Write([]byte("Hello World!"))
}
