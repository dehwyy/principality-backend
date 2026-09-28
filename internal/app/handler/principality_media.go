package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
)

func isImage(contentType string) bool {
	imageTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, t := range imageTypes {
		if contentType == t {
			return true
		}
	}
	return false
}

func isVideo(contentType string) bool {
	videoTypes := []string{
		"video/mp4",
		"video/webm",
		"video/quicktime",
	}

	for _, t := range videoTypes {
		if contentType == t {
			return true
		}
	}
	return false
}

func validateFileUpload(header *multipart.FileHeader, isExpectedType func(string) bool) (string, int, error) {
	file, err := header.Open()
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("не удалось получить файл")
	}

	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("не удалось прочитать файл")
	}

	contentType := http.DetectContentType(buffer)

	if !isExpectedType(contentType) {
		return "", http.StatusBadRequest, fmt.Errorf("файл %s имеет неподходящий тип %s", header.Filename, contentType)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("ошибка при обработке файла")
	}

	return contentType, 0, nil
}
