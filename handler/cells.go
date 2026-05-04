package samsara

import (
	"github.com/DiegoSandival/ouroboros-go"
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

func (s *CentralHandler) ReadCell(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.ReadCellReq(payload)
	if err != nil {
		return requestParseError(parser, payload, protocol.OpcodeReadCell, err, "read_cell.parse")
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeDatabaseNotFound, "read_cell.store")
	}

	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return errorWithID(parser, req.ID, protocol.ErrorCodeAuthenticationFailed, "read_cell.auth")
	}

	return parser.ReadCellResultBytes(req.ID, 1, parser.CellBytes(active.cell))
}

func (s *CentralHandler) Diferir(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.DiferirReq(payload)
	if err != nil {
		return requestParseError(parser, payload, protocol.OpcodeDiferir, err, "diferir.parse")
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeDatabaseNotFound, "diferir.store")
	}

	active, ok := store.resolveCell(req.CellIndex, req.Secret)
	if !ok {
		return errorWithID(parser, req.ID, protocol.ErrorCodeAuthenticationFailed, "diferir.auth")
	}

	if active.cell.Genoma&req.ChildGenome != req.ChildGenome {
		return errorWithID(parser, req.ID, protocol.ErrorCodeChildGenomeEscalation, "diferir.genome")
	}

	childCell := store.NewCellWithSecret([16]byte{}, req.ChildSecret, req.ChildGenome, req.X, req.Y, req.Z)

	childIndex, err := store.DB().Append(childCell)
	if err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeCellAppendFailed, "diferir.append")
	}

	return parser.DiferirResultBytes(req.ID, 1, childIndex)
}

func (s *CentralHandler) Cruzar(parser *protocol.ProtocolParser, payload []byte) []byte {
	req, err := parser.CruzarReq(payload)
	if err != nil {
		return requestParseError(parser, payload, protocol.OpcodeCruzar, err, "cruzar.parse")
	}

	store, exists := s.GetStore(string(req.DBName))
	if !exists {
		return errorWithID(parser, req.ID, protocol.ErrorCodeDatabaseNotFound, "cruzar.store")
	}

	activeA, ok := store.resolveCell(req.CellIndexA, req.SecretA)
	if !ok {
		return errorWithID(parser, req.ID, protocol.ErrorCodeAuthenticationFailed, "cruzar.auth_a")
	}

	activeB, ok := store.resolveCell(req.CellIndexB, req.SecretB)
	if !ok {
		return errorWithID(parser, req.ID, protocol.ErrorCodeAuthenticationFailed, "cruzar.auth_b")
	}

	//no se pueden cruzar si alguno de los dos no tiene la capacidad de fucionar (bit de fucionar en el genoma)
	if activeA.cell.Genoma&ouroboros.Merge == 0 || activeB.cell.Genoma&ouroboros.Merge == 0 {
		return errorWithID(parser, req.ID, protocol.ErrorCodeMergeCapabilityMissing, "cruzar.merge")
	}

	childGenome := activeA.cell.Genoma | activeB.cell.Genoma
	childCell := NewCellWithSecret([16]byte{}, req.ChildSecret, childGenome, req.X, req.Y, req.Z)

	childIndex, err := store.DB().Append(childCell)
	if err != nil {
		return errorWithID(parser, req.ID, protocol.ErrorCodeCellAppendFailed, "cruzar.append")
	}

	return parser.CruzarResultBytes(req.ID, 1, childIndex)
}
