package bootstrap

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Env struct {
	AppEnv                 string
	ServerAddress          string
	ContextTimeout         int
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPass                 string
	DBName                 string
	AccessTokenExpiryHour  int
	RefreshTokenExpiryHour int
	AccessTokenSecret      string
	RefreshTokenSecret     string
	CookieDomain           string
	CookieSecure           bool
	CookieSameSite         string
	CORSAllowedOrigins     string
	SuperAdminPhone        string
	SuperAdminPassword     string
	SuperAdminName         string
	SuperAdminEmail        string
	SuperAdminOrganization string
}

func NewEnv() *Env {
	loadLocalEnv()

	env := &Env{
		AppEnv:                 getEnv("APP_ENV", "development"),
		ServerAddress:          getEnv("SERVER_ADDRESS", ":8080"),
		ContextTimeout:         getEnvAsInt("CONTEXT_TIMEOUT", 2),
		DBHost:                 getEnv("DB_HOST", ""),
		DBPort:                 getEnv("DB_PORT", ""),
		DBUser:                 getEnv("DB_USER", ""),
		DBPass:                 getEnv("DB_PASS", ""),
		DBName:                 getEnv("DB_NAME", ""),
		AccessTokenExpiryHour:  getEnvAsInt("ACCESS_TOKEN_EXPIRY_HOUR", 2),
		RefreshTokenExpiryHour: getEnvAsInt("REFRESH_TOKEN_EXPIRY_HOUR", 168),
		AccessTokenSecret:      getEnv("ACCESS_TOKEN_SECRET", ""),
		RefreshTokenSecret:     getEnv("REFRESH_TOKEN_SECRET", ""),
		CookieDomain:           getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:           getEnvAsBool("COOKIE_SECURE", false),
		CookieSameSite:         getEnv("COOKIE_SAME_SITE", "lax"),
		CORSAllowedOrigins:     getEnv("CORS_ALLOWED_ORIGINS", ""),
		SuperAdminPhone:        getEnv("SUPER_ADMIN_PHONE", ""),
		SuperAdminPassword:     getEnv("SUPER_ADMIN_PASSWORD", ""),
		SuperAdminName:         getEnv("SUPER_ADMIN_NAME", ""),
		SuperAdminEmail:        getEnv("SUPER_ADMIN_EMAIL", ""),
		SuperAdminOrganization: getEnv("SUPER_ADMIN_ORGANIZATION", ""),
	}

	if strings.EqualFold(env.AppEnv, "development") {
		log.Println("The App is running in development env")
	}

	return env
}

func loadLocalEnv() {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENVIRONMENT")))
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	}
	if env == "" {
		env = "development"
	}
	if env == "production" {
		return
	}

	if err := godotenv.Load(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("warning: could not load .env file: %v", err)
		}
	}
}

func getEnv(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(valueStr) == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(strings.TrimSpace(valueStr))
	if err != nil {
		log.Fatalf("invalid value for %s: %v", key, err)
	}
	return value
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(valueStr) == "" {
		return defaultVal
	}
	value, err := strconv.ParseBool(strings.TrimSpace(valueStr))
	if err != nil {
		log.Fatalf("invalid value for %s: %v", key, err)
	}
	return value
}
