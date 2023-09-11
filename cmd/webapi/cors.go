package main

import (
	"github.com/gorilla/handlers"
	"net/http"
)

// applyCORSHandler applies a CORS policy to the router. CORS stands for Cross-Origin Resource Sharing: it's a security
// feature present in web browsers that blocks JavaScript requests going across different domains if not specified in a
// policy. This function sends the policy of this API server.
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"x-example-header",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// Do not modify the CORS origin and max age, they are used in the evaluation.
		handlers.AllowedOrigins([]string{"*"}),
		handlers.MaxAge(1),
	)(h)
}
