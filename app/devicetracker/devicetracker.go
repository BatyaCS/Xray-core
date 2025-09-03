package devicetracker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/net"
)

// DeviceInfo содержит информацию об устройстве
type DeviceInfo struct {
	IP              string
	Port            int
	Country         string
	City            string
	FirstSeen       time.Time
	LastSeen        time.Time
	TotalUplink     int64
	TotalDownlink   int64
	ConnectionCount int
	Protocols       map[string]bool // TCP, UDP
	Tags            map[string]bool // inbound tags
}

// DeviceTracker отслеживает устройства и сохраняет данные в файлы
type DeviceTracker struct {
	access      sync.RWMutex
	devices     map[string]*DeviceInfo // key: IP:Port
	outputDir   string
	currentFile string
	fileMutex   sync.Mutex

	ctx context.Context
}

// NewDeviceTracker создает новый экземпляр DeviceTracker
func NewDeviceTracker(ctx context.Context, outputDir string) (*DeviceTracker, error) {
	if outputDir == "" {
		outputDir = "./device_logs"
	}

	// Создаем директорию если не существует
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, errors.New("failed to create output directory").Base(err)
	}

	dt := &DeviceTracker{
		devices:   make(map[string]*DeviceInfo),
		outputDir: outputDir,
		ctx:       ctx,
	}

	// Инициализируем файл для текущего месяца
	if err := dt.initializeMonthlyFile(); err != nil {
		return nil, err
	}

	// Запускаем периодическое сохранение
	go dt.periodicSave()

	return dt, nil
}

// initializeMonthlyFile создает или открывает файл для текущего месяца
func (dt *DeviceTracker) initializeMonthlyFile() error {
	dt.fileMutex.Lock()
	defer dt.fileMutex.Unlock()

	now := time.Now()
	monthStr := now.Format("2006-01")
	fileName := fmt.Sprintf("devices_%s.txt", monthStr)
	filePath := filepath.Join(dt.outputDir, fileName)

	// Если файл уже открыт, ничего не делаем
	if dt.currentFile == filePath {
		return nil
	}

	// Закрываем предыдущий файл если он был
	if dt.currentFile != "" {
		dt.saveToFile(dt.currentFile)
	}

	// Создаем новый файл
	dt.currentFile = filePath

	// Создаем заголовок файла
	file, err := os.Create(filePath)
	if err != nil {
		return errors.New("failed to create monthly file").Base(err)
	}
	defer file.Close()

	// Записываем заголовок
	header := fmt.Sprintf(`Device Tracker - %s
Generated: %s

%-15s %-8s %-12s %-15s %-20s %-20s %-12s %-12s %-8s %-10s %-15s
%s
`,
		monthStr,
		now.Format("2006-01-02 15:04:05"),
		"IP Address", "Port", "Country", "City", "First Seen", "Last Seen", "Uplink (MB)", "Downlink (MB)", "Connections", "Protocols", "Tags",
		"----------------------------------------------------------------------------------------------------------------------------------------------------------------")

	_, err = file.WriteString(header)
	return err
}

// AddConnection добавляет информацию о новом подключении
func (dt *DeviceTracker) AddConnection(ctx context.Context, source net.Destination, protocol string, tag string) {
	dt.access.Lock()
	defer dt.access.Unlock()

	key := fmt.Sprintf("%s:%d", source.Address.String(), source.Port)
	now := time.Now()

	device, exists := dt.devices[key]
	if !exists {
		device = &DeviceInfo{
			IP:              source.Address.String(),
			Port:            int(source.Port),
			Country:         "Unknown",
			City:            "Unknown",
			FirstSeen:       now,
			LastSeen:        now,
			TotalUplink:     0,
			TotalDownlink:   0,
			ConnectionCount: 0,
			Protocols:       make(map[string]bool),
			Tags:            make(map[string]bool),
		}

		dt.devices[key] = device
	} else {
		device.LastSeen = now
	}

	device.ConnectionCount++
	device.Protocols[protocol] = true
	if tag != "" {
		device.Tags[tag] = true
	}
}

