package imagehelper

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
)

func IsImage(file multipart.File) (string, error) {
	signature, err := getFileSignature(file, 512)
	if err != nil {
		return "", fmt.Errorf("failed to read file signature: %w", err)
	}

	format, err := detectImageFormat(signature)
	if err != nil {
		return "", err
	}

	// 2. Reset file pointer to the beginning for image.DecodeConfig
	if _, err := file.Seek(0, io.SeekStart); err != nil { // Crucial: Reset the file pointer
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// 3. image.DecodeConfig (now reads from the beginning of the file)
	if _, _, err := image.DecodeConfig(file); err != nil {
		return "", fmt.Errorf("image.DecodeConfig failed: %w", err)
	}

	return format, nil
}

func getFileSignature(file io.Reader, size int) ([]byte, error) {
	header := make([]byte, size)
	n, err := io.ReadFull(file, header)
	if err == io.ErrUnexpectedEOF && n < size {
		return nil, fmt.Errorf("get file signature error, fail to small to check the signature %w", err)
	} else if err != nil {
		return nil, err
	}
	return header, nil
}

func detectImageFormat(signature []byte) (string, error) {
	// JPEG
	if bytes.HasPrefix(signature, []byte{0xFF, 0xD8}) {
		return "jpeg", nil
	}

	// PNG
	if bytes.HasPrefix(signature, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "png", nil
	}

	return "", errors.New("unsupported image format")
}
