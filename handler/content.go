package samsara

import (
	"github.com/DiegoSandival/ouroboros-go"
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

func (s *CentralHandler) Read(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.ReadReq(payload)
	if err != nil {
		return requestParseError(parser, payload, protocol.OpcodeRead, err, "read.parse")
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeDatabaseNotFound, "read.store")
	}

	//key string, cellIndex uint32, secret []byte
	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return errorWithID(parser, req.ID, protocol.ErrorCodeAuthenticationFailed, "read.auth")
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneReadFailed, "read.membrane")
	}

	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneNotFound, "read.not_found")
	}

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.ReadOwn, ouroboros.ReadAll)

	if active.cell.Genoma&requiredFlag == 0 {
		return errorWithID(parser, req.ID, protocol.ErrorCodePermissionDenied, "read.permission")
	}

	newIndex, refreshed := store.refresh(active.index, req.Secret)
	if !refreshed {
		return errorWithID(parser, req.ID, protocol.ErrorCodeCellRefreshFailed, "read.refresh")
	}

	return parser.ReadResultBytes(req.ID, 1, newIndex, cloneBytes(membrane.Value))
}

func (s *CentralHandler) Write(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.WriteReq(payload)
	if err != nil {
		return requestParseError(parser, payload, protocol.OpcodeWrite, err, "write.parse")
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeDatabaseNotFound, "write.store")
	}

	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return errorWithID(parser, req.ID, protocol.ErrorCodeAuthenticationFailed, "write.auth")
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneReadFailed, "write.membrane")
	}

	if !exists {
		err = store.putMembrane(string(req.Key), Membrane{
			OwnerIndex: active.originalIndex,
			Value:      cloneBytes(req.Value),
		})
		if err != nil {
			return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneWriteFailed, "write.insert")
		}

		newIndex, refreshed := store.refresh(active.index, req.Secret)
		if !refreshed {
			return errorWithID(parser, req.ID, protocol.ErrorCodeCellRefreshFailed, "write.insert_refresh")
		}

		return parser.WriteResultBytes(req.ID, 1, newIndex, nil)
	}

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.WriteOwn, ouroboros.WriteAll)

	if active.cell.Genoma&requiredFlag == 0 {
		return errorWithID(parser, req.ID, protocol.ErrorCodePermissionDenied, "write.permission")
	}

	membrane.Value = cloneBytes(req.Value)
	if err := store.putMembrane(string(req.Key), membrane); err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneWriteFailed, "write.update")
	}

	newIndex, refreshed := store.refresh(active.index, req.Secret)
	if !refreshed {
		return errorWithID(parser, req.ID, protocol.ErrorCodeCellRefreshFailed, "write.update_refresh")
	}

	return parser.WriteResultBytes(req.ID, 1, newIndex, nil)
}

func (s *CentralHandler) ReadFree(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.ReadFreeReq(payload)
	if err != nil {
		return requestParseError(parser, payload, protocol.OpcodeReadFree, err, "read_free.parse")
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeDatabaseNotFound, "read_free.store")
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneReadFailed, "read_free.membrane")
	}

	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneNotFound, "read_free.not_found")
	}

	return parser.ReadFreeResultBytes(req.ID, 1, cloneBytes(membrane.Value))
}

func (s *CentralHandler) Delete(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.DeleteReq(payload)
	if err != nil {
		return requestParseError(parser, payload, protocol.OpcodeDelete, err, "delete.parse")
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeDatabaseNotFound, "delete.store")
	}

	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return errorWithID(parser, req.ID, protocol.ErrorCodeAuthenticationFailed, "delete.auth")
	}

	membrane, exists, err := store.getMembrane(string(req.Key))
	if err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneReadFailed, "delete.membrane")
	}

	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneNotFound, "delete.not_found")
	}

	requiredFlag := store.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.DeleteOwn, ouroboros.DeleteAll)

	if active.cell.Genoma&requiredFlag == 0 {
		return errorWithID(parser, req.ID, protocol.ErrorCodePermissionDenied, "delete.permission")
	}

	if err := store.deleteMembrane(string(req.Key)); err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMembraneDeleteFailed, "delete.remove")
	}

	return parser.DeleteResultBytes(req.ID, 1)
}
