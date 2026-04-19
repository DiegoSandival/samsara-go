package main

import (
	"encoding/binary"
	"fmt"
)

/*CREATE_GENESIS (0x010)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[Secret Len: 4]
[Size: 4] |
[DB Name: N]
[Secret: M]*/

type CreateGenesisReqMessage struct {
	ID     []byte
	DBName []byte
	Secret []byte
	Size   uint32
}

func (p *ProtocolParser) CreateGenesisReq(msg []byte) (CreateGenesisReqMessage, error) {
	var cm CreateGenesisReqMessage

	// Opcode(4) + ID(16) + DBLen(4) + SecretLen(4) + Size(4) = 32 bytes
	if len(msg) < 32 {
		return cm, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0
	offset += 4 // opcode

	cm.ID = make([]byte, 16)
	copy(cm.ID, msg[offset:offset+16])
	offset += 16

	dbNameLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	secretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	cm.Size = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + secretLen)
	if len(msg) < offset+totalVariableLength {
		return cm, fmt.Errorf("mensaje incompleto")
	}

	cm.DBName = make([]byte, dbNameLen)
	copy(cm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	cm.Secret = make([]byte, secretLen)
	copy(cm.Secret, msg[offset:offset+int(secretLen)])

	return cm, nil
}

/*CREATE_GENESIS result
[ID: 16]
[Status: 4]
[CellIndex: 4]*/

type CreateGenesisResult struct {
	ID        []byte
	Status    int32
	CellIndex uint32
}

func (p *ProtocolParser) CreateGenesisResult(msg []byte) (CreateGenesisResult, error) {
	var cr CreateGenesisResult

	if len(msg) < 24 {
		return cr, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0

	cr.ID = make([]byte, 16)
	copy(cr.ID, msg[offset:offset+16])
	offset += 16

	cr.Status = int32(binary.BigEndian.Uint32(msg[offset : offset+4]))
	offset += 4

	cr.CellIndex = binary.BigEndian.Uint32(msg[offset : offset+4])

	return cr, nil
}

func (parser *ProtocolParser) testCreateGenesis() {
	// Aquí puedes agregar un mensaje de prueba para CREATE_GENESIS y verificar su análisis
}
