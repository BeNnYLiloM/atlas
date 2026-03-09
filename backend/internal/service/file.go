package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/atlas/backend/internal/domain"
	"github.com/your-org/atlas/backend/pkg/storage"
)

const (
	MaxFileSizeFree = 10 * 1024 * 1024  // 10 MB для бесплатного плана
	MaxFileSizePro  = 100 * 1024 * 1024 // 100 MB для платного плана
)

type FileRepository interface {
	Create(ctx context.Context, file *domain.File) error
	GetByID(ctx context.Context, id string) (*domain.File, error)
	GetByMessageID(ctx context.Context, messageID string) ([]*domain.File, error)
	Delete(ctx context.Context, id string) error
}

type FileService struct {
	repo    FileRepository
	storage *storage.MinIOStorage
}

func NewFileService(repo FileRepository, storage *storage.MinIOStorage) *FileService {
	return &FileService{repo: repo, storage: storage}
}

func (s *FileService) Upload(ctx context.Context, userID string, filename string, reader io.Reader, size int64, mimeType string) (*domain.File, error) {
	if size > MaxFileSizeFree {
		return nil, fmt.Errorf("file size %d exceeds limit %d", size, MaxFileSizeFree)
	}

	ext := filepath.Ext(filename)
	objectName := fmt.Sprintf("files/%s/%s%s",
		time.Now().Format("2006/01/02"),
		uuid.New().String(),
		ext,
	)

	if _, err := s.storage.Upload(ctx, objectName, reader, size, mimeType); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	file := &domain.File{
		ID:           uuid.New().String(),
		UserID:       userID,
		Filename:     sanitizeFilename(filename),
		OriginalName: filename,
		MimeType:     mimeType,
		SizeBytes:    size,
		StoragePath:  objectName,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, file); err != nil {
		// Откатываем загрузку файла при ошибке сохранения в БД
		_ = s.storage.Delete(ctx, objectName)
		return nil, fmt.Errorf("failed to save file info: %w", err)
	}

	file.URL = s.storage.GetURL(objectName)
	return file, nil
}

func (s *FileService) GetByID(ctx context.Context, id string) (*domain.File, error) {
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	file.URL = s.storage.GetURL(file.StoragePath)
	return file, nil
}

func (s *FileService) GetByMessageID(ctx context.Context, messageID string) ([]*domain.File, error) {
	files, err := s.repo.GetByMessageID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		f.URL = s.storage.GetURL(f.StoragePath)
	}
	return files, nil
}

func (s *FileService) Delete(ctx context.Context, id string, requestingUserID string) error {
	file, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if file.UserID != requestingUserID {
		return ErrForbidden
	}

	if err := s.storage.Delete(ctx, file.StoragePath); err != nil {
		return fmt.Errorf("failed to delete from storage: %w", err)
	}

	return s.repo.Delete(ctx, id)
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}
