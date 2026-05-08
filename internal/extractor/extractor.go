package extractor

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

var sanitizeRegex = regexp.MustCompile(`[^a-zA-Z0-9_.]`)
var multiUnderscore = regexp.MustCompile(`_+`)

func sanitizeFileName(name string) string {
	// Get extension with dot
	ext := filepath.Ext(name)
	// Base name without extension
	base := strings.TrimSuffix(name, ext)

	// Sanitize base: only letters, numbers, underscores
	base = sanitizeRegex.ReplaceAllString(base, "_")
	// Replace multiple underscores with single one
	base = multiUnderscore.ReplaceAllString(base, "_")
	// Trim underscores from start and end
	base = strings.Trim(base, "_")

	// If no extension, return base only
	if ext == "" {
		return base
	}

	// Extension: remove the dot, keep only letters
	ext = strings.TrimPrefix(ext, ".")
	ext = regexp.MustCompile(`[^a-zA-Z]`).ReplaceAllString(ext, "")
	if ext == "" {
		return base
	}

	return base + "." + ext
}

type Result struct {
	Files      map[string][]byte
	FileNames  []string
	ZipData    []byte
	SourceFile string
}

func Extract(pdfBytes []byte, sourceFile string) (*Result, error) {
	tmpFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Write(pdfBytes)
	tmpFile.Close()
	defer os.Remove(tmpPath)

	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	reader := bytes.NewReader(pdfBytes)
	embeds, err := api.Attachments(reader, conf)
	if err != nil || len(embeds) == 0 {
		return &Result{Files: make(map[string][]byte), FileNames: []string{}}, nil
	}

	tmpDir, err := os.MkdirTemp("", "pdf-extract-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	err = api.ExtractAttachmentsFile(tmpPath, tmpDir, nil, conf)
	if err != nil {
		return &Result{Files: make(map[string][]byte), FileNames: []string{}}, nil
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil || len(files) == 0 {
		return &Result{Files: make(map[string][]byte), FileNames: []string{}}, nil
	}

	result := &Result{
		Files:      make(map[string][]byte),
		FileNames:  []string{},
		SourceFile: sanitizeFileName(sourceFile),
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(tmpDir, f.Name()))
		if err != nil {
			continue
		}
		sanitizedName := sanitizeFileName(f.Name())
		result.Files[sanitizedName] = content
		result.FileNames = append(result.FileNames, sanitizedName)
	}

	// Generate ZIP only if more than 1 file
	if len(result.Files) > 1 {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)

		for name, content := range result.Files {
			w, err := zipWriter.Create(name)
			if err != nil {
				continue
			}
			w.Write(content)
		}

		zipWriter.Close()
		result.ZipData = buf.Bytes()
	}

	return result, nil
}
