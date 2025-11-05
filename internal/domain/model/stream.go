package model

import (
	"image"
	"time"
)

type StreamStatus int

const (
	StatusDisconnected StreamStatus = iota
	StatusConnecting
	StatusPlaying
	StatusError
	StatusReconnecting
)

func (s StreamStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "Отключено"
	case StatusConnecting:
		return "Подключение..."
	case StatusPlaying:
		return "Воспроизведение"
	case StatusError:
		return "Ошибка"
	case StatusReconnecting:
		return "Переподключение..."
	default:
		return "Неизвестно"
	}
}

type StreamInfo struct {
	Name             string
	Status           StreamStatus
	ErrorMessage     string
	Width            int
	Height           int
	FPS              float64
	Bitrate          int64
	LastFrameTime    time.Time
	ReconnectAttempt int
}

type FrameData struct {
	Image     image.Image
	Timestamp time.Time
}

type StreamStatusUpdate struct {
	StreamName string
	Status     StreamStatus
	Error      error
	Info       *StreamInfo
}
