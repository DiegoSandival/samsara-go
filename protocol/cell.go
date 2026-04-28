package protocol

import (
	"encoding/binary"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
)

func (p *ProtocolParser) CellBytes(cell ouroboros.Celula) []byte {

	msg := make([]byte, 16)
	// Genoma
	binary.BigEndian.PutUint32(msg[0:4], cell.Genoma)
	// X
	binary.BigEndian.PutUint32(msg[4:8], cell.X)
	// Y
	binary.BigEndian.PutUint32(msg[8:12], cell.Y)
	// Z
	binary.BigEndian.PutUint32(msg[12:16], cell.Z)
	return msg
}

func (p *ProtocolParser) CellFromBytes(data []byte) ouroboros.Celula {

	cell := ouroboros.Celula{}
	// Genoma
	cell.Genoma = binary.BigEndian.Uint32(data[0:4])
	// X
	cell.X = binary.BigEndian.Uint32(data[4:8])
	// Y
	cell.Y = binary.BigEndian.Uint32(data[8:12])
	// Z
	cell.Z = binary.BigEndian.Uint32(data[12:16])
	return cell
}

func (p *ProtocolParser) GenomaFromBytes(data []byte) uint32 {
	// Genoma
	genoma := binary.BigEndian.Uint32(data[0:4])
	return genoma
}

type GenomaDetail struct {
	ReadSelf        bool
	ReadNeighbors   bool
	WriteSelf       bool
	WriteNeighbors  bool
	DeleteSelf      bool
	DeleteNeighbors bool
	DiferirSelf     bool
	CruzarSelf      bool
	DominanceSelf   bool
	FreeRead        bool
	IsMigrated      bool
}

func (p *ProtocolParser) GenomaDetailFromBytes(genoma GenomaDetail) uint32 {
	var result uint32 = 0
	if genoma.ReadSelf {
		result |= 1 << 0
	}
	if genoma.ReadNeighbors {
		result |= 1 << 1
	}
	if genoma.WriteSelf {
		result |= 1 << 2
	}
	if genoma.WriteNeighbors {
		result |= 1 << 3
	}
	if genoma.DeleteSelf {
		result |= 1 << 4
	}
	if genoma.DeleteNeighbors {
		result |= 1 << 5
	}
	if genoma.DiferirSelf {
		result |= 1 << 6
	}
	if genoma.CruzarSelf {
		result |= 1 << 7
	}
	if genoma.DominanceSelf {
		result |= 1 << 9
	}
	if genoma.FreeRead {
		result |= 1 << 10
	}
	if genoma.IsMigrated {
		result |= 1 << 11
	}

	//llenar de bit 12 a 30 con 0s

	for i := 12; i <= 31; i++ {
		result &= ^(1 << i)
	}

	return result
}
