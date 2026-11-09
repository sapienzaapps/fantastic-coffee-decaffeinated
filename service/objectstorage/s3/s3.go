/*
Package s3 implements objectstorage.ObjectStorage on top of any S3-compatible service (Garage, MinIO, AWS S3, ...),
using the github.com/minio/minio-go/v7 client.
*/
package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"git.sapienzaapps.it/fantasticcoffee/fantastic-coffee-decaffeinated/service/objectstorage"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config contains the parameters to connect to the S3-compatible service. All fields are required.
type Config struct {
	// Endpoint is the URL of the S3 API, including the scheme, for example "http://localhost:3900" or
	// "https://s3.example.com". The scheme selects whether TLS (https) is used.
	Endpoint string
	// Region is the region used to sign requests. For Garage, it is the s3_region value in garage.toml.
	Region string
	// Bucket is the name of the bucket where objects are stored. It must already exist.
	Bucket string
	// AccessKeyID and SecretAccessKey are the credentials of the access key.
	AccessKeyID     string
	SecretAccessKey string
}

// s3Storage is the S3 implementation of objectstorage.ObjectStorage.
type s3Storage struct {
	client *minio.Client
	bucket string
}

// New connects to the S3-compatible service described by cfg, checks that the bucket exists and returns it as an
// objectstorage.ObjectStorage.
func New(ctx context.Context, cfg Config) (objectstorage.ObjectStorage, error) {
	switch {
	case cfg.Endpoint == "":
		return nil, errors.New("the S3 endpoint is required")
	case cfg.Region == "":
		return nil, errors.New("the S3 region is required")
	case cfg.Bucket == "":
		return nil, errors.New("the S3 bucket is required")
	case cfg.AccessKeyID == "" || cfg.SecretAccessKey == "":
		return nil, errors.New("the S3 access key ID and secret access key are required")
	}

	endpoint, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid S3 endpoint %q: %w", cfg.Endpoint, err)
	}
	if endpoint.Scheme != "http" && endpoint.Scheme != "https" {
		return nil, fmt.Errorf("invalid S3 endpoint %q: it must start with http:// or https://", cfg.Endpoint)
	}
	if endpoint.Host == "" || (endpoint.Path != "" && endpoint.Path != "/") || endpoint.RawQuery != "" {
		return nil, fmt.Errorf("invalid S3 endpoint %q: it must contain only the scheme, the host and the port",
			cfg.Endpoint)
	}

	client, err := minio.New(endpoint.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: endpoint.Scheme == "https",
		Region: cfg.Region,
		// Path-style URLs (http://host/bucket/key) work with every S3-compatible service, including Garage, without
		// any DNS configuration.
		BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		return nil, fmt.Errorf("creating the S3 client: %w", err)
	}

	s := &s3Storage{client: client, bucket: cfg.Bucket}
	if err := s.Ping(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *s3Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *s3Storage) Get(ctx context.Context, key string) (io.ReadCloser, objectstorage.ObjectInfo, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, objectstorage.ObjectInfo{}, err
	}
	// GetObject does not contact the server until the object is read: Stat does, and reports a missing object.
	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, objectstorage.ObjectInfo{}, convertError(err)
	}
	return obj, objectstorage.ObjectInfo{Size: info.Size, ContentType: info.ContentType}, nil
}

func (s *s3Storage) Stat(ctx context.Context, key string) (objectstorage.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return objectstorage.ObjectInfo{}, convertError(err)
	}
	return objectstorage.ObjectInfo{Size: info.Size, ContentType: info.ContentType}, nil
}

func (s *s3Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *s3Storage) PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *s3Storage) PresignedPutURL(ctx context.Context, key string, size int64, contentType string,
	expiry time.Duration) (string, error) {
	if size < 0 {
		return "", errors.New("the size of the object to upload must not be negative")
	}
	if contentType == "" {
		return "", errors.New("the content type of the object to upload is required")
	}

	// Headers listed here become part of the signature: the storage service rejects (HTTP 403) any upload whose
	// Content-Length or Content-Type differs from these values.
	signedHeaders := http.Header{}
	signedHeaders.Set("Content-Length", strconv.FormatInt(size, 10))
	signedHeaders.Set("Content-Type", contentType)

	u, err := s.client.PresignHeader(ctx, http.MethodPut, s.bucket, key, expiry, nil, signedHeaders)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// convertError translates the "object not found" errors of the S3 client into objectstorage.ErrObjectNotFound.
func convertError(err error) error {
	if minio.ToErrorResponse(err).Code == "NoSuchKey" {
		return objectstorage.ErrObjectNotFound
	}
	return err
}

func (s *s3Storage) Ping(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		if minio.ToErrorResponse(err).StatusCode == http.StatusBadRequest {
			// The server gives no details here (the request has no response body): a wrong region is the most
			// common cause.
			return fmt.Errorf("checking the S3 bucket %q: %w (check that the region matches the one configured "+
				"in the S3 service, e.g. s3_region in garage.toml)", s.bucket, err)
		}
		return fmt.Errorf("checking the S3 bucket %q: %w", s.bucket, err)
	}
	if !exists {
		return fmt.Errorf("the S3 bucket %q does not exist", s.bucket)
	}
	return nil
}
