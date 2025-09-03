package devicetracker

import (
	"context"

	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
)

// IntegrationHooks предоставляет функции для интеграции с workers
type IntegrationHooks struct {
	integrationManager *IntegrationManager
}

// NewIntegrationHooks создает новые хуки интеграции
func NewIntegrationHooks(im *IntegrationManager) *IntegrationHooks {
	return &IntegrationHooks{
		integrationManager: im,
	}
}

// OnTCPConnection вызывается при новом TCP подключении
func (h *IntegrationHooks) OnTCPConnection(ctx context.Context, conn net.Conn, tag string) {
	if h.integrationManager == nil {
		return
	}

	// Получаем информацию об источнике из контекста
	if inbound := session.InboundFromContext(ctx); inbound != nil {
		source := inbound.Source
		h.integrationManager.TrackTCPConnection(ctx, source, tag)
	} else {
		// Если нет inbound в контексте, используем информацию о соединении
		source := net.DestinationFromAddr(conn.RemoteAddr())
		h.integrationManager.TrackTCPConnection(ctx, source, tag)
	}
}

// OnUDPConnection вызывается при новом UDP подключении
func (h *IntegrationHooks) OnUDPConnection(ctx context.Context, source net.Destination, tag string) {
	if h.integrationManager == nil {
		return
	}

	h.integrationManager.TrackUDPConnection(ctx, source, tag)
}

// OnTrafficUpdate вызывается при обновлении статистики трафика
func (h *IntegrationHooks) OnTrafficUpdate(ctx context.Context, source net.Destination, tag string, uplink, downlink int64) {
	if h.integrationManager == nil {
		return
	}

	h.integrationManager.TrackTraffic(ctx, source, tag, uplink, downlink)
}

// GetIntegrationManager возвращает менеджер интеграции
func (h *IntegrationHooks) GetIntegrationManager() *IntegrationManager {
	return h.integrationManager
}

// SetIntegrationManager устанавливает менеджер интеграции
func (h *IntegrationHooks) SetIntegrationManager(im *IntegrationManager) {
	h.integrationManager = im
}
