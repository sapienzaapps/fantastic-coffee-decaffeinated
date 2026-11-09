package api

import (
	"net/http"
)

// Handler returns an instance of chi.Mux that handles the APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.Get("/", rt.getHelloWorld)

	// Example of a route with its own middleware, attached with With(...)
	rt.router.With(rt.exampleMiddleware).Get("/context", rt.getContextReply)

	// Example data routes (read and write through the database)
	rt.router.Get("/name", rt.getName)
	rt.router.Put("/name", rt.setName)

	// Example file routes (upload and download through the object storage, with presigned URLs)
	rt.router.Get("/example-photo", rt.getExamplePhoto)
	rt.router.Put("/example-photo", rt.setExamplePhoto)

	// Special routes
	rt.router.Get("/liveness", rt.liveness)

	return rt.router
}
