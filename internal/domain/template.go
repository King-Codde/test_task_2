package domain

import (
	"fmt"
	"ip-camera-viewer/internal/domain/model"
	"strings"
)

type TemplateResolver struct{}

func NewTemplateResolver() *TemplateResolver {
	return &TemplateResolver{}
}

func (tr *TemplateResolver) Resolve(template string, config *model.ConnectionConfig) string {
	result := template

	result = strings.ReplaceAll(result, "{login}", config.Login)
	result = strings.ReplaceAll(result, "{password}", config.Password)
	result = strings.ReplaceAll(result, "{ip}", config.IP)
	result = strings.ReplaceAll(result, "{port}", fmt.Sprintf("%d", config.Port))

	return result
}

func (tr *TemplateResolver) ResolveAll(config *model.ConnectionConfig) *model.ResolvedURIs {
	uri1 := tr.Resolve(config.RTSPURI1, config)
	uri2 := tr.Resolve(config.RTSPURI2, config)

	return &model.ResolvedURIs{
		URI1:       uri1,
		URI2:       uri2,
		URI1Masked: tr.MaskPassword(uri1),
		URI2Masked: tr.MaskPassword(uri2),
	}
}

func (tr *TemplateResolver) MaskPassword(uri string) string {

	if !strings.Contains(uri, "://") {
		return uri
	}

	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return uri
	}

	scheme := parts[0]
	rest := parts[1]

	atIndex := strings.Index(rest, "@")
	if atIndex == -1 {
		return uri
	}

	authPart := rest[:atIndex]
	hostPart := rest[atIndex:]

	colonIndex := strings.Index(authPart, ":")
	if colonIndex == -1 {
		return uri
	}

	login := authPart[:colonIndex]

	return fmt.Sprintf("%s://%s:***%s", scheme, login, hostPart)
}
