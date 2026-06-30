package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/raven-clown/raven-webmarket/backend/internal/config"
)

type Service struct {
	cfg    *config.Config
	client *minio.Client
}

func New(cfg *config.Config) (*Service, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}
	s := &Service{cfg: cfg, client: client}
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return s, nil
	}
	if !exists {
		_ = client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{})
	}
	return s, nil
}

func (s *Service) UploadSlip(ctx context.Context, ref, base64Data string) (string, error) {
	raw := base64Data
	if idx := strings.Index(raw, ","); idx >= 0 {
		raw = raw[idx+1:]
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}
	objectName := fmt.Sprintf("%s-%s.jpg", ref, uuid.New().String()[:8])
	_, err = s.client.PutObject(ctx, s.cfg.MinioBucket, objectName,
		bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.cfg.MinioPublicURL, "/"), objectName), nil
}
