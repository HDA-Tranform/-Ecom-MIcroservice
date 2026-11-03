package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	api "github.com/linhhuynhcoding/jss-microservices/rpc/gen/product"
	"go.uber.org/zap"
)

func (s *Service) UploadFile(ctx context.Context, req *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	log := s.logger.With(zap.String("func", "UploadFile"))

	// Generate a unique file ID
	fileID := generateFileID()

	// Create uploads directory if it doesn't exist
	uploadDir := s.cfg.UploadFolder
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Error("failed to create upload directory", zap.Error(err))
		return nil, fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Create file path with unique ID
	filename := fmt.Sprintf("%s_%s", fileID, req.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// Write file to disk
	if err := os.WriteFile(filePath, req.FileData, 0644); err != nil {
		log.Error("failed to write file", zap.Error(err))
		return nil, fmt.Errorf("failed to write file: %v", err)
	}

	cloud := s.adapter.cloudinaryAdapter
	if cloud == nil {
		return nil, fmt.Errorf("failed to upload file")
	}

	url, err := s.adapter.cloudinaryAdapter.UploadImage(ctx, filePath)
	if err != nil {
		log.Error("failed to upload file to couldinary", zap.Error(err))
		return nil, fmt.Errorf("failed to upload file to couldinary: %v", err)
	}
	log.Info("File uploaded successfully:",
		zap.Any("url", url))

	return &api.UploadFileResponse{
		Message:  "File uploaded successfully",
		FileId:   fileID,
		Filename: req.Filename,
		FileSize: req.FileSize,
		FileUrl:  url,
	}, nil
}

func generateFileID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *Service) UploadFileHTTP(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file_data")
	if err != nil {
		http.Error(w, "failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// đọc content type nếu có
	contentType := handler.Header.Get("Content-Type")

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// gọi lại RPC logic
	resp, err := s.UploadFile(r.Context(), &api.UploadFileRequest{
		FileData:    data,
		Filename:    handler.Filename,
		ContentType: contentType,
		FileSize:    int64(len(data)),
	})
	if err != nil {
		http.Error(w, "failed to upload file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
