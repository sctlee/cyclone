package controller

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// InitLogger inits logging
func InitLogger(logging *LoggingConfig) {
	log.WithField("level", logging.Level).Info("Setting log level")

	switch strings.ToLower(logging.Level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Fatalf("Unknown level: %s", logging.Level)
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		ForceColors:   true,
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
}
