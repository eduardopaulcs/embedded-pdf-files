package service

import (
	"errors"

	"embedded-pdf-files/internal/store"

	"github.com/google/uuid"
)

var (
	ErrInvalidFile     = errors.New("invalid file")
	ErrNoAttachments   = errors.New("no embedded files found")
	ErrSessionExpired  = errors.New("session expired or not found")
	ErrInvalidSession  = errors.New("invalid session data")
	ErrFileNotFound    = errors.New("file not found")
	ErrNoZip           = errors.New("ZIP not available")
	ErrMissingFileName = errors.New("missing filename")
)

type Service struct {
	store *store.SessionStore
}

func New(st *store.SessionStore) *Service {
	return &Service{store: st}
}

func (s *Service) CreateSession(result *Result) string {
	id := uuid.New().String()
	s.store.Set(id, result)
	return id
}
