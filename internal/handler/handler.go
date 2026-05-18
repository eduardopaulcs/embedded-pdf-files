package handler

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"

	"embedded-pdf-files/internal/config"
	"embedded-pdf-files/internal/middleware"
	"embedded-pdf-files/internal/model"
	"embedded-pdf-files/internal/service"
)

type Deps struct {
	Service       *service.Service
	RateLimiter   *middleware.RateLimiter
	StaticFS      fs.FS
	TemplatesFS   fs.FS
	ResourcesFS   fs.FS
	Config        *config.Config
	StaticVersion string
}

type Handler struct {
	svc           *service.Service
	rateLimiter   *middleware.RateLimiter
	staticFS      fs.FS
	templatesFS   fs.FS
	resourcesFS   fs.FS
	config        *config.Config
	staticVersion string
}

func New(deps Deps) *Handler {
	return &Handler{
		svc:           deps.Service,
		rateLimiter:   deps.RateLimiter,
		staticFS:      deps.StaticFS,
		templatesFS:   deps.TemplatesFS,
		resourcesFS:   deps.ResourcesFS,
		config:        deps.Config,
		staticVersion: deps.StaticVersion,
	}
}

func (h *Handler) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(h.staticFS))))

	// TODO: re-add /robots.txt and /sitemap.xml handlers
	// once a production domain (SITE_URL) is configured.

	mux.HandleFunc("/", h.HandleHome)
	mux.HandleFunc("/terms", h.HandleTerms)
	mux.HandleFunc("/privacy", h.HandlePrivacy)

	mux.HandleFunc("/upload", h.HandleUpload)
	mux.HandleFunc("/download", h.HandleDownload)

	return mux
}

func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFS(h.templatesFS, "templates/index.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, model.PageData{
		GAMeasurementID:          h.config.GAMeasurementID,
		GoogleAdsID:              h.config.GoogleAdsID,
		GoogleAdsConversionLabel: h.config.GoogleAdsConversionLabel,
		CacheTTL:                 service.HumanDuration(h.config.UploadLimitWindow),
		DonationURL:              h.config.DonationURL,
		SiteURL:                  h.config.SiteURL,
		StaticVersion:            h.staticVersion,
	})
}

func (h *Handler) HandleTerms(w http.ResponseWriter, r *http.Request) {
	mdFile, err := fs.ReadFile(h.resourcesFS, "resources/markdown/terms.md")
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	htmlContent, err := service.MarkdownToHTML(mdFile)
	if err != nil {
		log.Printf("Error parsing markdown: %v", err)
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFS(h.templatesFS, "templates/terms.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, model.PageData{
		Content:                  htmlContent,
		GAMeasurementID:          h.config.GAMeasurementID,
		GoogleAdsID:              h.config.GoogleAdsID,
		GoogleAdsConversionLabel: h.config.GoogleAdsConversionLabel,
		DonationURL:              h.config.DonationURL,
		SiteURL:                  h.config.SiteURL,
		StaticVersion:            h.staticVersion,
	})
}

func (h *Handler) HandlePrivacy(w http.ResponseWriter, r *http.Request) {
	mdFile, err := fs.ReadFile(h.resourcesFS, "resources/markdown/privacy.md")
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	htmlContent, err := service.MarkdownToHTML(mdFile)
	if err != nil {
		log.Printf("Error parsing markdown: %v", err)
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFS(h.templatesFS, "templates/privacy.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, model.PageData{
		Content:                  htmlContent,
		GAMeasurementID:          h.config.GAMeasurementID,
		GoogleAdsID:              h.config.GoogleAdsID,
		GoogleAdsConversionLabel: h.config.GoogleAdsConversionLabel,
		DonationURL:              h.config.DonationURL,
		SiteURL:                  h.config.SiteURL,
		StaticVersion:            h.staticVersion,
	})
}

func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		json.NewEncoder(w).Encode(model.UploadResponse{Error: "File too large (max 10MB)"})
		return
	}

	file, header, err := r.FormFile("pdf")
	if err != nil {
		json.NewEncoder(w).Encode(model.UploadResponse{Error: "Invalid file"})
		return
	}
	defer file.Close()

	pdfBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		json.NewEncoder(w).Encode(model.UploadResponse{Error: "Unexpected error"})
		return
	}

	result, err := h.svc.ProcessUpload(service.UploadInput{
		PDFData:  pdfBytes,
		FileName: header.Filename,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidFile) {
			json.NewEncoder(w).Encode(model.UploadResponse{Error: "Invalid file"})
			return
		}
		if errors.Is(err, service.ErrNoAttachments) {
			json.NewEncoder(w).Encode(model.UploadResponse{Files: []string{}})
			return
		}
		log.Printf("Error processing upload: %v", err)
		json.NewEncoder(w).Encode(model.UploadResponse{Error: "Unexpected error"})
		return
	}

	if !h.rateLimiter.Allow(middleware.ClientIP(r)) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(model.UploadResponse{
			Error: "Rate limit exceeded. Please try again later.",
		})
		return
	}

	id := h.svc.CreateSession(result)
	json.NewEncoder(w).Encode(model.UploadResponse{
		ID:     id,
		Files:  result.FileNames,
		HasZip: len(result.ZipData) > 0,
	})
}

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	all := r.URL.Query().Get("all") == "true"
	filename := r.URL.Query().Get("filename")

	out, err := h.svc.ProcessDownload(service.DownloadInput{
		SessionID: id,
		FileName:  filename,
		All:       all,
	})
	if err != nil {
		if errors.Is(err, service.ErrSessionExpired) {
			http.Error(w, "Session expired or not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrNoZip) || errors.Is(err, service.ErrFileNotFound) {
			http.Error(w, "File not found", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrMissingFileName) {
			http.Error(w, "Missing parameters", http.StatusBadRequest)
			return
		}
		log.Printf("Error processing download: %v", err)
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", out.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+out.FileName+"\"")
	w.Write(out.Data)
}
