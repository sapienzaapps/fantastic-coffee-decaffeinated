package api

import "net/http"

// exampleMiddleware is a no-op middleware.
//
// It exists to show how to attach a middleware to a single route (or a group of
// routes) with router.With(...), instead of applying it to the whole router.
// See the "/context" route in api-handler.go.
func (rt *_router) exampleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do something here, then call the next handler in the chain.
		rt.baseLogger.WithField("path", r.URL.Path).Debug("example middleware executed")
		next.ServeHTTP(w, r)
	})
}
