package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/objectstorage"
)

// examplePhotoURL is the JSON reply of the example photo endpoints.
type examplePhotoURL struct {
	// URL is the presigned URL to use to upload (PUT) or download (GET) the photo.
	URL string `json:"url"`
	// ContentType is the value of the Content-Type header that the upload request must have. It is empty for
	// downloads.
	ContentType string `json:"contentType,omitempty"`
}

// getExamplePhoto is an example of an HTTP endpoint that lets the client download a file from the object storage. It
// replies with a presigned URL: the client then downloads the photo directly from the object storage (for example, by
// using the URL as the src of an <img> element).
func (rt *_router) getExamplePhoto(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	if rt.storage == nil {
		// The object storage is optional: it is nil if it is not configured.
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Check that the photo exists before signing a URL for it.
	if _, err := rt.storage.Stat(r.Context(), examplePhotoKey); errors.Is(err, objectstorage.ErrObjectNotFound) {
		// The photo has not been uploaded yet.
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.Logger.WithError(err).Error("cannot check the example photo")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	url, err := rt.storage.PresignedGetURL(r.Context(), examplePhotoKey, examplePhotoURLExpiry)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot sign the example photo download URL")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(examplePhotoURL{URL: url}); err != nil {
		ctx.Logger.WithError(err).Error("cannot encode the response")
	}
}
