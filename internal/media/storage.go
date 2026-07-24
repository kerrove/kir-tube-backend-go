package media

import (
	"context"
	"errors"
	"io"

	"go/kir-tube/configs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ErrObjectNotFound is returned by ObjectStorage.Get when the key does not exist.
// The handler maps it to a 404.
var ErrObjectNotFound = errors.New("media: object not found")

// ObjectInfo carries the metadata the HTTP layer needs to serve an object.
type ObjectInfo struct {
	ContentType string
	Size        int64
}

// ObjectStorage abstracts the media blob store so the service and handler depend
// on an interface rather than the concrete MinIO client (DIP). Keys mirror the
// public URL layout: "<folder>/<name>" and "<folder>/<resolution>/<name>".
type ObjectStorage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error)
}

// MinioStorage is the MinIO/S3-backed ObjectStorage implementation.
type MinioStorage struct {
	client *minio.Client
	bucket string
}

// NewMinioStorage connects to MinIO and ensures the target bucket exists.
func NewMinioStorage(cfg configs.StorageConfig) (*MinioStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &MinioStorage{client: client, bucket: cfg.Bucket}, nil
}

func (s *MinioStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *MinioStorage) Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, ObjectInfo{}, err
	}

	// GetObject is lazy; Stat forces the request so a missing key fails here
	// instead of on the first Read, and gives us the content type/size.
	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, ObjectInfo{}, ErrObjectNotFound
		}
		return nil, ObjectInfo{}, err
	}

	return obj, ObjectInfo{ContentType: stat.ContentType, Size: stat.Size}, nil
}
