package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type LogPanel struct {
	widget.BaseWidget
	entry    *widget.Entry
	maxLines int
	logs     []string
}

func NewLogPanel() *LogPanel {
	p := &LogPanel{
		entry:    widget.NewMultiLineEntry(),
		maxLines: 100,
		logs:     make([]string, 0),
	}

	p.entry.SetPlaceHolder("Логи и сообщения валидации...")

	p.ExtendBaseWidget(p)
	return p
}

func (p *LogPanel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.entry)
}

func (p *LogPanel) AddLog(message string) {
	p.logs = append(p.logs, message)

	if len(p.logs) > p.maxLines {
		p.logs = p.logs[len(p.logs)-p.maxLines:]
	}

	p.entry.SetText(strings.Join(p.logs, "\n"))

	p.entry.CursorRow = len(p.logs)
}

func (p *LogPanel) Clear() {
	p.logs = make([]string, 0)
	p.entry.SetText("")
}

func (p *LogPanel) GetWidget() *widget.Entry {
	return p.entry
}
