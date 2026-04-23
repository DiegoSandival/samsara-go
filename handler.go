// handler.go
package samsara

import (
	"crypto/rand"
	"log"
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

func (h *CentralHandler) DeleteStore(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.stores, name)
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
		return parser.CreateDBResultBytes(req.ID, 2)
	}

	cell := store.NewCellWithSecret(salt, req.Secret, ouroboros.GenomaGenesis, 0, 0, 0)

	store.DB().Append(cell)

	return parser.CreateDBResultBytes(req.ID, 1)

}

func (h *CentralHandler) DelDB(parser *protocol.ProtocolParser, payload []byte) []byte {

	req, err := parser.DeleteDBReq(payload)
	if err != nil {
		return []byte("error parseando requerimiento")
	}

	store, exists := h.GetStore(string(req.DBName))
	if !exists {
		//return []byte("base de datos no encontrada")
		return parser.DeleteDBResultBytes(req.ID, 2)
	}

	_, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		//return ReadResult{Status: StatusUnauthorized}
		return parser.ReadResultBytes(req.ID, 2, 0, nil)
	}

	err = store.Destroy()
	if err != nil {
		//return []byte("error eliminando base de datos")
		return parser.DeleteDBResultBytes(req.ID, 3)
	}

	h.DeleteStore(string(req.DBName))

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
		return parser.WriteResultBytes(req.ID, 10, 0, nil)
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return parser.WriteResultBytes(req.ID, 11, 0, nil)
	}

	log.Printf("Intentando resolver cellIndex %d con secreto %s\n", req.CellIndex, string(req.Secret))
	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return parser.WriteResultBytes(req.ID, 12, 0, nil)
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		return parser.WriteResultBytes(req.ID, 13, 0, nil)
	}

	if !exists {
		err = store.putMembrane(string(req.Key), Membrane{
			OwnerIndex: active.originalIndex,
			Value:      cloneBytes(req.Value),
		})
		if err != nil {
			return parser.WriteResultBytes(req.ID, 14, 0, nil)
		}

		newIndex, refreshed := store.refresh(active.index, req.Secret)
		if !refreshed {
			return parser.WriteResultBytes(req.ID, 15, 0, nil)
		}

		return parser.WriteResultBytes(req.ID, 1, newIndex, nil)
	}

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.EscribirSelf, ouroboros.EscribirAny)

	if active.cell.Genoma&requiredFlag == 0 {
		return parser.WriteResultBytes(req.ID, 16, active.index, nil)
	}

	membrane.Value = cloneBytes(req.Value)
	if err := store.putMembrane(string(req.Key), membrane); err != nil {
		return parser.WriteResultBytes(req.ID, 17, 0, nil)
	}

	newIndex, refreshed := store.refresh(active.index, req.Secret)
	if !refreshed {
		return parser.WriteResultBytes(req.ID, 18, 0, nil)
	}

	return parser.WriteResultBytes(req.ID, 1, newIndex, nil)

}

func (s *CentralHandler) ReadFree(parser *protocol.ProtocolParser, payload []byte) []byte {

	req, err := parser.ReadFreeReq(payload)
	if err != nil {
		return parser.ReadFreeResultBytes(req.ID, 1, nil)
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return parser.ReadFreeResultBytes(req.ID, 2, nil)
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		return parser.ReadFreeResultBytes(req.ID, 3, nil)
	}

	if !exists {
		return parser.ReadFreeResultBytes(req.ID, 4, nil)
	}

	return parser.ReadFreeResultBytes(req.ID, 1, cloneBytes(membrane.Value))

}

func (s *CentralHandler) Delete(parser *protocol.ProtocolParser, payload []byte) []byte {

	req, err := parser.DeleteReq(payload)
	if err != nil {
		return parser.DeleteDBResultBytes(req.ID, 2)
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return parser.DeleteDBResultBytes(req.ID, 2)
	}

	log.Printf("Intentando resolver cellIndex %d con secreto %s\n", req.CellIndex, string(req.Secret))
	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return parser.DeleteDBResultBytes(req.ID, 2)
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		return parser.DeleteDBResultBytes(req.ID, 2)
	}

	if !exists {
		return parser.DeleteDBResultBytes(req.ID, 3)
	}

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.BorrarSelf, ouroboros.BorrarAny)

	if active.cell.Genoma&requiredFlag == 0 {
		return parser.DeleteDBResultBytes(req.ID, 16)
	}

	store.deleteMembrane(string(req.Key))

	//newIndex, refreshed := store.refresh(active.index, req.Secret)
	//if !refreshed {
	//	return parser.DeleteDBResultBytes(req.ID, 2)
	//}

	return parser.DeleteDBResultBytes(req.ID, 1)

}

func (s *CentralHandler) ReadCell(parser *protocol.ProtocolParser, payload []byte) []byte {

	req, err := parser.ReadCellReq(payload)
	if err != nil {
		return parser.ReadCellResultBytes(req.ID, 2, nil)
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return parser.ReadCellResultBytes(req.ID, 3, nil)
	}

	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return parser.ReadCellResultBytes(req.ID, 4, nil)
	}

	return parser.ReadCellResultBytes(req.ID, 1, parser.CellBytes(active.cell))

}

func (s *CentralHandler) Diferir(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.DiferirReq(payload)
	if err != nil {
		return parser.DiferirResultBytes(req.ID, 2, 0)
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return parser.DiferirResultBytes(req.ID, 3, 0)
	}

	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return parser.DiferirResultBytes(req.ID, 4, 0)
	}

	if active.cell.Genoma&req.ChildGenome != req.ChildGenome {
		return parser.DiferirResultBytes(req.ID, 16, active.index)
	}

	childCell := store.NewCellWithSecret([16]byte{}, req.ChildSecret, req.ChildGenome, req.X, req.Y, req.Z)

	childIndex, _ := store.DB().Append(childCell)

	return parser.DiferirResultBytes(req.ID, 1, childIndex)
}

func (s *CentralHandler) Cruzar(parser *protocol.ProtocolParser, payload []byte) []byte {

	req, err := parser.CruzarReq(payload)
	if err != nil {
		return parser.CruzarResultBytes(req.ID, 2, 0)
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return parser.CruzarResultBytes(req.ID, 3, 0)
	}

	active, ok := store.resolveCell(req.CellIndexA, req.SecretA)
	if !ok {
		return parser.CruzarResultBytes(req.ID, 4, 0)
	}

	if active.cell.Genoma&req.ChildGenome != req.ChildGenome {
		return parser.CruzarResultBytes(req.ID, 16, active.index)
	}

	childCell := store.NewCellWithSecret([16]byte{}, req.ChildSecret, req.ChildGenome, req.X, req.Y, req.Z)

	childIndex, _ := store.DB().Append(childCell)

	return parser.CruzarResultBytes(req.ID, 1, childIndex)
}
