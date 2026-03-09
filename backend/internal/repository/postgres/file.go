package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/atlas/backend/internal/domain"
)

type FileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	query := `
		INSERT INTO files (id, message_id, user_id, filename, original_name, mime_type, size_bytes, storage_path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		file.ID,
		file.MessageID,
		file.UserID,
		file.Filename,
		file.OriginalName,
		file.MimeType,
		file.SizeBytes,
		file.StoragePath,
		file.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("FileRepository.Create: %w", err)
	}
	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id string) (*domain.File, error) {
	query := `
		SELECT id, message_id, user_id, filename, original_name, mime_type, size_bytes, storage_path, created_at
		FROM files WHERE id = $1
	`
	file := &domain.File{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&file.ID,
		&file.MessageID,
		&file.UserID,
		&file.Filename,
		&file.OriginalName,
		&file.MimeType,
		&file.SizeBytes,
		&file.StoragePath,
		&file.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("FileRepository.GetByID: %w", err)
	}
	return file, nil
}

func (r *FileRepository) GetByMessageID(ctx context.Context, messageID string) ([]*domain.File, error) {
	query := `
		SELECT id, message_id, user_id, filename, original_name, mime_type, size_bytes, storage_path, created_at
		FROM files WHERE message_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("FileRepository.GetByMessageID: %w", err)
	}
	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		file := &domain.File{}
		if err := rows.Scan(
			&file.ID,
			&file.MessageID,
			&file.UserID,
			&file.Filename,
			&file.OriginalName,
			&file.MimeType,
			&file.SizeBytes,
			&file.StoragePath,
			&file.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("FileRepository.GetByMessageID scan: %w", err)
		}
		files = append(files, file)
	}
	return files, nil
}

func (r *FileRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM files WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("FileRepository.Delete: %w", err)
	}
	return nil
}
