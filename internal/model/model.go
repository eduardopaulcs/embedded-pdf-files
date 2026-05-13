package model

import "html/template"

type UploadResponse struct {
	ID     string   `json:"id"`
	Files  []string `json:"files"`
	HasZip bool     `json:"hasZip"`
	Error  string   `json:"error,omitempty"`
}

type PageData struct {
	Content                 template.HTML
	GAMeasurementID         string
	GoogleAdsID             string
	GoogleAdsConversionLabel string
	CacheTTL                string
	DonationURL             string
}
