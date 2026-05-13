package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                    string
	UploadLimitMax          int
	UploadLimitWindow       time.Duration
	GAMeasurementID         string
	GoogleAdsID             string
	GoogleAdsConversionLabel string
	DonationURL             string
}

func envStr(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

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

func Load() *Config {
	return &Config{
		Port:              envStr("PORT", "8080"),
		UploadLimitMax:    envInt("UPLOAD_LIMIT_MAX", 3),
		UploadLimitWindow: envDuration("UPLOAD_LIMIT_WINDOW", 10*time.Minute),
		GAMeasurementID:          os.Getenv("GA_MEASUREMENT_ID"),
		GoogleAdsID:              os.Getenv("GOOGLE_ADS_ID"),
		GoogleAdsConversionLabel: os.Getenv("GOOGLE_ADS_CONVERSION_LABEL"),
		DonationURL:       os.Getenv("DONATION_URL"),
	}
}
