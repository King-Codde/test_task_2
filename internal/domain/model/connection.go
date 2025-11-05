package model

import "time"

type ConnectionConfig struct {
	IP       string
	Port     int
	Login    string
	Password string
	RTSPURI1 string
	RTSPURI2 string
}

type ResolvedURIs struct {
	URI1         string
	URI2         string
	URI1Masked   string
	URI2Masked   string
	AreIdentical bool
}

type StreamConfig struct {
	Name     string
	RTSPURI  string
	Login    string
	Password string
	Timeout  time.Duration
}

func NewStreamConfig(name, rtspURI, login, password string) *StreamConfig {
	return &StreamConfig{
		Name:     name,
		RTSPURI:  rtspURI,
		Login:    login,
		Password: password,
		Timeout:  10 * time.Second,
	}
}
