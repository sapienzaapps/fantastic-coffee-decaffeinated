package main

import (
	"net/http"

	"github.com/go-chi/cors"
)

// applyCORSHandler applies a CORS policy to the router. CORS stands for Cross-Origin Resource Sharing: it's a security
// feature present in web browsers that blocks JavaScript requests going across different domains if not specified in a
// policy. This function sends the policy of this API server.
func applyCORSHandler(h http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedHeaders: []string{
			"x-example-header",
			"Content-Type",
		},
		// Only the methods actually used by this sample API. Add the methods your own endpoints use (for example
		// "POST" or "DELETE"); otherwise the browser will block them. OPTIONS (the CORS preflight request) is always
		// handled automatically and must not be listed here.
		AllowedMethods: []string{"GET", "PUT"},
		// Do not modify the CORS origin and max age, they are used in the evaluation.
		AllowedOrigins: []string{"*"},
		MaxAge:         1,
	})(h)
}
