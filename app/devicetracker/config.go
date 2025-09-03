package devicetracker

// Config конфигурация для модуля DeviceTracker
type Config struct {
	OutputDir             string `json:"output_dir"`
	SaveInterval          uint32 `json:"save_interval"`
	EnableTCPTracking     bool   `json:"enable_tcp_tracking"`
	EnableUDPTracking     bool   `json:"enable_udp_tracking"`
	EnableTrafficTracking bool   `json:"enable_traffic_tracking"`
}
