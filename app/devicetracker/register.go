package devicetracker

import (
	"context"

	"github.com/xtls/xray-core/common"
)

// CreateObject создает экземпляр DeviceTracker из конфигурации
func CreateObject(ctx context.Context, config interface{}) (interface{}, error) {
	deviceTrackerConfig := config.(*Config)

	outputDir := deviceTrackerConfig.OutputDir
	if outputDir == "" {
		outputDir = "./device_logs"
	}

	manager, err := NewDeviceTrackerManager(ctx, outputDir)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// init регистрирует модуль в системе
func init() {
	common.Must(common.RegisterConfig((*Config)(nil), CreateObject))
}
