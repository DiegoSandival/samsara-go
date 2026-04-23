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
