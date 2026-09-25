package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

// setNameRequest is the JSON body accepted by the setName endpoint.
type setNameRequest struct {
	Name string `json:"name"`
}

// setName is an example of an HTTP endpoint that writes data to the database.
// It reads a JSON body, validates it, and stores the name.
func (rt *_router) setName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req setNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Warn("cannot decode the request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := rt.db.SetName(req.Name); err != nil {
		ctx.Logger.WithError(err).Error("cannot set the name")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nameInfo{Name: req.Name}); err != nil {
		ctx.Logger.WithError(err).Error("cannot encode the response")
	}
}
