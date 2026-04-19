package main

import (
	"encoding/binary"
	"fmt"
)

/*CREATE_DB (0x08)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[Secret Len: 4]
[Genesis DB len: 4]
[Genesis index: 4] |
[DB Name: N]
[Secret: M]
[Genesis DB: P]*/

type CreateDBReqMessage struct {
	ID           []byte
	DBName       []byte
	Secret       []byte
	GenesisDB    []byte
	GenesisIndex uint32
}

func (p *ProtocolParser) CreateDBReq(msg []byte) (CreateDBReqMessage, error) {
	var cm CreateDBReqMessage

	// Opcode(4) + ID(16) + DBLen(4) + SecretLen(4) + GenesisDBLen(4) + GenesisIndex(4) = 36 bytes
	if len(msg) < 36 {
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

	genesisDBLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	cm.GenesisIndex = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + secretLen + genesisDBLen)
	if len(msg) < offset+totalVariableLength {
		return cm, fmt.Errorf("mensaje incompleto")
	}

	cm.DBName = make([]byte, dbNameLen)
	copy(cm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	cm.Secret = make([]byte, secretLen)
	copy(cm.Secret, msg[offset:offset+int(secretLen)])
	offset += int(secretLen)

	cm.GenesisDB = make([]byte, genesisDBLen)
	copy(cm.GenesisDB, msg[offset:offset+int(genesisDBLen)])

	return cm, nil
}

/*CREATE_DB result
[ID: 16]
[Status: 4]*/

type CreateDBResult struct {
	ID     []byte
	Status int32
}

func (p *ProtocolParser) CreateDBResult(msg []byte) (CreateDBResult, error) {
	var cr CreateDBResult

	if len(msg) < 20 {
		return cr, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0

	cr.ID = make([]byte, 16)
	copy(cr.ID, msg[offset:offset+16])
	offset += 16

	cr.Status = int32(binary.BigEndian.Uint32(msg[offset : offset+4]))

	return cr, nil
}
