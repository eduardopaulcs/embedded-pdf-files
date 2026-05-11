package service

import (
	"path/filepath"
	"strings"
)

type DownloadInput struct {
	SessionID string
	FileName  string
	All       bool
}

type DownloadOutput struct {
	Data        []byte
	ContentType string
	FileName    string
}

func (s *Service) ProcessDownload(input DownloadInput) (*DownloadOutput, error) {
	raw, ok := s.store.Get(input.SessionID)
	if !ok {
		return nil, ErrSessionExpired
	}

	result, ok := raw.(*Result)
	if !ok {
		return nil, ErrInvalidSession
	}

	if input.All {
		if len(result.ZipData) == 0 {
			return nil, ErrNoZip
		}
		baseName := strings.TrimSuffix(result.SourceFile, filepath.Ext(result.SourceFile))
		zipName := baseName + "_embedded_files.zip"
		return &DownloadOutput{
			Data:        result.ZipData,
			ContentType: "application/zip",
			FileName:    zipName,
		}, nil
	}

	if input.FileName == "" {
		return nil, ErrMissingFileName
	}

	content, exists := result.Files[input.FileName]
	if !exists {
		return nil, ErrFileNotFound
	}

	return &DownloadOutput{
		Data:        content,
		ContentType: "application/octet-stream",
		FileName:    input.FileName,
	}, nil
}
