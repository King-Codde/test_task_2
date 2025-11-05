package service

import (
	"ip-camera-viewer/internal/domain"
	"ip-camera-viewer/internal/domain/model"
	"net"
	"net/url"
	"strings"
)

type ValidationService struct {
	normalizer       *domain.StreamNormalizer
	templateResolver *domain.TemplateResolver
}

func NewValidationService() *ValidationService {
	return &ValidationService{
		normalizer:       domain.NewStreamNormalizer(),
		templateResolver: domain.NewTemplateResolver(),
	}
}

func (vs *ValidationService) ValidateConnectionConfig(config *model.ConnectionConfig) *model.ValidationResult {
	result := model.NewValidationResult()

	if !IsValidIPv4(config.IP) {
		result.AddError("IP", "Некорректный IPv4-адрес")
	}

	if !IsValidPort(config.Port) {
		result.AddError("Port", "Порт должен быть числом 1–65535")
	}

	if strings.TrimSpace(config.Login) == "" {
		result.AddError("Login", "Логин не должен быть пустым")
	}

	if strings.TrimSpace(config.Password) == "" {
		result.AddError("Password", "Пароль не должен быть пустым")
	}

	if !IsValidRTSPURIWithPlaceholders(config.RTSPURI1) {
		result.AddError("RTSP URI #1", "RTSP-URI должен начинаться с rtsp:// и быть корректным URL")
	}

	if !IsValidRTSPURIWithPlaceholders(config.RTSPURI2) {
		result.AddError("RTSP URI #2", "RTSP-URI должен начинаться с rtsp:// и быть корректным URL")
	}

	if result.Valid {
		vs.checkStreamUniqueness(config, result)
	}

	return result
}

func (vs *ValidationService) checkStreamUniqueness(config *model.ConnectionConfig, result *model.ValidationResult) {
	uri1 := vs.templateResolver.Resolve(config.RTSPURI1, config)
	uri2 := vs.templateResolver.Resolve(config.RTSPURI2, config)

	identical, err := vs.normalizer.AreIdentical(uri1, uri2)
	if err != nil {
		result.AddWarning("Не удалось проверить уникальность RTSP URI: " + err.Error())
		return
	}

	if identical {
		result.AddError("RTSP URIs", "Оба RTSP-URI указывают на один и тот же поток после нормализации. Измените один из URI, чтобы избежать двойной загрузки.")
	}
}

func (vs *ValidationService) ValidateAndResolve(config *model.ConnectionConfig) (*model.ResolvedURIs, *model.ValidationResult) {
	result := vs.ValidateConnectionConfig(config)

	if !result.Valid {
		return nil, result
	}

	resolved := vs.templateResolver.ResolveAll(config)

	identical, _ := vs.normalizer.AreIdentical(resolved.URI1, resolved.URI2)
	resolved.AreIdentical = identical

	return resolved, result
}

func IsValidIPv4(ip string) bool {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return parsedIP.To4() != nil
}

func IsValidPort(port int) bool {
	return port >= 1 && port <= 65535
}

func IsValidRTSPURI(uri string) bool {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return false
	}

	if !strings.HasPrefix(strings.ToLower(uri), "rtsp://") {
		return false
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return false
	}

	if strings.ToLower(parsedURL.Scheme) != "rtsp" {
		return false
	}

	if parsedURL.Host == "" {
		return false
	}

	return true
}

func IsValidRTSPURIWithPlaceholders(uri string) bool {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return false
	}

	if !strings.HasPrefix(strings.ToLower(uri), "rtsp://") {
		return false
	}

	temp := uri
	temp = strings.ReplaceAll(temp, "{login}", "user")
	temp = strings.ReplaceAll(temp, "{password}", "pass")
	temp = strings.ReplaceAll(temp, "{ip}", "192.168.1.1")
	temp = strings.ReplaceAll(temp, "{port}", "554")

	return IsValidRTSPURI(temp)
}
