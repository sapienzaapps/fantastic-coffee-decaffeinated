package api

import (
	"encoding/json"
	"net/http"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/api/reqcontext"
)

// The example photo endpoints show how to upload and download a file with presigned URLs: the web API only checks the
// request and signs a temporary URL, then the client (e.g., the browser) uploads or downloads the file directly to/from
// the object storage. The file content never passes through the web API.
const (
	// examplePhotoKey is the object storage key of the example photo. It is fixed because this is just an example: a
	// real application usually generates a new key for each file (e.g., a random UUID) and stores it in the database.
	examplePhotoKey = "example-photo"

	// examplePhotoContentType is the only file type accepted by the example.
	examplePhotoContentType = "image/jpeg"

	// examplePhotoMaxSize is the maximum size of the example photo (10 MB).
	examplePhotoMaxSize = 10 * 1024 * 1024

	// examplePhotoURLExpiry is how long the presigned URLs remain valid.
	examplePhotoURLExpiry = 5 * time.Minute
)

// setExamplePhotoRequest is the JSON body accepted by the setExamplePhoto endpoint.
type setExamplePhotoRequest struct {
	// Size is the size of the photo that the client is going to upload, in bytes.
	Size int64 `json:"size"`
}

// setExamplePhoto is an example of an HTTP endpoint that lets the client upload a file to the object storage. The
// client sends the size of the photo; the endpoint checks it and replies with a presigned URL. The client then uploads
// the photo directly to the object storage, with an HTTP PUT request to that URL, with the photo as the body and the
// returned content type in the Content-Type header. For example, in JavaScript:
//
//	await fetch(reply.url, {method: "PUT", body: file, headers: {"Content-Type": reply.contentType}})
//
// The size and the content type are part of the URL signature: the object storage rejects the upload if the file has
// a different size, or if it is sent with a different content type. Note that the object storage does not check that
// the file is actually a JPEG image: the client declares it.
func (rt *_router) setExamplePhoto(w http.ResponseWriter, r *http.Request) {
	ctx := reqcontext.FromContext(r.Context())

	if rt.storage == nil {
		// The object storage is optional: it is nil if it is not configured.
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	var req setExamplePhotoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ctx.Logger.WithError(err).Warn("cannot decode the request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Size <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Size > examplePhotoMaxSize {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	url, err := rt.storage.PresignedPutURL(r.Context(), examplePhotoKey, req.Size, examplePhotoContentType,
		examplePhotoURLExpiry)
	if err != nil {
		ctx.Logger.WithError(err).Error("cannot sign the example photo upload URL")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(examplePhotoURL{URL: url, ContentType: examplePhotoContentType}); err != nil {
		ctx.Logger.WithError(err).Error("cannot encode the response")
	}
}
