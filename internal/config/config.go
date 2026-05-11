package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	UploadLimitMax    int
	UploadLimitWindow time.Duration
	UmamiURL          string
	UmamiWebsiteID    string
	DonationURL       string
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
		UmamiURL:          os.Getenv("UMAMI_URL"),
		UmamiWebsiteID:    os.Getenv("UMAMI_WEBSITE_ID"),
		DonationURL:       os.Getenv("DONATION_URL"),
	}
}
