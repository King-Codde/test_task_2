package ui

import (
	"ip-camera-viewer/internal/domain/model"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ConnectionForm struct {
	widget.BaseWidget

	ipEntry       *widget.Entry
	portEntry     *widget.Entry
	loginEntry    *widget.Entry
	passwordEntry *widget.Entry
	rtspURI1Entry *widget.Entry
	rtspURI2Entry *widget.Entry

	checkButton      *widget.Button
	connectButton    *widget.Button
	disconnectButton *widget.Button

	onCheck      func(*model.ConnectionConfig)
	onConnect    func(*model.ConnectionConfig)
	onDisconnect func()

	container *fyne.Container
}

func NewConnectionForm() *ConnectionForm {
	f := &ConnectionForm{
		ipEntry:       widget.NewEntry(),
		portEntry:     widget.NewEntry(),
		loginEntry:    widget.NewEntry(),
		passwordEntry: widget.NewPasswordEntry(),
		rtspURI1Entry: widget.NewEntry(),
		rtspURI2Entry: widget.NewEntry(),
	}

	f.checkButton = widget.NewButton("Проверить", func() {
		if f.onCheck != nil {
			config := f.GetConfig()
			f.onCheck(config)
		}
	})

	f.connectButton = widget.NewButton("Подключиться", func() {
		if f.onConnect != nil {
			config := f.GetConfig()
			f.onConnect(config)
		}
	})

	f.disconnectButton = widget.NewButton("Отключиться", func() {
		if f.onDisconnect != nil {
			f.onDisconnect()
		}
	})

	f.disconnectButton.Disable()
	f.connectButton.Disable()

	f.buildUI()
	f.ExtendBaseWidget(f)
	return f
}

func (cf *ConnectionForm) buildUI() {
	title := widget.NewLabel("IP Camera Viewer")
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter

	ip := widget.NewLabel("IP:")
	port := widget.NewLabel("Port:")
	login := widget.NewLabel("Login:")
	password := widget.NewLabel("Password:")

	ipContainer := container.NewBorder(
		nil, nil,
		container.NewVBox(ip),
		nil,
		container.NewVBox(cf.ipEntry),
	)

	portContainer := container.NewBorder(
		nil, nil,
		container.NewVBox(port),
		nil,
		container.NewVBox(cf.portEntry),
	)

	loginContainer := container.NewBorder(
		nil, nil,
		container.NewVBox(login),
		nil,
		container.NewVBox(cf.loginEntry),
	)

	passwordContainer := container.NewBorder(
		nil, nil,
		container.NewVBox(password),
		nil,
		container.NewVBox(cf.passwordEntry),
	)

	rtspH := widget.NewLabel("RTSP High:")
	rtspL := widget.NewLabel("RTSP Low: ")

	rtspHContainer := container.NewBorder(
		nil, nil,
		container.NewVBox(rtspH),
		nil,
		container.NewVBox(cf.rtspURI1Entry),
	)

	rtspLContainer := container.NewBorder(
		nil, nil,
		container.NewVBox(rtspL),
		nil,
		container.NewVBox(cf.rtspURI2Entry),
	)

	rtspContainer := container.NewBorder(
		nil, nil,
		nil,
		container.NewVBox(
			cf.checkButton,
		),
		container.NewVBox(
			rtspHContainer,
			rtspLContainer,
		),
	)

	cf.container = container.NewVBox(
		container.NewPadded(title),
		container.NewPadded(widget.NewSeparator()),
		container.NewPadded(
			container.NewGridWithColumns(4,
				ipContainer,
				portContainer,
				loginContainer,
				passwordContainer,
			),
		),
		container.NewPadded(widget.NewSeparator()),
		container.NewPadded(rtspContainer),
		container.NewPadded(widget.NewSeparator()),
	)
}

func (f *ConnectionForm) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(f.container)
}

func (f *ConnectionForm) GetConfig() *model.ConnectionConfig {
	port, _ := strconv.Atoi(f.portEntry.Text)

	return &model.ConnectionConfig{
		IP:       f.ipEntry.Text,
		Port:     port,
		Login:    f.loginEntry.Text,
		Password: f.passwordEntry.Text,
		RTSPURI1: f.rtspURI1Entry.Text,
		RTSPURI2: f.rtspURI2Entry.Text,
	}
}

func (f *ConnectionForm) SetOnCheck(handler func(*model.ConnectionConfig)) {
	f.onCheck = handler
}

func (f *ConnectionForm) SetOnConnect(handler func(*model.ConnectionConfig)) {
	f.onConnect = handler
}

func (f *ConnectionForm) SetOnDisconnect(handler func()) {
	f.onDisconnect = handler
}

func (f *ConnectionForm) SetConnected(connected bool) {
	if connected {
		f.connectButton.Disable()
		f.disconnectButton.Enable()
		f.checkButton.Disable()
	} else {
		f.connectButton.Enable()
		f.disconnectButton.Disable()
		f.checkButton.Enable()
	}
}

func (f *ConnectionForm) ConnectionPermission(permission bool) {
	if !permission {
		f.connectButton.Enable()
	} else {
		f.connectButton.Disable()
	}
}

func (f *ConnectionForm) LoadConfig(config *model.ConnectionConfig) {
	f.ipEntry.SetText(config.IP)
	f.portEntry.SetText(strconv.Itoa(config.Port))
	f.loginEntry.SetText(config.Login)
	f.rtspURI1Entry.SetText(config.RTSPURI1)
	f.rtspURI2Entry.SetText(config.RTSPURI2)
}
