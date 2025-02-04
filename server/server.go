package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	documentsHandler "go-es/internal/service/document/handler"
	indicesHandler "go-es/internal/service/index/handler"
	middlewares "go-es/server/middleware"
	"log"
	"net/http"
	"os"
	"time"
)

// Config struct with validation tags
type Config struct {
	BasePath string     `mapstructure:"base_path" validate:"required"`            // Must not be empty
	Port     int        `mapstructure:"port" validate:"required,min=1,max=65535"` // Must be between 1 and 65535
	Secret   string     `mapstructure:"api_secret"`
	Cors     CorsConfig `mapstructure:"cors"`
}

// CorsConfig defines the configuration for CORS (Cross-Origin Resource Sharing).
// It specifies the allowed origins, methods, headers, and whether credentials are included.
type CorsConfig struct {
	CorsEnable  bool     `mapstructure:"cors_enable"`                                                                    // Indicates if CORS is enabled
	Origins     []string `mapstructure:"origins" validate:"omitempty,dive,url"`                                          // List of allowed origins
	Methods     []string `mapstructure:"cors_methods" validate:"omitempty,dive,oneof=GET POST PUT DELETE PATCH OPTIONS"` // List of allowed HTTP methods
	Headers     []string `mapstructure:"cors_headers" validate:"omitempty,dive"`                                         // List of allowed headers
	Credentials bool     `mapstructure:"cors_credentials"`                                                               // Indicates if credentials are allowed
}

// Server represents an interface for a server with methods to run and shutdown.
type Server interface {
	// Run starts the server.
	Run() error

	// Shutdown gracefully shuts down the server with context, cancel function, and signal channel.
	Shutdown(ctx context.Context, cancel context.CancelFunc, sig chan os.Signal) error
}

type server struct {
	cfg *Config
	srv *http.Server
	esc *elasticsearch.Client
}

// NewServer creates a new server instance with the given configuration and elasticsearch client.
// It will panic if the configuration is invalid.
func NewServer(cfg *Config, esc *elasticsearch.Client) Server {
	srv := &server{cfg, nil, esc}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Logger())

	if srv.cfg.Cors.CorsEnable {
		corsCfg := cors.Config{
			AllowOrigins:     srv.cfg.Cors.Origins,
			AllowMethods:     srv.cfg.Cors.Methods,
			AllowHeaders:     srv.cfg.Cors.Headers,
			AllowCredentials: srv.cfg.Cors.Credentials,
		}
		r.Use(cors.New(corsCfg))
	}

	if srv.cfg.Secret != "" {
		r.Use(middlewares.Auth(srv.cfg.Secret))
	}

	basePath := srv.cfg.BasePath

	r.GET(basePath, func(c *gin.Context) {
		c.JSON(200, "Hello, World!")
	})
	// Group routes for indices/aliases management
	indices := r.Group(basePath + "indices")
	{
		indices.POST("/", indicesHandler.CreateIndex(srv.esc))         // Create index
		indices.DELETE("/:alias", indicesHandler.DeleteIndex(srv.esc)) // Delete index
		indices.GET("/", indicesHandler.ListIndices(srv.esc))          // List all indicesHandler
		indices.GET("/:alias/exists", indicesHandler.Exists(srv.esc))  // Check if index exists
		indices.GET("/:alias/info", indicesHandler.GetIndex(srv.esc))  // Get index information
		indices.PUT("/:alias", indicesHandler.UpdateIndex(srv.esc))
	}

	// Group routes for documents management
	documents := r.Group(basePath + "documents")
	{
		documents.POST("/:alias/add", documentsHandler.AddData(srv.esc))            // Add documentsHandler (bulk or single)
		documents.POST("/:alias/search", documentsHandler.Search(srv.esc))          // Search documentsHandler
		documents.POST("/:alias/suggest", documentsHandler.AutoComplete(srv.esc))   // Get document suggestions
		documents.POST("/:alias/export", documentsHandler.ExportDocuments(srv.esc)) // Export documents
		documents.POST("/:alias/import", documentsHandler.ImportDocuments(srv.esc)) // Import documents
		documents.DELETE("/:alias/:id", documentsHandler.DeleteDocument(srv.esc))   // Delete document by ID
		documents.PUT("/:alias/:id", documentsHandler.UpdateDocument(srv.esc))      // Update document by ID
		documents.GET("/:alias", documentsHandler.ListAllDocuments(srv.esc))        // List all documentsHandler in index
		documents.GET("/:alias/:id", documentsHandler.GetDocumentByID(srv.esc))     // Get document by ID

	}

	srv.srv = &http.Server{
		Addr:              fmt.Sprintf(":%d", srv.cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 60 * time.Second,
		WriteTimeout:      60 * time.Second,
	}
	return srv
}

// Run starts the server.
func (s *server) Run() error {
	return s.srv.ListenAndServe()
}

// Shutdown gracefully shuts down the server with context, cancel function, and signal channel.
// It will return an error if the shutdown fails.
func (s *server) Shutdown(ctx context.Context, cancel context.CancelFunc, sig chan os.Signal) error {
	if s == nil || s.srv == nil {
		return fmt.Errorf("server or http.Server instance is nil")
	}

	<-sig
	defer cancel()

	go func() {
		<-ctx.Done()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Printf("graceful shutdown timed out.. forcing exit")
		}
	}()

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	return nil
}
