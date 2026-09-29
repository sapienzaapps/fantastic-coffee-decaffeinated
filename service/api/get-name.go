package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/database"
)

// nameInfo is the JSON representation of the example name, shared by the name
// endpoints.
type nameInfo struct {
	Name string `json:"name"`
}

// getName is an example of an HTTP endpoint that reads data from the database
// and returns it as JSON. It retrieves the request-specific logger from the
// request context (see reqcontext.FromContext).
func (rt *_router) getName(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	name, err := rt.db.GetName()
	if errors.Is(err, database.ErrNameNotFound) {
		// The name has not been set yet.
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot get the name")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(nameInfo{Name: name}); err != nil {
		ctx.Logger.WithError(err).Error("cannot encode the response")
	}
}
