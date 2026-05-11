package service

import (
	"bytes"
	"fmt"
)

type UploadInput struct {
	PDFData  []byte
	FileName string
}

func (s *Service) ProcessUpload(input UploadInput) (*Result, error) {
	if !bytes.HasPrefix(input.PDFData, []byte("%PDF")) {
		return nil, ErrInvalidFile
	}

	result, err := Extract(input.PDFData, input.FileName)
	if err != nil {
		return nil, fmt.Errorf("extraction failed: %w", err)
	}

	if len(result.FileNames) == 0 {
		return nil, ErrNoAttachments
	}

	return result, nil
}
