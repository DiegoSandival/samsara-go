// handler.go
package main

import (
	"sync"

	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

type CentralHandler struct {
	baseDir string
	mu      sync.RWMutex // Mejor usar RWMutex para permitir múltiples lecturas simultáneas
	stores  map[string]*Store
}

func NewCentralHandler(baseDir string) *CentralHandler {
	return &CentralHandler{
		baseDir: baseDir,
		stores:  make(map[string]*Store),
	}
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

// CreateDB ejecuta la lógica y SIEMPRE retorna un []byte con la respuesta
func (h *CentralHandler) CreateDB(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.CreateDBReq(payload)
	if err != nil {
		// Ojo: Asegúrate de que Serialize() devuelva []byte
		// return protocol.CreateDBResponse{Success: false, Error: err.Error()}.Serialize()
		return []byte("error parseando requerimiento")
	}

	store, err := NewStore(string(req.DBName), req.GenesisIndex) // Renombrado a NewStore por claridad
	if err != nil {
		return []byte("error creando base de datos")
	}

	h.RegisterStore(string(req.DBName), store)

	return parser.CreateDBResultByte(req.ID, 1)
}
