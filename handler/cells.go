package samsara

import (
	"github.com/DiegoSandival/ouroboros-go"
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

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

	activeA, ok := store.resolveCell(req.CellIndexA, req.SecretA)
	if !ok {
		return parser.CruzarResultBytes(req.ID, 4, 0)
	}

	activeB, ok := store.resolveCell(req.CellIndexB, req.SecretB)
	if !ok {
		return parser.CruzarResultBytes(req.ID, 5, 0)
	}

	//no se pueden cruzar si alguno de los dos no tiene la capacidad de fucionar (bit de fucionar en el genoma)
	if activeA.cell.Genoma&ouroboros.Merge == 0 || activeB.cell.Genoma&ouroboros.Merge == 0 {

		return parser.CruzarResultBytes(req.ID, 6, 0)
	}

	childGenome := activeA.cell.Genoma | activeB.cell.Genoma
	childCell := NewCellWithSecret([16]byte{}, req.ChildSecret, childGenome, req.X, req.Y, req.Z)

	childIndex, _ := store.DB().Append(childCell)

	return parser.CruzarResultBytes(req.ID, 1, childIndex)
}
