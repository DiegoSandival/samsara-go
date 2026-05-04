package samsara

import (
	"crypto/rand"
	"path/filepath"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

// CreateDB ejecuta la lógica y SIEMPRE retorna un []byte con la respuesta
func (h *CentralHandler) CreateDB(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.CreateDBReq(payload)
	if err != nil {
		return requestParseError(parser, payload, protocol.OpcodeCreateDB, err, "create_db.parse")
	}

	if req.Genome&ouroboros.IsMigrated != 0 {
		return errorWithID(parser, req.ID, protocol.ErrorCodeInvalidGenomeFlags, "create_db.genome")
	}

	if _, exists := h.GetStore(string(req.DBName)); exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeDatabaseAlreadyExists, "create_db.exists")
	}

	fullPath := filepath.Join(h.baseDir, string(req.DBName))
	store, err := NewStore(fullPath, req.DBSize) // Renombrado a NewStore por claridad
	if err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeStoreCreateFailed, "create_db.store")
	}

	var salt [16]byte
	// Read llena el slice con bytes aleatorios seguros
	_, err = rand.Read(salt[:])
	if err != nil {
		// Este error es extremadamente raro, pero debe manejarse
		_ = store.Destroy()
		return errorWithID(parser, req.ID, protocol.ErrorCodeRandomSourceFailed, "create_db.salt")
	}

	cell := store.NewCellWithSecret(salt, req.Secret, req.Genome, 0, 0, 0)

	if _, err := store.DB().Append(cell); err != nil {
		_ = store.Destroy()
		return errorWithID(parser, req.ID, protocol.ErrorCodeInitialCellAppendFail, "create_db.root_append")
	}

	h.RegisterStore(string(req.DBName), store)

	return parser.CreateDBResultBytes(req.ID, 1)
}

func (h *CentralHandler) DelDB(parser *protocol.ProtocolParser, payload []byte) []byte {

	req, err := parser.DeleteDBReq(payload)
	if err != nil {
		return requestParseError(parser, payload, protocol.OpcodeDeleteDB, err, "delete_db.parse")
	}

	store, exists := h.GetStore(string(req.DBName))
	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeDatabaseNotFound, "delete_db.store")
	}

	authorized := store.ResolveCellAuth(req.CellIndex, req.Secret)
	if !authorized {
		return errorWithID(parser, req.ID, protocol.ErrorCodeAuthenticationFailed, "delete_db.auth")
	}

	err = store.Destroy()
	if err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeStoreDestroyFailed, "delete_db.destroy")
	}

	h.DeleteStore(string(req.DBName))

	return parser.DeleteDBResultBytes(req.ID, 1)
}
