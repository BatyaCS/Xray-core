package devicetracker

import (
	"context"

	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/features"
)

// DeviceTrackerFeature интерфейс для интеграции с системой features
type DeviceTrackerFeature interface {
	features.Feature

	// AddConnection добавляет информацию о новом подключении
	AddConnection(ctx context.Context, sourceIP string, sourcePort int, protocol string, tag string)

	// AddTraffic добавляет информацию о трафике
	AddTraffic(ctx context.Context, sourceIP string, sourcePort int, uplink, downlink int64)

	// GetDeviceInfo возвращает информацию об устройстве
	GetDeviceInfo(ip string, port int) *DeviceInfo

	// GetAllDevices возвращает все устройства
	GetAllDevices() map[string]*DeviceInfo
}

// DeviceTrackerManager управляет экземпляром DeviceTracker
type DeviceTrackerManager struct {
	tracker *DeviceTracker
}

// NewDeviceTrackerManager создает новый менеджер DeviceTracker
func NewDeviceTrackerManager(ctx context.Context, outputDir string) (*DeviceTrackerManager, error) {
	tracker, err := NewDeviceTracker(ctx, outputDir)
	if err != nil {
		return nil, err
	}

	return &DeviceTrackerManager{
		tracker: tracker,
	}, nil
}

// Type реализует features.Feature
func (*DeviceTrackerManager) Type() interface{} {
	return DeviceTrackerType()
}

// Start запускает менеджер
func (m *DeviceTrackerManager) Start() error {
	// DeviceTracker уже запущен в NewDeviceTracker
	return nil
}

// Close закрывает менеджер
func (m *DeviceTrackerManager) Close() error {
	return m.tracker.Close()
}

// AddConnection добавляет информацию о новом подключении
func (m *DeviceTrackerManager) AddConnection(ctx context.Context, sourceIP string, sourcePort int, protocol string, tag string) {
	// Создаем временный Destination для передачи в tracker
	source := net.Destination{
		Address: net.ParseAddress(sourceIP),
		Port:    net.Port(sourcePort),
		Network: net.Network_TCP, // По умолчанию TCP, но это не важно для отслеживания
	}

	m.tracker.AddConnection(ctx, source, protocol, tag)
}

// AddTraffic добавляет информацию о трафике
func (m *DeviceTrackerManager) AddTraffic(ctx context.Context, sourceIP string, sourcePort int, uplink, downlink int64) {
	source := net.Destination{
		Address: net.ParseAddress(sourceIP),
		Port:    net.Port(sourcePort),
		Network: net.Network_TCP,
	}

	m.tracker.AddTraffic(ctx, source, uplink, downlink)
}

// GetDeviceInfo возвращает информацию об устройстве
func (m *DeviceTrackerManager) GetDeviceInfo(ip string, port int) *DeviceInfo {
	return m.tracker.GetDeviceInfo(ip, port)
}

// GetAllDevices возвращает все устройства
func (m *DeviceTrackerManager) GetAllDevices() map[string]*DeviceInfo {
	return m.tracker.GetAllDevices()
}

// DeviceTrackerType возвращает тип для DeviceTracker
func DeviceTrackerType() interface{} {
	return (*DeviceTrackerFeature)(nil)
}
