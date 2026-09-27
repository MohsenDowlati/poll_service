package server

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	route "github.com/amitshekhariitbhu/go-backend-clean-architecture/api/route"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/bootstrap"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/internal/validation"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Run bootstraps the application and starts the HTTP server.
func Run() error {
	env := bootstrap.NewEnv()
	if err := env.Validate(); err != nil {
		return fmt.Errorf("invalid environment: %w", err)
	}
	app := bootstrap.AppWithEnv(env)

	db := app.Mongo.Database(env.DBName)
	defer app.CloseDBConnection()

	timeout := time.Duration(env.ContextTimeout) * time.Second

	if err := bootstrap.SeedSuperAdmin(env, db, timeout); err != nil {
		return fmt.Errorf("seed super admin: %w", err)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
			return validation.Phone(fl.Field().String())
		}); err != nil {
			return fmt.Errorf("register phone validator: %w", err)
		}
	}

	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token", "Idempotency-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	allowedOrigins := strings.Split(env.CORSAllowedOrigins, ",")
	formattedOrigins := expandLocalhostOrigins(allowedOrigins)

	if len(formattedOrigins) == 0 {
		if strings.EqualFold(env.AppEnv, "production") {
			return fmt.Errorf("no CORS origins configured")
		}
		corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	} else {
		corsConfig.AllowOrigins = formattedOrigins
	}
	for _, origin := range corsConfig.AllowOrigins {
		if origin == "*" && corsConfig.AllowCredentials {
			return fmt.Errorf("wildcard CORS origin cannot be used with credentials")
		}
	}

	setGinMode(env.AppEnv)
	router := gin.Default()
	if err := router.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("configure trusted proxies: %w", err)
	}
	router.Use(requestSecurityMiddleware())
	router.Use(csrfMiddleware())
	router.Use(rateLimitMiddleware())
	router.Use(cors.New(corsConfig))

	if !strings.EqualFold(env.AppEnv, "production") {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	route.Setup(env, timeout, db, router)

	server := &http.Server{
		Addr:              env.ServerAddress,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("run gin server: %w", err)
	}

	return nil
}

type rateWindow struct {
	start time.Time
	count int
}

func rateLimitMiddleware() gin.HandlerFunc {
	var mu sync.Mutex
	windows := make(map[string]rateWindow)
	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP() + ":" + c.Request.URL.Path
		limit := 120
		if strings.HasSuffix(c.Request.URL.Path, "/login") ||
			strings.HasSuffix(c.Request.URL.Path, "/signup") ||
			strings.HasSuffix(c.Request.URL.Path, "/refresh") ||
			strings.HasSuffix(c.Request.URL.Path, "/submit") {
			limit = 30
		}

		mu.Lock()
		window := windows[key]
		if window.start.IsZero() || now.Sub(window.start) >= time.Minute {
			window = rateWindow{start: now, count: 0}
		}
		window.count++
		windows[key] = window
		allowed := window.count <= limit
		if len(windows) > 10000 {
			for entryKey, entry := range windows {
				if now.Sub(entry.start) >= time.Minute {
					delete(windows, entryKey)
				}
			}
		}
		mu.Unlock()

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, domain.ErrorResponse{Message: "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func requestSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Next()
	}
}

func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if strings.HasSuffix(path, "/login") || strings.HasSuffix(path, "/signup") ||
			strings.HasSuffix(path, "/refresh") || strings.HasSuffix(path, "/submit") {
			c.Next()
			return
		}
		if _, err := c.Cookie(domain.AccessTokenCookieName); err != nil {
			c.Next()
			return
		}
		cookie, err := c.Cookie(domain.CSRFTokenCookieName)
		header := c.GetHeader("X-CSRF-Token")
		if err != nil || cookie == "" || header == "" || subtle.ConstantTimeCompare([]byte(cookie), []byte(header)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, domain.ErrorResponse{Message: "csrf validation failed"})
			return
		}
		c.Next()
	}
}

func setGinMode(appEnv string) {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "production", "prod", "release":
		gin.SetMode(gin.ReleaseMode)
	case "test", "testing":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

func expandLocalhostOrigins(origins []string) []string {
	formatted := make([]string, 0, len(origins))
	seen := make(map[string]struct{})

	appendUnique := func(origin string) {
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
		formatted = append(formatted, origin)
	}

	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}

		parsed, err := url.Parse(trimmed)
		if err != nil || parsed.Host == "" || parsed.User != nil ||
			(parsed.Scheme != "http" && parsed.Scheme != "https") {
			continue
		}
		appendUnique(trimmed)

		host, port, hasPort := strings.Cut(parsed.Host, ":")
		// Expand only exact local development hosts; substring matching would
		// trust attacker-controlled hosts such as localhost.example.com.
		if host == "localhost" {
			clone := *parsed
			if hasPort { clone.Host = "127.0.0.1:" + port } else { clone.Host = "127.0.0.1" }
			appendUnique(clone.String())
		}

		if host == "127.0.0.1" {
			clone := *parsed
			if hasPort { clone.Host = "localhost:" + port } else { clone.Host = "localhost" }
			appendUnique(clone.String())
		}
	}

	return formatted
}
