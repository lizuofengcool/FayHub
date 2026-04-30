package logger

import (
	"context"
	"fayhub/pkg/utils"
	"fmt"
	"os"
	"sync"

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
	fallbackOnce sync.Once
)

func InitLogger(cfg *config.Config) error {
	var err error
	loggerOnce.Do(func() {
		globalLogger, err = newZapLogger(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "初始化日志系统失败: %v，将使用降级日志\n", err)
		}
	})
	return err
}

func GetLogger() Logger {
	if globalLogger == nil {
		cfg := &config.Config{
			Logging: config.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
		}
		if err := InitLogger(cfg); err != nil {
			fallbackOnce.Do(func() {
				fmt.Fprintf(os.Stderr, "日志系统降级初始化失败: %v，使用stderr输出\n", err)
			})
			return &fallbackLogger{}
		}
	}
	return globalLogger
}

type fallbackLogger struct{}

func (l *fallbackLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", msg)
}

func (l *fallbackLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fmt.Fprintf(os.Stderr, "[INFO] %s\n", msg)
}

func (l *fallbackLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fmt.Fprintf(os.Stderr, "[WARN] %s\n", msg)
}

func (l *fallbackLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
}

func (l *fallbackLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fmt.Fprintf(os.Stderr, "[FATAL] %s\n", msg)
}

func (l *fallbackLogger) With(fields ...zap.Field) Logger {
	return l
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
	core = newTenantLogCore(core)

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

	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if tenantID, ok := utils.GetTenantIDFromCtx(ctx); ok && tenantID > 0 {
		fields = append(fields, zap.Uint("tenant_id", tenantID))
	}

	if userID, ok := ctx.Value("user_id").(uint); ok && userID > 0 {
		fields = append(fields, zap.Uint("user_id", userID))
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
