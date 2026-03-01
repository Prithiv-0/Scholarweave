package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port                string
	DatabaseURL         string
	OpenAlexBaseURL     string
	OpenAlexHTTPTimeout time.Duration
	JWTSecret           string
	JWTExpiration       time.Duration
	FrontendOrigin      string
}

func Load() Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3000"
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))

	openAlexBaseURL := strings.TrimSpace(os.Getenv("OPENALEX_BASEURL"))
	if openAlexBaseURL == "" {
		openAlexBaseURL = "https://api.openalex.org"
	}

	openAlexHTTPTimeout := 12 * time.Second
	if rawTimeout := strings.TrimSpace(os.Getenv("OPENALEX_HTTP_TIMEOUT")); rawTimeout != "" {
		if parsedTimeout, err := time.ParseDuration(rawTimeout); err == nil && parsedTimeout > 0 {
			openAlexHTTPTimeout = parsedTimeout
		}
	}

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		if os.Getenv("RENDER") != "" || strings.EqualFold(os.Getenv("ENV"), "production") {
			fmt.Fprintln(os.Stderr, "FATAL: JWT_SECRET must be set in production")
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "WARNING: JWT_SECRET not set, using insecure default (dev only)")
		jwtSecret = "scholarweave-dev-secret"
	}

	jwtExpiration := 24 * time.Hour
	if rawJWTExpiration := strings.TrimSpace(os.Getenv("JWT_EXPIRATION")); rawJWTExpiration != "" {
		if parsedExpiration, err := time.ParseDuration(rawJWTExpiration); err == nil && parsedExpiration > 0 {
			jwtExpiration = parsedExpiration
		}
	}

	frontendOrigin := strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN"))
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}

	return Config{
		Port:                port,
		DatabaseURL:         databaseURL,
		OpenAlexBaseURL:     strings.TrimRight(openAlexBaseURL, "/"),
		OpenAlexHTTPTimeout: openAlexHTTPTimeout,
		JWTSecret:           jwtSecret,
		JWTExpiration:       jwtExpiration,
		FrontendOrigin:      frontendOrigin,
	}
}
