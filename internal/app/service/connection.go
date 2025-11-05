package service

import (
	"context"
	"ip-camera-viewer/internal/domain"
	"ip-camera-viewer/internal/domain/model"
)

type ConnectionService struct {
	validationService *ValidationService
	streamManager     *StreamManager
	logger            *LoggerService
	templateResolver  *domain.TemplateResolver
}

func NewConnectionService(logger *LoggerService, streamManager *StreamManager) *ConnectionService {
	return &ConnectionService{
		validationService: NewValidationService(),
		streamManager:     streamManager,
		logger:            logger,
		templateResolver:  domain.NewTemplateResolver(),
	}
}

func (cs *ConnectionService) ValidateConfig(config *model.ConnectionConfig) (*model.ResolvedURIs, *model.ValidationResult) {
	return cs.validationService.ValidateAndResolve(config)
}

func (cs *ConnectionService) Connect(ctx context.Context, config *model.ConnectionConfig) (
	highFrames <-chan *model.FrameData,
	lowFrames <-chan *model.FrameData,
	err error,
) {
	resolved, result := cs.validationService.ValidateAndResolve(config)
	if !result.Valid {
		cs.logger.Error("Ошибка валидации конфигурации", nil)
		return nil, nil, &model.AppError{
			Type:        model.ErrorTypeValidation,
			Message:     "Валидация не пройдена",
			UserMessage: result.GetErrorMessage(),
		}
	}

	cs.logger.Info("Начало подключения к камере %s:%d", config.IP, config.Port)

	highConfig := model.NewStreamConfig("High", resolved.URI1, config.Login, config.Password)
	lowConfig := model.NewStreamConfig("Low", resolved.URI2, config.Login, config.Password)

	highChan, err := cs.streamManager.StartStream(ctx, highConfig)
	if err != nil {
		cs.logger.Error("Ошибка запуска High потока", err)
		return nil, nil, err
	}

	lowChan, err := cs.streamManager.StartStream(ctx, lowConfig)
	if err != nil {
		cs.logger.Error("Ошибка запуска Low потока", err)
		cs.streamManager.StopStream("High")
		return nil, nil, err
	}

	cs.logger.Info("Оба потока успешно запущены")

	return highChan, lowChan, nil
}

func (cs *ConnectionService) Disconnect() {
	cs.logger.Info("Отключение от всех потоков")
	cs.streamManager.StopAllStreams()
}

func (cs *ConnectionService) GetStatusChannel() <-chan *model.StreamStatusUpdate {
	return cs.streamManager.GetStatusChannel()
}
