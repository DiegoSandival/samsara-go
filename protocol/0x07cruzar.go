package main

import (
	"encoding/binary"
	"fmt"
)

/*CRUZAR (0x07)
[Opcode: 4]
[ID: 16]
[CellIndexA: 4]
[CellIndexB: 4]
[X: 4]
[Y: 4]
[Z: 4]
[ChildSalt: 16]
[DB Name Len: 4]
[SecretA Len: 4]
[SecretB Len: 4]
[ChildSecret Len: 4] |
[DB Name: N]
[SecretA: M]
[SecretB: P]
[ChildSecret: Q]*/

type CruzarReqMessage struct {
	ID          []byte
	CellIndexA  uint32
	CellIndexB  uint32
	X           uint32
	Y           uint32
	Z           uint32
	ChildSalt   [16]byte
	DBName      []byte
	SecretA     []byte
	SecretB     []byte
	ChildSecret []byte
}

func (p *ProtocolParser) CruzarReq(msg []byte) (CruzarReqMessage, error) {
	var cm CruzarReqMessage

	// Opcode(4) + ID(16) + CellIndexA(4) + CellIndexB(4) + X(4) + Y(4) + Z(4) + ChildSalt(16) + DBLen(4) + SecretALen(4) + SecretBLen(4) + ChildSecretLen(4) = 60 bytes
	if len(msg) < 60 {
		return cm, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0
	offset += 4 // opcode

	cm.ID = make([]byte, 16)
	copy(cm.ID, msg[offset:offset+16])
	offset += 16

	cm.CellIndexA = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	cm.CellIndexB = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	cm.X = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	cm.Y = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	cm.Z = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	copy(cm.ChildSalt[:], msg[offset:offset+16])
	offset += 16

	dbNameLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	secretALen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	secretBLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	childSecretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + secretALen + secretBLen + childSecretLen)
	if len(msg) < offset+totalVariableLength {
		return cm, fmt.Errorf("mensaje incompleto")
	}

	cm.DBName = make([]byte, dbNameLen)
	copy(cm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	cm.SecretA = make([]byte, secretALen)
	copy(cm.SecretA, msg[offset:offset+int(secretALen)])
	offset += int(secretALen)

	cm.SecretB = make([]byte, secretBLen)
	copy(cm.SecretB, msg[offset:offset+int(secretBLen)])
	offset += int(secretBLen)

	cm.ChildSecret = make([]byte, childSecretLen)
	copy(cm.ChildSecret, msg[offset:offset+int(childSecretLen)])

	return cm, nil
}

/*CRUZAR result
[ID: 16]
[Status: 4]
[CellIndex: 4]*/

type CruzarResult struct {
	ID        []byte
	Status    int32
	CellIndex uint32
}

func (p *ProtocolParser) CruzarResult(msg []byte) (CruzarResult, error) {
	var cr CruzarResult

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

func (parser *ProtocolParser) testCruzar() {
	// Aquí puedes agregar un mensaje de prueba para CRUZAR y verificar su análisis
}

func (parser *ProtocolParser) testCreateDB() {
	// Aquí puedes agregar un mensaje de prueba para CREATE_DB y verificar su análisis
}
