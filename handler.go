// handler.go
package samsara

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/DiegoSandival/ouroboros-go"
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

type CentralHandler struct {
	baseDir string
	mu      sync.RWMutex // Mejor usar RWMutex para permitir múltiples lecturas simultáneas
	stores  map[string]*Store
}

func NewCentralHandler() *CentralHandler {
	// Cargar configuración desde .env
	config, _ := LoadConfig(".env")

	// Crear el directorio base si no existe
	if config.DBPath != "" {
		os.MkdirAll(config.DBPath, 0755)
	}

	handler := &CentralHandler{
		baseDir: config.DBPath,
		stores:  make(map[string]*Store),
	}

	// Cargar todas las bases de datos existentes en la ruta configurada
	entries, err := os.ReadDir(config.DBPath)
	if err == nil {
		for _, entry := range entries {
			// Las DBs de ouroboros no tienen extensión. Verificamos que no sea un directorio
			// y que no termine en .bolt, para luego ver si tiene su archivo .bolt correspondiente.
			if !entry.IsDir() && !strings.HasSuffix(entry.Name(), ".bolt") {
				dbName := entry.Name()
				fullPath := filepath.Join(config.DBPath, dbName)

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

// CreateDB ejecuta la lógica y SIEMPRE retorna un []byte con la respuesta
func (h *CentralHandler) CreateDB(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.CreateDBReq(payload)
	if err != nil {
		return []byte("error parseando requerimiento")
	}

	fullPath := filepath.Join(h.baseDir, string(req.DBName))
	store, err := NewStore(fullPath, req.DBSize) // Renombrado a NewStore por claridad
	if err != nil {
		return []byte("error creando base de datos")
	}

	h.RegisterStore(string(req.DBName), store)

	var salt [16]byte
	// Read llena el slice con bytes aleatorios seguros
	_, err = rand.Read(salt[:])
	if err != nil {
		// Este error es extremadamente raro, pero debe manejarse
		return parser.CreateDBResultByte(req.ID, 2)
	}

	cell := store.NewCellWithSecret(salt, req.Secret, ouroboros.GenomaGenesis, 0, 0, 0)

	store.DB().Append(cell)

	return parser.CreateDBResultByte(req.ID, 1)

}

func (h *CentralHandler) DelDB(parser *protocol.ProtocolParser, payload []byte) []byte {

	req, err := parser.CreateDBReq(payload)
	if err != nil {
		return []byte("error parseando requerimiento")
	}

	_, exists := h.GetStore(string(req.DBName))
	if !exists {
		//return []byte("base de datos no encontrada")
		return parser.DeleteDBResultBytes(req.ID, 2)
	}

	fullPath := filepath.Join(h.baseDir, string(req.DBName))
	err = DeleteDB(fullPath)
	if err != nil {
		return parser.DeleteDBResultBytes(req.ID, 2)
	}

	return parser.DeleteDBResultBytes(req.ID, 1)
}

func (s *CentralHandler) Read(parser *protocol.ProtocolParser, payload []byte) []byte {

	req, err := parser.ReadReq(payload)
	if err != nil {
		//return []byte("error parseando requerimiento")
		return parser.ReadResultBytes(req.ID, 2, 0, nil)
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		//return []byte("base de datos no encontrada")
		return parser.ReadResultBytes(req.ID, 2, 0, nil)
	}

	//key string, cellIndex uint32, secret []byte
	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		//return ReadResult{Status: StatusUnauthorized}
		return parser.ReadResultBytes(req.ID, 2, 0, nil)
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		//return ReadResult{Status: StatusErrorDB}
		return parser.ReadResultBytes(req.ID, 2, 0, nil)
	}

	if !exists {
		return parser.ReadResultBytes(req.ID, 3, active.index, nil)
	}

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.LeerSelf, ouroboros.LeerAny)

	if active.cell.Genoma&requiredFlag == 0 {
		return parser.ReadResultBytes(req.ID, 2, active.index, nil)
	}

	newIndex, refreshed := store.refresh(active.index, req.Secret)
	if !refreshed {
		return parser.ReadResultBytes(req.ID, 2, active.index, nil)
	}

	return parser.ReadResultBytes(req.ID, 1, newIndex, cloneBytes(membrane.Value))

}

func (s *CentralHandler) Write(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.WriteReq(payload)
	if err != nil {
		return parser.WriteResultBytes(req.ID, 2, 0, nil)
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return parser.WriteResultBytes(req.ID, 2, 0, nil)
	}

	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return parser.WriteResultBytes(req.ID, 2, 0, nil)
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		return parser.WriteResultBytes(req.ID, 2, 0, nil)
	}

	if !exists {
		err = store.putMembrane(string(req.Key), Membrane{
			OwnerIndex: active.originalIndex,
			Value:      cloneBytes(req.Value),
		})
		if err != nil {
			return parser.WriteResultBytes(req.ID, 2, 0, nil)
		}

		newIndex, refreshed := store.refresh(active.index, req.Secret)
		if !refreshed {
			return parser.WriteResultBytes(req.ID, 2, 0, nil)
		}

		return parser.WriteResultBytes(req.ID, 1, newIndex, nil)
	}

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.EscribirSelf, ouroboros.EscribirAny)

	if active.cell.Genoma&requiredFlag == 0 {
		return parser.WriteResultBytes(req.ID, 2, active.index, nil)
	}

	membrane.Value = cloneBytes(req.Value)
	if err := store.putMembrane(string(req.Key), membrane); err != nil {
		return parser.WriteResultBytes(req.ID, 2, 0, nil)
	}

	newIndex, refreshed := store.refresh(active.index, req.Secret)
	if !refreshed {
		return parser.WriteResultBytes(req.ID, 2, 0, nil)
	}

	return parser.WriteResultBytes(req.ID, 1, newIndex, nil)

}
