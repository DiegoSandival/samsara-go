package main

import (
	"fmt"
)

/*
READ (0x01)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Key Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Key: M]
[Secret: P]
*/
type OpcodeReqMessage struct {
	Opcode []byte
}

func (p *ProtocolParser) Opcode(msg []byte) (OpcodeReqMessage, error) {
	var rm OpcodeReqMessage

	// retornamos el opcode (4 bytes)
	if len(msg) < 4 {
		return rm, fmt.Errorf("mensaje demasiado corto para el opcode")
	}
	rm.Opcode = make([]byte, 4)
	copy(rm.Opcode, msg[:4])

	return rm, nil
}
