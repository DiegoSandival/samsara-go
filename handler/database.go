package samsara

import (
	"crypto/rand"
	"path/filepath"

	"github.com/DiegoSandival/ouroboros-go"
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

// CreateDB ejecuta la lógica y SIEMPRE retorna un []byte con la respuesta
func (h *CentralHandler) CreateDB(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.CreateDBReq(payload)
	if err != nil {
		return parser.CreateDBResultBytes(req.ID, 2)
	}

	//primero verificamos que no exista una base de datos con el mismo nombre
	if _, exists := h.GetStore(string(req.DBName)); exists {
		return parser.CreateDBResultBytes(req.ID, 3)
	}

	fullPath := filepath.Join(h.baseDir, string(req.DBName))
	store, err := NewStore(fullPath, req.DBSize) // Renombrado a NewStore por claridad
	if err != nil {
		return parser.CreateDBResultBytes(req.ID, 3)
	}

	h.RegisterStore(string(req.DBName), store)

	var salt [16]byte
	// Read llena el slice con bytes aleatorios seguros
	_, err = rand.Read(salt[:])
	if err != nil {
		// Este error es extremadamente raro, pero debe manejarse
		return parser.CreateDBResultBytes(req.ID, 4)
	}

	cell := store.NewCellWithSecret(salt, req.Secret, ouroboros.GenomaGenesis, 0, 0, 0)

	store.DB().Append(cell)

	return parser.CreateDBResultBytes(req.ID, 1)
}

func (h *CentralHandler) DelDB(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.DeleteDBReq(payload)
	if err != nil {
		return parser.DeleteDBResultBytes(req.ID, 2)
	}

	store, exists := h.GetStore(string(req.DBName))
	if !exists {
		//return []byte("base de datos no encontrada")
		return parser.DeleteDBResultBytes(req.ID, 3)
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
