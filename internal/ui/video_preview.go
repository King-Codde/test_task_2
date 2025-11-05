package ui

import (
	"context"
	"image"
	"ip-camera-viewer/internal/domain/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type VideoPreviewWidget struct {
	widget.BaseWidget
	streamName   string
	image        *canvas.Image
	statusLabel  *widget.Label
	frameChannel <-chan *model.FrameData
	cancelFunc   context.CancelFunc
	container    *fyne.Container
}

func NewVideoPreviewWidget(streamName string) *VideoPreviewWidget {
	w := &VideoPreviewWidget{
		streamName:  streamName,
		statusLabel: widget.NewLabel("Статус: Отключено"),
	}

	placeholder := image.NewRGBA(image.Rect(0, 0, 640, 360))
	w.image = canvas.NewImageFromImage(placeholder)
	w.image.FillMode = canvas.ImageFillContain
	w.image.SetMinSize(fyne.NewSize(640, 360))

	w.container = container.NewBorder(
		nil,
		w.statusLabel,
		nil,
		nil,
		w.image,
	)

	w.ExtendBaseWidget(w)
	return w
}

func (w *VideoPreviewWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.container)
}

func (w *VideoPreviewWidget) StartStreaming(ctx context.Context, frameChannel <-chan *model.FrameData) {
	streamCtx, cancel := context.WithCancel(ctx)
	w.cancelFunc = cancel
	w.frameChannel = frameChannel

	go w.updateLoop(streamCtx)
}

func (w *VideoPreviewWidget) StopStreaming() {
	if w.cancelFunc != nil {
		w.cancelFunc()
		w.cancelFunc = nil
	}
}

func (w *VideoPreviewWidget) updateLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case frame, ok := <-w.frameChannel:
			if !ok {
				return
			}
			if frame != nil && frame.Image != nil {
				w.image.Image = frame.Image
				w.image.Refresh()
			}
		}
	}
}

func (w *VideoPreviewWidget) UpdateStatus(status model.StreamStatus, info *model.StreamInfo) {
	statusText := status.String()
	if info != nil {
		if status == model.StatusError && info.ErrorMessage != "" {
			statusText = status.String() + ": " + info.ErrorMessage
		}
	}
	w.statusLabel.SetText("Статус: " + statusText)
}
