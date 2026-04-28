package samsara

import (
	"github.com/DiegoSandival/ouroboros-go"
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

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

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.ReadOwn, ouroboros.ReadAll)

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

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.WriteOwn, ouroboros.WriteAll)

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
		return parser.DeleteDBResultBytes(req.ID, 3)
	}

	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return parser.DeleteDBResultBytes(req.ID, 4)
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		return parser.DeleteDBResultBytes(req.ID, 5)
	}

	if !exists {
		return parser.DeleteDBResultBytes(req.ID, 6)
	}

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.DeleteOwn, ouroboros.DeleteAll)

	if active.cell.Genoma&requiredFlag == 0 {
		return parser.DeleteDBResultBytes(req.ID, 7)
	}

	store.deleteMembrane(string(req.Key))

	return parser.DeleteDBResultBytes(req.ID, 1)
}
