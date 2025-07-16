package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// storageEmployeeImages - путь к директории для хранения изображений сотрудников

// Утилита: сохраняет загруженный файл и возвращает имя сохранённого файла
func saveUploadedPhoto(fileHeader *multipart.FileHeader, empID int) (string, error) {
	// Открываем загруженный файл
	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("не удалось открыть загруженный файл: %v", err)
	}
	defer src.Close()

	// Определяем расширение из исходного имени
	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("emp_%d%s", empID, ext)
	fullPath := filepath.Join(storageEmployeeImages, filename)

	// Создаём файл на диске
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("ошибка создания файла на сервере: %v", err)
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("ошибка сохранения файла: %v", err)
	}

	return filename, nil
}

// Утилита: удаляет файл по имени
// Возвращает ошибку, если файл не найден или не может быть удалён
func deletePhoto(filename string) error {

	fullPath := filepath.Join(storageEmployeeImages, filename)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("ошибка удаления файла %s: %v", filename, err)
	}
	return nil
}
