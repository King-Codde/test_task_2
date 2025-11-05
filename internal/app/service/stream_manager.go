package service

import (
	"context"
	"fmt"
	"ip-camera-viewer/internal/domain/model"
	"ip-camera-viewer/internal/infrastructure/rtsp"
	"sync"
)

type StreamManager struct {
	logger        *LoggerService
	streams       map[string]*StreamController
	statusChannel chan *model.StreamStatusUpdate
	mu            sync.RWMutex
	cancelFuncs   map[string]context.CancelFunc
}

type StreamController struct {
	Config       *model.StreamConfig
	Client       *rtsp.Client
	FrameChannel chan *model.FrameData
	Status       model.StreamStatus
	Info         *model.StreamInfo
}

func NewStreamManager(logger *LoggerService) *StreamManager {
	return &StreamManager{
		logger:        logger,
		streams:       make(map[string]*StreamController),
		statusChannel: make(chan *model.StreamStatusUpdate, 100),
		cancelFuncs:   make(map[string]context.CancelFunc),
	}
}

func (sm *StreamManager) StartStream(ctx context.Context, config *model.StreamConfig) (<-chan *model.FrameData, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	streamCtx, cancel := context.WithCancel(ctx)
	sm.cancelFuncs[config.Name] = cancel

	frameChannel := make(chan *model.FrameData, 30)

	client := rtsp.NewClient(config)

	controller := &StreamController{
		Config:       config,
		Client:       client,
		FrameChannel: frameChannel,
		Status:       model.StatusConnecting,
		Info: &model.StreamInfo{
			Name:   config.Name,
			Status: model.StatusConnecting,
		},
	}

	sm.streams[config.Name] = controller

	sm.sendStatus(config.Name, model.StatusConnecting, nil)

	go sm.handleStream(streamCtx, controller)

	return frameChannel, nil
}

func (sm *StreamManager) handleStream(ctx context.Context, controller *StreamController) {
	defer close(controller.FrameChannel)

	err := controller.Client.Connect(ctx)
	if err != nil {
		sm.logger.Error("Ошибка подключения к потоку %s", err, controller.Config.Name)
		sm.sendStatus(controller.Config.Name, model.StatusError, err)
		return
	}

	sm.logger.Info("Успешно подключен к потоку %s", controller.Config.Name)
	sm.sendStatus(controller.Config.Name, model.StatusPlaying, nil)

	controller.Status = model.StatusPlaying

	err = controller.Client.StartStreaming(ctx, controller.FrameChannel)
	if err != nil {
		sm.logger.Error("Ошибка стриминга потока %s", err, controller.Config.Name)
		sm.sendStatus(controller.Config.Name, model.StatusError, err)
		return
	}

	<-ctx.Done()

	controller.Client.Close()
	sm.sendStatus(controller.Config.Name, model.StatusDisconnected, nil)
	sm.logger.Info("Поток %s отключен", controller.Config.Name)
}

func (sm *StreamManager) StopStream(streamName string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cancel, exists := sm.cancelFuncs[streamName]
	if !exists {
		return fmt.Errorf("поток %s не найден", streamName)
	}

	cancel()
	delete(sm.cancelFuncs, streamName)
	delete(sm.streams, streamName)

	sm.logger.Info("Остановка потока %s", streamName)
	return nil
}

func (sm *StreamManager) StopAllStreams() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for name, cancel := range sm.cancelFuncs {
		cancel()
		sm.logger.Info("Остановка потока %s", name)
	}

	sm.cancelFuncs = make(map[string]context.CancelFunc)
	sm.streams = make(map[string]*StreamController)
}

func (sm *StreamManager) GetStatusChannel() <-chan *model.StreamStatusUpdate {
	return sm.statusChannel
}

func (sm *StreamManager) sendStatus(streamName string, status model.StreamStatus, err error) {
	update := &model.StreamStatusUpdate{
		StreamName: streamName,
		Status:     status,
		Error:      err,
		Info: &model.StreamInfo{
			Name:   streamName,
			Status: status,
		},
	}

	select {
	case sm.statusChannel <- update:
	default:
	}
}
