package logging

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

type AppLogger struct {
	*zap.Logger
	config *AppLoggerConfig
}

type AppLoggerConfig struct {
	FilePath string
	Level    zapcore.Level
}

func NewAppLogger(logger *zap.Logger, config *AppLoggerConfig) *AppLogger {
	logLevel := zapcore.DebugLevel
	filePath := "logs/app.log"

	if config != nil {
		if config.FilePath != "" {
			filePath = config.FilePath
		}

		logLevel = config.Level
	}

	if logger == nil {
		cfg := zap.NewProductionEncoderConfig()
		cfg.EncodeTime = zapcore.ISO8601TimeEncoder

		fileEncoder := zapcore.NewJSONEncoder(cfg)
		consoleEncoder := zapcore.NewConsoleEncoder(cfg)

		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		})

		core := zapcore.NewTee(
			zapcore.NewCore(fileEncoder, writer, logLevel),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel),
		)
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Named("comff-stores")
	}

	return &AppLogger{
		Logger: logger,
		config: config,
	}
}

func NewTestAppLogger(dir string) *AppLogger {
	logFilePath := "logs/app.log"

	if dir != "" {
		logFilePath = filepath.Join(dir, "logs/app.log")
	}

	logCfg := AppLoggerConfig{
		Level:    zapcore.DebugLevel,
		FilePath: logFilePath,
	}

	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(cfg)
	consoleEncoder := zapcore.NewConsoleEncoder(cfg)

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logCfg.FilePath,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, logCfg.Level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logCfg.Level),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Named("comff-stores-test")
	return &AppLogger{
		Logger: logger,
		config: &logCfg,
	}
}

func NewZapLogger() *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return Logger
}
