package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// New constructs a zap.Logger using the supplied log level.  The logger
// writes JSON formatted logs to stdout which makes it easy to ingest
// into log aggregation systems.  Supported levels are "debug", "info",
// "warn" and "error".  Any unknown level falls back to "info".
func New(level string) *zap.Logger {
    var lvl zapcore.Level
    switch level {
    case "debug":
        lvl = zapcore.DebugLevel
    case "info":
        lvl = zapcore.InfoLevel
    case "warn":
        lvl = zapcore.WarnLevel
    case "error":
        lvl = zapcore.ErrorLevel
    default:
        lvl = zapcore.InfoLevel
    }
    cfg := zap.NewProductionConfig()
    cfg.Level = zap.NewAtomicLevelAt(lvl)
    cfg.Encoding = "json"
    // Use console friendly time format
    cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    logger, err := cfg.Build()
    if err != nil {
        panic(err)
    }
    return logger
}