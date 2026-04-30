package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fayhub/pkg/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageDriver interface {
	Upload(key string, reader io.Reader) (string, error)
	Download(key string) (io.ReadCloser, error)
	Delete(key string) error
	GetURL(key string) string
	Exists(key string) bool
}

var currentDriver StorageDriver

func Init(cfg *config.Config) error {
	switch cfg.Storage.Driver {
	case "local":
		currentDriver = &LocalStorage{basePath: cfg.Storage.LocalPath}
	case "s3":
		currentDriver = &S3Storage{
			endpoint:  cfg.Storage.S3Endpoint,
			region:    cfg.Storage.S3Region,
			bucket:    cfg.Storage.S3Bucket,
			accessKey: cfg.Storage.S3AccessKey,
			secretKey: cfg.Storage.S3SecretKey,
			useSSL:    cfg.Storage.S3UseSSL,
		}
	default:
		currentDriver = &LocalStorage{basePath: cfg.Storage.LocalPath}
	}

	if ls, ok := currentDriver.(*LocalStorage); ok {
		if err := os.MkdirAll(ls.basePath, 0755); err != nil {
			return fmt.Errorf("创建上传目录失败: %w", err)
		}
	}

	return nil
}

func GetDriver() StorageDriver {
	return currentDriver
}

type LocalStorage struct {
	basePath string
}

func (l *LocalStorage) Upload(key string, reader io.Reader) (string, error) {
	fullPath := filepath.Join(l.basePath, key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return key, nil
}

func (l *LocalStorage) Download(key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, key)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	return f, nil
}

func (l *LocalStorage) Delete(key string) error {
	fullPath := filepath.Join(l.basePath, key)
	return os.Remove(fullPath)
}

func (l *LocalStorage) GetURL(key string) string {
	return "/api/files/" + key
}

func (l *LocalStorage) Exists(key string) bool {
	fullPath := filepath.Join(l.basePath, key)
	_, err := os.Stat(fullPath)
	return err == nil
}

type S3Storage struct {
	endpoint  string
	region    string
	bucket    string
	accessKey string
	secretKey string
	useSSL    bool
	client    *minio.Client
}

func (s *S3Storage) initClient() error {
	if s.client != nil {
		return nil
	}

	client, err := minio.New(s.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.accessKey, s.secretKey, ""),
		Region: s.region,
		Secure: s.useSSL,
	})
	if err != nil {
		return fmt.Errorf("创建S3客户端失败: %w", err)
	}

	s.client = client
	return nil
}

func (s *S3Storage) ensureBucket(ctx context.Context) error {
	if err := s.initClient(); err != nil {
		return err
	}

	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("检查Bucket失败: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: s.region}); err != nil {
			return fmt.Errorf("创建Bucket失败: %w", err)
		}
	}
	return nil
}

func (s *S3Storage) Upload(key string, reader io.Reader) (string, error) {
	ctx := context.Background()
	if err := s.ensureBucket(ctx); err != nil {
		return "", err
	}

	opts := minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	}

	info, err := s.client.PutObject(ctx, s.bucket, key, reader, -1, opts)
	if err != nil {
		return "", fmt.Errorf("上传文件到S3失败: %w", err)
	}

	_ = info
	return key, nil
}

func (s *S3Storage) Download(key string) (io.ReadCloser, error) {
	if err := s.initClient(); err != nil {
		return nil, err
	}

	ctx := context.Background()
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("从S3下载文件失败: %w", err)
	}

	return obj, nil
}

func (s *S3Storage) Delete(key string) error {
	if err := s.initClient(); err != nil {
		return err
	}

	ctx := context.Background()
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *S3Storage) GetURL(key string) string {
	scheme := "https"
	if !s.useSSL {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, key)
}

func (s *S3Storage) Exists(key string) bool {
	if err := s.initClient(); err != nil {
		return false
	}

	ctx := context.Background()
	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	return err == nil
}

func GenerateFileKey(filename string) string {
	ext := filepath.Ext(filename)
	now := time.Now()
	datePath := now.Format("2006/01/02")
	name := fmt.Sprintf("%d%s", now.UnixNano(), ext)
	return filepath.Join(datePath, name)
}

func IsAllowedType(filename string, allowedTypes string) bool {
	if allowedTypes == "" || allowedTypes == "*" {
		return true
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	types := strings.Split(allowedTypes, ",")
	for _, t := range types {
		if strings.TrimSpace(strings.ToLower(t)) == ext {
			return true
		}
	}
	return false
}
