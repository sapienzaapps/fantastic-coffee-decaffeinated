/*
Package objectstorage defines the contract between the app and an object storage, used to store files (images,
attachments, ...) outside the database. Each file is an "object" identified by a key (for example "avatars/42.png");
the database usually stores only the key.

The rest of the program (for example the API handlers) depends only on the ObjectStorage interface, never on a
concrete storage service. The included implementation, in the s3 sub-package, works with any S3-compatible service
(Garage, MinIO, AWS S3, ...):

	storage, err := s3.New(ctx, s3.Config{...})
	if err != nil {
		return err
	}

Object storage is optional: the web API starts without it when no endpoint is configured.
*/
package objectstorage

import (
	"context"
	"errors"
	"io"
	"time"
)

// ObjectStorage is the high level interface for the object storage.
type ObjectStorage interface {
	// Put stores the content read from r under key, replacing any existing object with the same key. size is the
	// number of bytes that r will return, or -1 if unknown (an unknown size uses more memory while uploading).
	// contentType is the MIME type of the content (for example "image/png").
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error

	// Get returns the content of the object stored under key, and its metadata. The caller must close the returned
	// io.ReadCloser. If the object does not exist, it returns ErrObjectNotFound.
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error)

	// Stat returns the metadata of the object stored under key, without downloading it. If the object does not
	// exist, it returns ErrObjectNotFound. Use it, for example, to check that a client actually completed an upload
	// made with a PresignedPutURL before saving the key in the database.
	Stat(ctx context.Context, key string) (ObjectInfo, error)

	// Delete removes the object stored under key. Deleting an object that does not exist is not an error.
	Delete(ctx context.Context, key string) error

	// PresignedGetURL returns a temporary URL that allows anyone who has it to download the object stored under key
	// directly from the storage service, without going through the web API, until expiry elapses. The object
	// storage endpoint must be reachable by whoever uses the URL (for example, the browser).
	PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// PresignedPutURL returns a temporary URL that allows anyone who has it to upload one object under key directly
	// to the storage service, with an HTTP PUT request, without going through the web API, until expiry elapses.
	//
	// size (in bytes) and contentType are part of the signature: the storage service rejects the upload unless the
	// request body is exactly size bytes and the request has the header "Content-Type: <contentType>". The web API
	// should check them (for example, a maximum size and the allowed image types) before calling this method.
	//
	// As for PresignedGetURL, the object storage endpoint must be reachable by whoever uses the URL.
	PresignedPutURL(ctx context.Context, key string, size int64, contentType string, expiry time.Duration) (string, error)

	// Ping checks that the storage service is reachable and that the bucket exists.
	Ping(ctx context.Context) error
}

// ObjectInfo contains the metadata of a stored object.
type ObjectInfo struct {
	// Size is the size of the object, in bytes.
	Size int64
	// ContentType is the MIME type given when the object was stored.
	ContentType string
}

// ErrObjectNotFound is returned by Get when no object is stored under the requested key.
var ErrObjectNotFound = errors.New("object not found")
