package services

import (
	"sync"

	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Usamos un contenedor sqlstore temporal solo para inicializar dispositivos
// pero NO guardamos nada en SQL, todo se guarda en memoria
var tempContainer *sqlstore.Container
var tempContainerOnce sync.Once

// MemoryContainer simula el comportamiento de sqlstore.Container pero usando solo memoria
type MemoryContainer struct {
	mu      sync.RWMutex
	devices map[string]*store.Device
}

// NewMemoryContainer crea un nuevo contenedor en memoria
func NewMemoryContainer() *MemoryContainer {
	return &MemoryContainer{
		devices: make(map[string]*store.Device),
	}
}

// GetAllDevices retorna todos los dispositivos almacenados en memoria
func (m *MemoryContainer) GetAllDevices(ctx context.Context) ([]*store.Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]*store.Device, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}

	return devices, nil
}

// NewDevice crea un nuevo dispositivo en memoria
// Usa un contenedor temporal solo para inicializar correctamente el Device
// pero el dispositivo se guarda en memoria, no en SQL
func (m *MemoryContainer) NewDevice() *store.Device {
	// Inicializar el contenedor temporal una sola vez
	tempContainerOnce.Do(func() {
		ctx := context.Background()
		var err error
		// Usamos :memory: que es SQLite en memoria (no se persiste en disco)
		// Solo lo usamos para inicializar el Device correctamente
		tempContainer, err = sqlstore.New(ctx, "sqlite3", ":memory:", nil)
		if err != nil {
			zap.L().Error("failed to create temp container for device initialization", zap.Error(err))
		}
	})

	if tempContainer != nil {
		// Use the temporary container only to create the Device with initialized stores
		device := tempContainer.NewDevice()
		return device
	}

	// Fallback: create empty Device (may cause issues but better than panic)
	zap.L().Warn("using empty device, may cause issues")
	return &store.Device{}
}

// SaveDevice guarda un dispositivo en memoria
func (m *MemoryContainer) SaveDevice(ctx context.Context, device *store.Device) error {
	if device == nil || device.ID == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Crear una copia del dispositivo para evitar modificaciones externas
	deviceCopy := *device
	m.devices[device.ID.String()] = &deviceCopy

	return nil
}

// DeleteDevice elimina un dispositivo de la memoria
func (m *MemoryContainer) DeleteDevice(ctx context.Context, device *store.Device) error {
	if device == nil || device.ID == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.devices, device.ID.String())
	zap.L().Debug("device deleted from memory", zap.String("jid", device.ID.String()))

	return nil
}

// GetDevice obtiene un dispositivo por su JID
func (m *MemoryContainer) GetDevice(ctx context.Context, jid types.JID) (*store.Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	device, exists := m.devices[jid.String()]
	if !exists {
		return nil, nil
	}

	// Retornar una copia
	deviceCopy := *device
	return &deviceCopy, nil
}

