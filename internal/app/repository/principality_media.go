package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/minio/minio-go/v7"
)

var principalityMediaExtensions = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
	"image/gif":       ".gif",
	"video/mp4":       ".mp4",
	"video/webm":      ".webm",
	"video/quicktime": ".mov",
}

func NewPrincipalityMediaKey(contentType string) (string, error) {
	extension, known := principalityMediaExtensions[contentType]
	if !known {
		return "", fmt.Errorf("неподдерживаемый тип файла: %s", contentType)
	}

	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("не удалось сгенерировать имя файла: %w", err)
	}

	return "principality_" + hex.EncodeToString(randomBytes) + extension, nil
}

func (r *Repository) UploadPrincipalityMedia(header *multipart.FileHeader, mediaKey string) error {
	file, err := header.Open()
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	contentType := http.DetectContentType(buffer)

	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("ошибка перемещения по файловому потоку: %w", err)
	}

	_, err = r.minio.PutObject(
		context.Background(),
		r.minioBucketName,
		mediaKey,
		file,
		header.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
		})
	if err != nil {
		return fmt.Errorf("не удалось добавить объект в хранилище minio: %w", err)
	}

	return nil
}

func (r *Repository) RemovePrincipalityMedia(mediaKey string) error {
	err := r.minio.RemoveObject(context.Background(), r.minioBucketName, mediaKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("не удалось удалить объект из хранилища minio: %w", err)
	}
	return nil
}
