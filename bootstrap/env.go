package bootstrap

import (
	"errors"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Env struct {
	AppEnv                 string
	ServerAddress          string
	ContextTimeout         int
	MongoURI               string
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPass                 string
	DBName                 string
	DBAuthSource           string
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
		MongoURI:               getFirstNonEmpty("MONGODB_URI", "DATABASE_URL"),
		DBHost:                 getEnv("DB_HOST", ""),
		DBPort:                 getEnv("DB_PORT", ""),
		DBUser:                 getEnv("DB_USER", ""),
		DBPass:                 getEnv("DB_PASS", ""),
		DBName:                 getEnv("DB_NAME", ""),
		DBAuthSource:           getEnv("DB_AUTH_SOURCE", ""),
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

	if strings.TrimSpace(env.DBName) == "" && strings.TrimSpace(env.MongoURI) != "" {
		if db := extractDBNameFromURI(env.MongoURI); db != "" {
			env.DBName = db
		}
	}

	if strings.EqualFold(env.AppEnv, "development") {
		log.Println("The App is running in development env")
	}

	if env.DBAuthSource == "" {
		env.DBAuthSource = env.DBName
		if env.DBAuthSource == "" {
			env.DBAuthSource = "admin"
		}
	}

	return env
}

func extractDBNameFromURI(uri string) string {
	parsed, err := url.Parse(strings.TrimSpace(uri))
	if err != nil {
		return ""
	}

	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return ""
	}

	if idx := strings.Index(path, "/"); idx != -1 {
		path = path[:idx]
	}

	return path
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

func getFirstNonEmpty(keys ...string) string {
	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	return ""
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
