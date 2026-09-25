package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// Example data routes (read and write through the database)
	rt.router.GET("/name", rt.wrap(rt.getName))
	rt.router.PUT("/name", rt.wrap(rt.setName))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
