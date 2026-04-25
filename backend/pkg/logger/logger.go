package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"fayhub/pkg/config"
)

// Logger 日志接口
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
}

// zapLogger Zap日志实现
type zapLogger struct {
	logger *zap.Logger
}

var (
	globalLogger Logger
	loggerOnce   sync.Once
)

// InitLogger 初始化日志系统
func InitLogger(cfg *config.Config) error {
	var err error
	loggerOnce.Do(func() {
		globalLogger, err = newZapLogger(cfg)
		if err != nil {
			fmt.Printf("初始化日志系统失败: %v\n", err)
			os.Exit(1)
		}
	})
	return err
}

// GetLogger 获取全局日志实例
func GetLogger() Logger {
	if globalLogger == nil {
		// 如果未初始化，使用默认配置创建
		cfg := &config.Config{
			Logging: config.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
		}
		if err := InitLogger(cfg); err != nil {
			panic("日志系统初始化失败")
		}
	}
	return globalLogger
}

// newZapLogger 创建Zap日志实例
func newZapLogger(cfg *config.Config) (Logger, error) {
	var cores []zapcore.Core
	
	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	
	// 控制台输出
	if cfg.Logging.Output == "stdout" || cfg.Logging.Output == "both" {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), getLogLevel(cfg.Logging.Level))
		cores = append(cores, consoleCore)
	}
	
	// 文件输出
	if cfg.Logging.Output == "file" || cfg.Logging.Output == "both" {
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileWriter := getLogWriter(&cfg.Logging.File)
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), getLogLevel(cfg.Logging.Level))
		cores = append(cores, fileCore)
	}
	
	// 创建核心
	core := zapcore.NewTee(cores...)
	
	// 创建日志实例
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	
	return &zapLogger{logger: logger}, nil
}

// getLogLevel 获取日志级别
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// getLogWriter 获取日志写入器
func getLogWriter(fileCfg *config.LoggingFileConfig) zapcore.WriteSyncer {
	logFile := fileCfg.GetLogFilePath()
	if logFile == "" {
		return zapcore.AddSync(os.Stderr)
	}
	
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    fileCfg.MaxSize,    // MB
		MaxBackups: fileCfg.MaxBackups, // 备份数量
		MaxAge:     fileCfg.MaxAge,     // 天数
		Compress:   true,               // 压缩
	}
	
	return zapcore.AddSync(lumberJackLogger)
}

// 日志方法实现
func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Debug(msg, append(fields, getContextFields(ctx)...)...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Info(msg, append(fields, getContextFields(ctx)...)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Warn(msg, append(fields, getContextFields(ctx)...)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Error(msg, append(fields, getContextFields(ctx)...)...)
}

func (l *zapLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, append(fields, getContextFields(ctx)...)...)
}

func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

// getContextFields 从上下文中获取日志字段
func getContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	
	// 获取请求ID
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}
	
	// 获取租户ID
	if tenantID, ok := ctx.Value("tenant_id").(string); ok && tenantID != "" {
		fields = append(fields, zap.String("tenant_id", tenantID))
	}
	
	// 获取用户ID
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}
	
	return fields
}

// 便捷函数
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger().Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger().Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger().Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger().Error(ctx, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	GetLogger().Fatal(ctx, msg, fields...)
}

// Sync 同步日志缓冲区
func Sync() error {
	if logger, ok := globalLogger.(*zapLogger); ok {
		return logger.logger.Sync()
	}
	return nil
}

var sync sync.Once