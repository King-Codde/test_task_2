package service

import (
	"encoding/json"
	"ip-camera-viewer/internal/domain/model"
	"os"
	"path/filepath"
)

type ConfigurationService struct {
	logger     *LoggerService
	configPath string
}

type SavedConfig struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Login    string `json:"login"`
	RTSPURI1 string `json:"rtsp_uri_1"`
	RTSPURI2 string `json:"rtsp_uri_2"`
}

func NewConfigurationService(logger *LoggerService) *ConfigurationService {
	currentDir, _ := os.Getwd()
	configPath := filepath.Join(currentDir, "config.json")

	return &ConfigurationService{
		logger:     logger,
		configPath: configPath,
	}
}

func (cs *ConfigurationService) SaveConfig(config *model.ConnectionConfig) error {
	saved := &SavedConfig{
		IP:       config.IP,
		Port:     config.Port,
		Login:    config.Login,
		RTSPURI1: config.RTSPURI1,
		RTSPURI2: config.RTSPURI2,
	}

	dir := filepath.Dir(cs.configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(cs.configPath, data, 0600)
	if err != nil {
		cs.logger.Error("Ошибка сохранения конфигурации", err)
		return err
	}

	cs.logger.Info("Конфигурация сохранена")
	return nil
}

func (cs *ConfigurationService) LoadConfig() (*SavedConfig, error) {
	data, err := os.ReadFile(cs.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var saved SavedConfig
	if err := json.Unmarshal(data, &saved); err != nil {
		cs.logger.Error("Ошибка чтения конфигурации", err)
		return nil, err
	}

	cs.logger.Info("Конфигурация загружена")
	return &saved, nil
}