// AddTraffic добавляет информацию о трафике для устройства
func (dt *DeviceTracker) AddTraffic(ctx context.Context, source net.Destination, uplink, downlink int64) {
	dt.access.RLock()
	defer dt.access.RUnlock()

	key := fmt.Sprintf("%s:%d", source.Address.String(), source.Port)
	if device, exists := dt.devices[key]; exists {
		device.TotalUplink += uplink
		device.TotalDownlink += downlink
	}
}

// GetDeviceInfo возвращает информацию об устройстве
func (dt *DeviceTracker) GetDeviceInfo(ip string, port int) *DeviceInfo {
	dt.access.RLock()
	defer dt.access.RUnlock()

	key := fmt.Sprintf("%s:%d", ip, port)
	return dt.devices[key]
}

// GetAllDevices возвращает все устройства
func (dt *DeviceTracker) GetAllDevices() map[string]*DeviceInfo {
	dt.access.RLock()
	defer dt.access.RUnlock()

	result := make(map[string]*DeviceInfo)
	for k, v := range dt.devices {
		result[k] = v
	}
	return result
}

// saveToFile сохраняет данные в указанный файл
func (dt *DeviceTracker) saveToFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New("failed to open file for writing").Base(err)
	}
	defer file.Close()

	dt.access.RLock()
	devices := make([]*DeviceInfo, 0, len(dt.devices))
	for _, device := range dt.devices {
		devices = append(devices, device)
	}
	dt.access.RUnlock()

	for _, device := range devices {
		// Форматируем протоколы
		protocols := ""
		for protocol := range device.Protocols {
			if protocols != "" {
				protocols += ","
			}
			protocols += protocol
		}

		// Форматируем теги
		tags := ""
		for tag := range device.Tags {
			if tags != "" {
				tags += ","
			}
			tags += tag
		}

		// Форматируем размеры трафика в MB
		uplinkMB := float64(device.TotalUplink) / (1024 * 1024)
		downlinkMB := float64(device.TotalDownlink) / (1024 * 1024)

		line := fmt.Sprintf("%-15s %-8d %-12s %-15s %-20s %-20s %-12.2f %-12.2f %-8d %-10s %-15s\n",
			device.IP,
			device.Port,
			device.Country,
			device.City,
			device.FirstSeen.Format("2006-01-02 15:04:05"),
			device.LastSeen.Format("2006-01-02 15:04:05"),
			uplinkMB,
			downlinkMB,
			device.ConnectionCount,
			protocols,
			tags,
		)

		if _, err := file.WriteString(line); err != nil {
			return errors.New("failed to write to file").Base(err)
		}
	}

	return nil
}

// periodicSave периодически сохраняет данные и проверяет смену месяца
func (dt *DeviceTracker) periodicSave() {
	ticker := time.NewTicker(5 * time.Minute) // Сохраняем каждые 5 минут
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Проверяем, не сменился ли месяц
			now := time.Now()
			currentMonth := now.Format("2006-01")
			if dt.currentFile != "" {
				fileName := filepath.Base(dt.currentFile)
				if !strings.Contains(fileName, currentMonth) {
					// Месяц сменился, создаем новый файл
					if err := dt.initializeMonthlyFile(); err != nil {
						errors.LogWarningInner(dt.ctx, err, "failed to initialize new monthly file")
					}
				}
			}

			// Сохраняем текущие данные
			if dt.currentFile != "" {
				if err := dt.saveToFile(dt.currentFile); err != nil {
					errors.LogWarningInner(dt.ctx, err, "failed to save device data")
				}
			}

		case <-dt.ctx.Done():
			return
		}
	}
}

// Close закрывает трекер и сохраняет все данные
func (dt *DeviceTracker) Close() error {
	if dt.currentFile != "" {
		return dt.saveToFile(dt.currentFile)
	}
	return nil
}
