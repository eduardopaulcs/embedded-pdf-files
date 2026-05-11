package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"embedded-pdf-files/internal/extractor"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

//go:embed resources
var resourcesFS embed.FS

var fileCache *cache.Cache

// Rate limiter configuration (defaults, overridable via environment variables)
const (
	defaultUploadLimitMax    = 3
	defaultUploadLimitWindow = 10 * time.Minute
)

var (
	uploadLimitMax    int
	uploadLimitWindow time.Duration
	uploadLimiter     *cache.Cache
	uploadMu          sync.Mutex
)

func envInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}

func envDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func init() {
	uploadLimitMax = envInt("UPLOAD_LIMIT_MAX", defaultUploadLimitMax)
	uploadLimitWindow = envDuration("UPLOAD_LIMIT_WINDOW", defaultUploadLimitWindow)
	uploadLimiter = cache.New(2*uploadLimitWindow, uploadLimitWindow)
	fileCache = cache.New(uploadLimitWindow, uploadLimitWindow/2)
}

func humanDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	if minutes >= 60 {
		hours := minutes / 60
		mins := minutes % 60
		if mins == 0 {
			if hours == 1 {
				return "1 hour"
			}
			return fmt.Sprintf("%d hours", hours)
		}
		if hours == 1 {
			return fmt.Sprintf("1 hour %d minutes", mins)
		}
		return fmt.Sprintf("%d hours %d minutes", hours, mins)
	}
	if minutes == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}

type UploadResponse struct {
	ID     string   `json:"id"`
	Files  []string `json:"files"`
	HasZip bool     `json:"hasZip"`
	Error  string   `json:"error,omitempty"`
}

type PageData struct {
	Content        template.HTML
	UmamiURL       string
	UmamiWebsiteID string
	CacheTTL       string
	DonationURL    string
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// checkUploadLimit verifies the IP has not exceeded the rate limit
// (uploadLimitMax uploads in the last uploadLimitWindow). If allowed,
// it records the current timestamp and returns true.
// It is safe for concurrent use.
func checkUploadLimit(ip string) bool {
	uploadMu.Lock()
	defer uploadMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-uploadLimitWindow)

	raw, found := uploadLimiter.Get(ip)
	var timestamps []time.Time
	if found {
		timestamps = raw.([]time.Time)
	}

	// Purge timestamps older than the window
	var active []time.Time
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			active = append(active, ts)
		}
	}

	if len(active) >= uploadLimitMax {
		uploadLimiter.Set(ip, active, cache.DefaultExpiration)
		return false
	}

	active = append(active, now)
	uploadLimiter.Set(ip, active, cache.DefaultExpiration)
	return true
}

func main() {
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub filesystem: %v", err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSubFS))))

	umamiURL := os.Getenv("UMAMI_URL")
	umamiWebsiteID := os.Getenv("UMAMI_WEBSITE_ID")
	donationURL := os.Getenv("DONATION_URL")

	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		data, err := resourcesFS.ReadFile("resources/robots.txt")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(data)
	})

	http.HandleFunc("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
		data, err := resourcesFS.ReadFile("resources/sitemap.xml")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Write(data)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmpl, err := template.ParseFS(templatesFS, "templates/index.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, PageData{
			UmamiURL:       umamiURL,
			UmamiWebsiteID: umamiWebsiteID,
			CacheTTL:       humanDuration(uploadLimitWindow),
			DonationURL:    donationURL,
		})
	})

	http.HandleFunc("/terms", func(w http.ResponseWriter, r *http.Request) {
		mdFile, err := resourcesFS.ReadFile("resources/markdown/terms.md")
		if err != nil {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		var buf bytes.Buffer
		md := goldmark.New(
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
			),
		)
		if err := md.Convert(mdFile, &buf); err != nil {
			log.Printf("Error parsing markdown: %v", err)
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFS(templatesFS, "templates/terms.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, PageData{
			Content:        template.HTML(buf.String()),
			UmamiURL:       umamiURL,
			UmamiWebsiteID: umamiWebsiteID,
			DonationURL:    donationURL,
		})
	})

	http.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		mdFile, err := resourcesFS.ReadFile("resources/markdown/privacy.md")
		if err != nil {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}

		var buf bytes.Buffer
		md := goldmark.New(
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
			),
		)
		if err := md.Convert(mdFile, &buf); err != nil {
			log.Printf("Error parsing markdown: %v", err)
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFS(templatesFS, "templates/privacy.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, PageData{
			Content:        template.HTML(buf.String()),
			UmamiURL:       umamiURL,
			UmamiWebsiteID: umamiWebsiteID,
			DonationURL:    donationURL,
		})
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			json.NewEncoder(w).Encode(UploadResponse{Files: []string{}, Error: "File too large (max 10MB)"})
			return
		}

		file, header, err := r.FormFile("pdf")
		if err != nil {
			json.NewEncoder(w).Encode(UploadResponse{Files: []string{}, Error: "Invalid file"})
			return
		}
		defer file.Close()

		pdfBytes, err := io.ReadAll(file)
		if err != nil {
			log.Printf("Error reading file: %v", err)
			json.NewEncoder(w).Encode(UploadResponse{Files: []string{}, Error: "Unexpected error"})
			return
		}

		if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
			json.NewEncoder(w).Encode(UploadResponse{Files: []string{}, Error: "Invalid file"})
			return
		}

		result, err := extractor.Extract(pdfBytes, header.Filename)
		if err != nil {
			log.Printf("Error extracting files: %v", err)
			json.NewEncoder(w).Encode(UploadResponse{Files: []string{}, Error: "Unexpected error"})
			return
		}

		if len(result.FileNames) == 0 {
			json.NewEncoder(w).Encode(UploadResponse{Files: []string{}})
			return
		}

		if !checkUploadLimit(clientIP(r)) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(UploadResponse{
				Error: "Rate limit exceeded. Please try again later.",
			})
			return
		}

		id := uuid.New().String()
		fileCache.Set(id, result, cache.DefaultExpiration)
		json.NewEncoder(w).Encode(UploadResponse{
			ID:     id,
			Files:  result.FileNames,
			HasZip: len(result.ZipData) > 0,
		})
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing parameters", http.StatusBadRequest)
			return
		}

		data, found := fileCache.Get(id)
		if !found {
			http.Error(w, "Session expired or not found", http.StatusNotFound)
			return
		}

		result, ok := data.(*extractor.Result)
		if !ok {
			log.Printf("Invalid session data type")
			http.Error(w, "Unexpected error", http.StatusInternalServerError)
			return
		}

		// Check if requesting ZIP
		if r.URL.Query().Get("all") == "true" {
			if len(result.ZipData) == 0 {
				http.Error(w, "File not found", http.StatusBadRequest)
				return
			}
			baseName := strings.TrimSuffix(result.SourceFile, filepath.Ext(result.SourceFile))
			zipName := baseName + "_embedded_files.zip"
			w.Header().Set("Content-Disposition", "attachment; filename=\""+zipName+"\"")
			w.Header().Set("Content-Type", "application/zip")
			w.Write(result.ZipData)
			return
		}

		filename := r.URL.Query().Get("filename")
		if filename == "" {
			http.Error(w, "Missing parameters", http.StatusBadRequest)
			return
		}

		content, exists := result.Files[filename]
		if !exists {
			http.Error(w, "File not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(content)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server started at http://localhost:%s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
