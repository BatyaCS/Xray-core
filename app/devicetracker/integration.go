package devicetracker

import (
	"context"
	"sync"

	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
)

// IntegrationManager управляет интеграцией DeviceTracker с workers
type IntegrationManager struct {
	access   sync.RWMutex
	trackers map[string]*DeviceTrackerManager
	ctx      context.Context
}

// NewIntegrationManager создает новый менеджер интеграции
func NewIntegrationManager(ctx context.Context) *IntegrationManager {
	return &IntegrationManager{
		trackers: make(map[string]*DeviceTrackerManager),
		ctx:      ctx,
	}
}

// RegisterTracker регистрирует DeviceTracker для указанного тега
func (im *IntegrationManager) RegisterTracker(tag string, tracker *DeviceTrackerManager) {
	im.access.Lock()
	defer im.access.Unlock()
	im.trackers[tag] = tracker
}

// GetTracker возвращает DeviceTracker для указанного тега
func (m *IntegrationManager) GetTracker(tag string) *DeviceTrackerManager {
	m.access.RLock()
	defer m.access.RUnlock()
	return m.trackers[tag]
}

// TrackTCPConnection отслеживает TCP подключение
func (im *IntegrationManager) TrackTCPConnection(ctx context.Context, source net.Destination, tag string) {
	if tracker := im.GetTracker(tag); tracker != nil {
		tracker.AddConnection(ctx, source.Address.String(), int(source.Port), "TCP", tag)
	}
}

// TrackUDPConnection отслеживает UDP подключение
func (im *IntegrationManager) TrackUDPConnection(ctx context.Context, source net.Destination, tag string) {
	if tracker := im.GetTracker(tag); tracker != nil {
		tracker.AddConnection(ctx, source.Address.String(), int(source.Port), "UDP", tag)
	}
}

// TrackTraffic отслеживает трафик
func (im *IntegrationManager) TrackTraffic(ctx context.Context, source net.Destination, tag string, uplink, downlink int64) {
	if tracker := im.GetTracker(tag); tracker != nil {
		tracker.AddConnection(ctx, source.Address.String(), int(source.Port), "UDP", tag)
	}
}

// Close закрывает менеджер интеграции
func (im *IntegrationManager) Close() error {
	im.access.Lock()
	defer im.access.Unlock()

	for _, tracker := range im.trackers {
		if err := tracker.Close(); err != nil {
			errors.LogWarningInner(im.ctx, err, "failed to close device tracker")
		}
	}

	return nil
}
