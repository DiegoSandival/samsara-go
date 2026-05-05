// handler.go
package samsara

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type CentralHandler struct {
	baseDir string
	mu      sync.RWMutex // Mejor usar RWMutex para permitir múltiples lecturas simultáneas
	stores  map[string]*Store
}

func NewCentralHandler() *CentralHandler {
	// Cargar configuración desde .env
	config, _ := LoadConfig(".env")
	if config == nil {
		config = &Config{}
	}

	return NewCentralHandlerWithDBPath(config.DBPath)
}

func NewCentralHandlerWithDBPath(dbPath string) *CentralHandler {
	if strings.TrimSpace(dbPath) == "" {
		dbPath = "./data"
	}

	// Crear el directorio base si no existe
	if dbPath != "" {
		os.MkdirAll(dbPath, 0755)
	}

	handler := &CentralHandler{
		baseDir: dbPath,
		stores:  make(map[string]*Store),
	}

	// Cargar todas las bases de datos existentes en la ruta configurada
	entries, err := os.ReadDir(dbPath)
	if err == nil {
		for _, entry := range entries {
			// Las DBs de ouroboros no tienen extensión. Verificamos que no sea un directorio
			// y que no termine en .bolt, para luego ver si tiene su archivo .bolt correspondiente.
			if !entry.IsDir() && !strings.HasSuffix(entry.Name(), ".bolt") {
				dbName := entry.Name()
				fullPath := filepath.Join(dbPath, dbName)

				// Verificamos que exista el archivo .bolt emparejado
				if _, errBolt := os.Stat(fullPath + ".bolt"); errBolt == nil {
					// Abrimos el Store existente
					store, errOpen := Open(fullPath)
					if errOpen == nil {
						handler.stores[dbName] = store
					}
				}
			}
		}
	}

	return handler
}

func (h *CentralHandler) RegisterStore(name string, store *Store) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stores[name] = store
}

func (h *CentralHandler) GetStore(name string) (*Store, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	store, exists := h.stores[name]
	return store, exists
}

func (h *CentralHandler) DeleteStore(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.stores, name)
}
