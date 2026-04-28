package samsara

import (
	"crypto/rand"
	"log"
	"path/filepath"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

// CreateDB ejecuta la lógica y SIEMPRE retorna un []byte con la respuesta
func (h *CentralHandler) CreateDB(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.CreateDBReq(payload)
	if err != nil {
		return parser.CreateDBResultBytes(req.ID, 2)
	}

	//no se puede crear si el genoma incluye migrada activada
	if req.Genome&ouroboros.IsMigrated != 0 {
		return parser.CreateDBResultBytes(req.ID, 3)
	}

	if _, exists := h.GetStore(string(req.DBName)); exists {
		return parser.CreateDBResultBytes(req.ID, 4)
	}

	fullPath := filepath.Join(h.baseDir, string(req.DBName))
	store, err := NewStore(fullPath, req.DBSize) // Renombrado a NewStore por claridad
	if err != nil {
		return parser.CreateDBResultBytes(req.ID, 5)
	}

	h.RegisterStore(string(req.DBName), store)

	var salt [16]byte
	// Read llena el slice con bytes aleatorios seguros
	_, err = rand.Read(salt[:])
	if err != nil {
		// Este error es extremadamente raro, pero debe manejarse
		return parser.CreateDBResultBytes(req.ID, 6)
	}

	cell := store.NewCellWithSecret(salt, req.Secret, req.Genome, 0, 0, 0)

	store.DB().Append(cell)

	return parser.CreateDBResultBytes(req.ID, 1)
}

func (h *CentralHandler) DelDB(parser *protocol.ProtocolParser, payload []byte) []byte {
	log.Println("DelDB called")
	req, err := parser.DeleteDBReq(payload)
	if err != nil {
		return parser.DeleteDBResultBytes(req.ID, 2)
	}

	log.Printf("Attempting to delete DB: %s\n", string(req.DBName))
	store, exists := h.GetStore(string(req.DBName))
	if !exists {
		//return []byte("base de datos no encontrada")
		return parser.DeleteDBResultBytes(req.ID, 3)
	}

	log.Printf("DB found: %s, verifying secret...\n", string(req.DBName))
	log.Printf("Cell index: %d, Secret: %s\n", req.CellIndex, string(req.Secret))

	authorized := store.ResolveCellAuth(req.CellIndex, req.Secret)
	if !authorized {
		return parser.DeleteDBResultBytes(req.ID, 4)
	}

	log.Printf("Secret verified for DB: %s, destroying store...\n", string(req.DBName))
	err = store.Destroy()
	if err != nil {
		//return []byte("error eliminando base de datos")
		return parser.DeleteDBResultBytes(req.ID, 5)
	}

	log.Printf("Store destroyed for DB: %s, removing from handler...\n", string(req.DBName))
	h.DeleteStore(string(req.DBName))

	return parser.DeleteDBResultBytes(req.ID, 1)
}
