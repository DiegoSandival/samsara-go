package main

import (
	"encoding/binary"
	"fmt"
)

/*DELETE_DB (0x09)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Secret: M]
*/

type DeleteDBReqMessage struct {
	ID        []byte
	CellIndex uint32
	DBName    []byte
	Secret    []byte
}

func (p *ProtocolParser) DeleteDBReq(msg []byte) (DeleteDBReqMessage, error) {
	var dm DeleteDBReqMessage

	// Opcode(4) + ID(16) + CellIndex(4) + DBLen(4) + SecretLen(4) = 32 bytes
	if len(msg) < 32 {
		return dm, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0
	offset += 4 // opcode

	dm.ID = make([]byte, 16)
	copy(dm.ID, msg[offset:offset+16])
	offset += 16

	dm.CellIndex = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	dbNameLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	secretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + secretLen)
	if len(msg) < offset+totalVariableLength {
		return dm, fmt.Errorf("mensaje incompleto")
	}

	dm.DBName = make([]byte, dbNameLen)
	copy(dm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	dm.Secret = make([]byte, secretLen)
	copy(dm.Secret, msg[offset:offset+int(secretLen)])

	return dm, nil
}

/*DELETE_DB result
[ID: 16]
[Status: 4]*/

type DeleteDBResult struct {
	ID     []byte
	Status int32
}

func (p *ProtocolParser) DeleteDBResult(msg []byte) (DeleteDBResult, error) {
	var dr DeleteDBResult

	if len(msg) < 20 {
		return dr, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0

	dr.ID = make([]byte, 16)
	copy(dr.ID, msg[offset:offset+16])
	offset += 16

	dr.Status = int32(binary.BigEndian.Uint32(msg[offset : offset+4]))

	return dr, nil
}

func (parser *ProtocolParser) testDeleteDB() {
	// Aquí puedes agregar un mensaje de prueba para DELETE_DB y verificar su análisis
}
