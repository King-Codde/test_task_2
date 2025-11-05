package ui

import (
	"context"
	"ip-camera-viewer/internal/app/service"
	"ip-camera-viewer/internal/domain/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type MainWindow struct {
	app               fyne.App
	window            fyne.Window
	connectionService *service.ConnectionService
	configService     *service.ConfigurationService
	logger            *service.LoggerService

	connectionForm *ConnectionForm
	highPreview    *VideoPreviewWidget
	lowPreview     *VideoPreviewWidget
	logPanel       *LogPanel

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewMainWindow() *MainWindow {
	logger := service.NewLoggerService()
	streamManager := service.NewStreamManager(logger)
	connectionService := service.NewConnectionService(logger, streamManager)
	configService := service.NewConfigurationService(logger)

	myApp := app.New()
	win := myApp.NewWindow("IP Camera Viewer")

	ctx, cancel := context.WithCancel(context.Background())

	mw := &MainWindow{
		app:               myApp,
		window:            win,
		connectionService: connectionService,
		configService:     configService,
		logger:            logger,
		connectionForm:    NewConnectionForm(),
		highPreview:       NewVideoPreviewWidget("High"),
		lowPreview:        NewVideoPreviewWidget("Low"),
		logPanel:          NewLogPanel(),
		ctx:               ctx,
		cancelFunc:        cancel,
	}

	mw.setupUI()
	mw.setupHandlers()
	mw.loadSavedConfig()
	mw.startStatusMonitoring()

	return mw
}

func (mw *MainWindow) setupUI() {
	topSection := mw.connectionForm

	videoSection := container.NewGridWithColumns(2,
		container.NewBorder(
			widget.NewLabel("Preview High"),
			nil, nil, nil,
			mw.highPreview,
		),
		container.NewBorder(
			widget.NewLabel("Preview Low"),
			nil, nil, nil,
			mw.lowPreview,
		),
	)

	logSection := container.NewBorder(
		widget.NewLabel("Логи и сообщения:"),
		nil,
		nil, container.NewBorder(
			mw.connectionForm.connectButton,
			mw.connectionForm.disconnectButton,
			nil, nil,
			nil,
		),
		mw.logPanel.GetWidget(),
	)

	mainContainer := container.NewBorder(
		topSection,
		logSection,
		nil,
		nil,
		videoSection,
	)

	mw.window.SetContent(mainContainer)
	mw.window.Resize(fyne.NewSize(1280, 800))
}

func (mw *MainWindow) setupHandlers() {
	mw.connectionForm.SetOnCheck(func(config *model.ConnectionConfig) {
		mw.handleCheck(config)
	})

	mw.connectionForm.SetOnConnect(func(config *model.ConnectionConfig) {
		mw.handleConnect(config)
	})

	mw.connectionForm.SetOnDisconnect(func() {
		mw.handleDisconnect()
	})

	mw.window.SetOnClosed(func() {
		mw.cancelFunc()
		mw.connectionService.Disconnect()
	})
}

func (mw *MainWindow) handleCheck(config *model.ConnectionConfig) {
	mw.logPanel.AddLog("Проверка конфигурации...")

	resolved, result := mw.connectionService.ValidateConfig(config)

	if !result.Valid {
		mw.logPanel.AddLog("Ошибки валидации:")
		for _, err := range result.Errors {
			mw.logPanel.AddLog("  - " + err.Field + ": " + err.Message)
		}
		dialog.ShowError(&validationError{result.GetErrorMessage()}, mw.window)
		return
	}

	mw.logPanel.AddLog("Валидация пройдена")
	mw.logPanel.AddLog("RTSP URI #1: " + resolved.URI1Masked)
	mw.logPanel.AddLog("RTSP URI #2: " + resolved.URI2Masked)
	mw.configService.SaveConfig(config)
	mw.connectionForm.rtspURI1Entry.SetText(resolved.URI1Masked)
	mw.connectionForm.rtspURI2Entry.SetText(resolved.URI2Masked)

	mw.connectionForm.ConnectionPermission(resolved.AreIdentical)

	if resolved.AreIdentical {
		mw.logPanel.AddLog("Оба RTSP-URI идентичны")
	}

	for _, warn := range result.Warnings {
		mw.logPanel.AddLog(warn)
	}
}

func (mw *MainWindow) handleConnect(config *model.ConnectionConfig) {
	mw.logPanel.AddLog("Подключение к камере")

	_, result := mw.connectionService.ValidateConfig(config)
	if !result.Valid {
		mw.logPanel.AddLog("Ошибка валидации")
		dialog.ShowError(&validationError{result.GetErrorMessage()}, mw.window)
		return
	}

	highChan, lowChan, err := mw.connectionService.Connect(mw.ctx, config)
	if err != nil {
		mw.logPanel.AddLog("Ошибка подключения: " + err.Error())
		dialog.ShowError(err, mw.window)
		return
	}

	mw.highPreview.StartStreaming(mw.ctx, highChan)
	mw.lowPreview.StartStreaming(mw.ctx, lowChan)

	mw.connectionForm.SetConnected(true)
	mw.logPanel.AddLog("Подключение установлено")

	mw.configService.SaveConfig(config)
}

func (mw *MainWindow) handleDisconnect() {
	mw.logPanel.AddLog("Отключение")

	mw.highPreview.StopStreaming()
	mw.lowPreview.StopStreaming()
	mw.connectionService.Disconnect()

	mw.connectionForm.SetConnected(false)
	mw.logPanel.AddLog("Отключено")
}

func (mw *MainWindow) startStatusMonitoring() {
	go func() {
		statusChan := mw.connectionService.GetStatusChannel()
		for {
			select {
			case <-mw.ctx.Done():
				return
			case update := <-statusChan:
				mw.handleStatusUpdate(update)
			}
		}
	}()
}

func (mw *MainWindow) handleStatusUpdate(update *model.StreamStatusUpdate) {
	msg := update.StreamName + ": " + update.Status.String()
	if update.Error != nil {
		msg += " - " + update.Error.Error()
	}
	mw.logPanel.AddLog(msg)

	switch update.StreamName {
	case "High":
		mw.highPreview.UpdateStatus(update.Status, update.Info)
	case "Low":
		mw.lowPreview.UpdateStatus(update.Status, update.Info)
	}
}

func (mw *MainWindow) loadSavedConfig() {
	saved, err := mw.configService.LoadConfig()
	if err != nil {
		mw.logger.Error("Ошибка загрузки конфигурации", err)
		return
	}

	if saved != nil {
		config := &model.ConnectionConfig{
			IP:       saved.IP,
			Port:     saved.Port,
			Login:    saved.Login,
			RTSPURI1: saved.RTSPURI1,
			RTSPURI2: saved.RTSPURI2,
		}
		mw.connectionForm.LoadConfig(config)
		mw.logPanel.AddLog("Конфигурация загружена")
	}
}

func (mw *MainWindow) ShowAndRun() {
	mw.window.ShowAndRun()
}

type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}
