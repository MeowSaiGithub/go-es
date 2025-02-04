package main

import (
	"context"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go-es/config"
	"go-es/internal/esc"
	"go-es/internal/response"
	"go-es/logger"
	"go-es/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type app struct {
	cfg *config.Config
	srv server.Server
	ec  *elasticsearch.Client
}

func main() {

	// Define a command-line flag for the config file
	pflag.String("config", "config.yaml", "Path to the configuration file")
	pflag.Parse() // Parse flags

	// Bind flags with Viper
	viper.BindPFlags(pflag.CommandLine)

	var err error
	a := app{}

	a.cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("configuration error")
	}
	response.Init(a.cfg.DetailError)

	logger.InitializeLogger(a.cfg.LogLevel)

	if a.ec, err = esc.NewClient(a.cfg.ElasticSearch); err != nil {
		log.Fatal().Err(err).Msg("failed to create an elastic client")
	}

	a.srv = server.NewServer(a.cfg.Server, a.ec)

	// Set up signal handling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Start server in a goroutine
	go func() {
		log.Info().Str("base_path", "/v1").Str("address", ":8080").Msg("server started")
		if err := a.srv.Run(); err != nil {
			log.Error().Err(err).Msg("server error occurred")
		}
	}()

	// Call Shutdown when a signal is received
	a.srv.Shutdown(ctx, cancel, sig)
	log.Info().Msgf("server exit properly")

}
