package service

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
)

type LoggerService struct {
	logger *slog.Logger
}

func NewLoggerService() *LoggerService {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &LoggerService{
		logger: slog.New(handler),
	}
}

func (ls *LoggerService) MaskSensitiveData(message string) string {
	masked := message

	rtspPattern := regexp.MustCompile(`(rtsp://[^:]+:)([^@]+)(@)`)
	masked = rtspPattern.ReplaceAllString(masked, "${1}***${3}")

	passwordPattern := regexp.MustCompile(`(?i)(password|pwd|pass)\s*[:=]\s*["']?([^"'\s,&]+)["']?`)
	masked = passwordPattern.ReplaceAllString(masked, "${1}=***")

	jsonPasswordPattern := regexp.MustCompile(`(?i)"(password|pwd|pass)"\s*:\s*"([^"]+)"`)
	masked = jsonPasswordPattern.ReplaceAllString(masked, `"${1}":"***"`)

	return masked
}

func (ls *LoggerService) Info(message string, args ...any) {
	masked := ls.MaskSensitiveData(fmt.Sprintf(message, args...))
	ls.logger.Info(masked)
}

func (ls *LoggerService) Error(message string, err error, args ...any) {
	masked := ls.MaskSensitiveData(fmt.Sprintf(message, args...))
	if err != nil {
		maskedErr := ls.MaskSensitiveData(err.Error())
		ls.logger.Error(masked, "error", maskedErr)
	} else {
		ls.logger.Error(masked)
	}
}

func (ls *LoggerService) Warn(message string, args ...any) {
	masked := ls.MaskSensitiveData(fmt.Sprintf(message, args...))
	ls.logger.Warn(masked)
}

func (ls *LoggerService) Debug(message string, args ...any) {
	masked := ls.MaskSensitiveData(fmt.Sprintf(message, args...))
	ls.logger.Debug(masked)
}

func (ls *LoggerService) FormatConnectionAttempt(streamName, uri string) string {
	return fmt.Sprintf("Попытка подключения к потоку %s: %s", streamName, ls.MaskSensitiveData(uri))
}
