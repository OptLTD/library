package support

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	logger     *slog.Logger
	loggerOnce sync.Once
)

// initLogger 初始化 slog Logger（单例模式）
func initLogger() {
	loggerOnce.Do(func() {
		// 从环境变量读取日志级别，默认 INFO
		levelStr := os.Getenv("LOG_LEVEL")
		levelStr = If(levelStr, levelStr, "INFO")
		level := parseLogLevel(strings.ToUpper(levelStr))

		opts := &slog.HandlerOptions{Level: level}
		logger = slog.New(slog.NewJSONHandler(os.Stderr, opts))
	})
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// GetLogger 获取 Logger 实例
func GetLogger() *slog.Logger {
	initLogger()
	return logger
}

// SetLogLevel 设置日志级别（运行时动态修改）
func SetLogLevel(level string) {
	initLogger()
	newLevel := parseLogLevel(strings.ToUpper(level))
	opts := &slog.HandlerOptions{Level: newLevel}
	logger = slog.New(slog.NewJSONHandler(os.Stderr, opts))
}

// logWithLevel 内部函数：按级别记录日志
func logWithLevel(logid string, level slog.Level, args ...interface{}) {
	l := GetLogger()
	msg := fmt.Sprint(args...)
	logid = If(logid, logid, "unknown")
	l.Log(context.Background(), level, msg, "logid", logid)
}

// logfWithLevel 内部函数：格式化并按级别记录日志
func logfWithLevel(logid string, level slog.Level, format string, args ...interface{}) {
	l := GetLogger()
	msg := fmt.Sprintf(format, args...)
	logid = If(logid, logid, "unknown")
	l.Log(context.Background(), level, msg, "logid", logid)
}

func logKV(logid string, level slog.Level, msg string, attrs ...any) {
	l := GetLogger()
	logid = If(logid, logid, "unknown")
	base := []any{"logid", logid}
	base = append(base, attrs...)
	l.Log(context.Background(), level, msg, base...)
}

// Log 记录日志（默认 INFO 级别，兼容旧代码）
func Log(logid string, args ...interface{}) {
	LogInfo(logid, args...)
}

// Logf 格式化日志（默认 INFO 级别，兼容旧代码）
func Logf(logid string, format string, args ...interface{}) {
	LogInfof(logid, format, args...)
}

// LogDebug 记录 DEBUG 级别日志（详细调试信息）
func LogDebug(logid string, args ...interface{}) {
	logWithLevel(logid, slog.LevelDebug, args...)
}

// LogDebugf 格式化 DEBUG 级别日志
func LogDebugf(logid string, format string, args ...interface{}) {
	logfWithLevel(logid, slog.LevelDebug, format, args...)
}

// LogInfo 记录 INFO 级别日志（一般信息）
func LogInfo(logid string, args ...interface{}) {
	logWithLevel(logid, slog.LevelInfo, args...)
}

// LogInfof 格式化 INFO 级别日志
func LogInfof(logid string, format string, args ...interface{}) {
	logfWithLevel(logid, slog.LevelInfo, format, args...)
}

// LogWarn 记录 WARN 级别日志（警告信息）
func LogWarn(logid string, args ...interface{}) {
	logWithLevel(logid, slog.LevelWarn, args...)
}

// LogWarnf 格式化 WARN 级别日志
func LogWarnf(logid string, format string, args ...interface{}) {
	logfWithLevel(logid, slog.LevelWarn, format, args...)
}

// LogError 记录 ERROR 级别日志（错误信息）
func LogError(logid string, args ...interface{}) {
	logWithLevel(logid, slog.LevelError, args...)
}

// LogErrorf 格式化 ERROR 级别日志
func LogErrorf(logid string, format string, args ...interface{}) {
	logfWithLevel(logid, slog.LevelError, format, args...)
}

// LogDebugKV 结构化 DEBUG 日志
func LogDebugKV(logid string, msg string, attrs ...any) {
	logKV(logid, slog.LevelDebug, msg, attrs...)
}

// LogInfoKV 结构化 INFO 日志
func LogInfoKV(logid string, msg string, attrs ...any) {
	logKV(logid, slog.LevelInfo, msg, attrs...)
}

// LogWarnKV 结构化 WARN 日志
func LogWarnKV(logid string, msg string, attrs ...any) {
	logKV(logid, slog.LevelWarn, msg, attrs...)
}

// LogErrorKV 结构化 ERROR 日志
func LogErrorKV(logid string, msg string, attrs ...any) {
	logKV(logid, slog.LevelError, msg, attrs...)
}
