package app

import "go.uber.org/zap"

// NewLogger returns a new instance of a SugaredLogger from the zap package.
//
// It uses the zap.NewProduction() function to create a logger with the default
// configuration for production environments. The logger is then wrapped in a
// SugaredLogger using the Sugar() method, which provides a more convenient
// API for logging.
//
// Returns:
// - *zap.SugaredLogger: A new SugaredLogger instance.
func NewLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	return logger.Sugar()
}
