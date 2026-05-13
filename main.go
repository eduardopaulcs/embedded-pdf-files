package main

import (
	"io/fs"
	"log"
	"net/http"

	"embedded-pdf-files/internal/config"
	"embedded-pdf-files/internal/handler"
	"embedded-pdf-files/internal/middleware"
	"embedded-pdf-files/internal/service"
	"embedded-pdf-files/internal/store"
)

func main() {
	cfg := config.Load()

	st := store.New(cfg.UploadLimitWindow)
	svc := service.New(st)
	rl := middleware.New(cfg.UploadLimitWindow, cfg.UploadLimitMax)

	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub filesystem: %v", err)
	}

	staticVersion := service.ComputeStaticHash(staticFS)

	h := handler.New(handler.Deps{
		Service:       svc,
		RateLimiter:   rl,
		StaticFS:      staticSubFS,
		TemplatesFS:   templatesFS,
		ResourcesFS:   resourcesFS,
		Config:        cfg,
		StaticVersion: staticVersion,
	})

	log.Printf("Server started at http://localhost:%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, h.Mux()))
}
