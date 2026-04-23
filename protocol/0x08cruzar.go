package protocol

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

	rawCruzarReqMsg := []byte{
		0x00, 0x00, 0x00, 0x07, // Opcode
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // ID
		0x00, 0x00, 0x00, 0x04, // DB Name Len
		0x00, 0x00, 0x00, 0x01, // CellIndex
		0x00, 0x00, 0x00, 0x04, // ParentSecret Len
		0x00, 0x00, 0x00, 0x02, // ChildGenome
		0x00, 0x00, 0x00, 0x03, // X
		0x00, 0x00, 0x00, 0x04, // Y
		0x00, 0x00, 0x00, 0x05, // Z
		0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x09, // ChildSalt
		0x00, 0x00, 0x00, 0x04, // ChildSecret Len
		0x0A, 0x0B, 0x0C, 0x0D, // DB Name
		0x0E, 0x0F, 0x10, 0x11, // ParentSecret
		0x12, 0x13, 0x14, 0x15, // ChildSecret
	}

	readCruzarReq, err := parser.CruzarReq(rawCruzarReqMsg)
	if err != nil {
		fmt.Printf("Error al parsear CruzarReq: %v\n", err)
	} else {
		fmt.Println("CruzarReq parseado correctamente")
	}

	fmt.Printf("Opcode: 0x07 (Cruzar)\n")
	fmt.Printf("Cruzar Req ID: %s\n", string(readCruzarReq.ID))
	fmt.Printf("CellIndexA: %d\n", readCruzarReq.CellIndexA)
	fmt.Printf("CellIndexB: %d\n", readCruzarReq.CellIndexB)
	fmt.Printf("X: %d\n", readCruzarReq.X)
	fmt.Printf("Y: %d\n", readCruzarReq.Y)
	fmt.Printf("Z: %d\n", readCruzarReq.Z)
	fmt.Printf("DBName: %s\n", string(readCruzarReq.DBName))
	fmt.Printf("SecretA: %s\n", string(readCruzarReq.SecretA))
	fmt.Printf("SecretB: %s\n", string(readCruzarReq.SecretB))
	fmt.Printf("ChildSecret: %s\n", string(readCruzarReq.ChildSecret))
	fmt.Println("--------------------------------------------------")

	rawCruzarResultMsg := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // ID
		0x00, 0x00, 0x00, 0x01, // Status
		0x00, 0x00, 0x00, 0x2A, // CellIndex
	}

	readCruzarResult, err := parser.CruzarResult(rawCruzarResultMsg)
	if err != nil {
		fmt.Printf("Error al parsear CruzarResult: %v\n", err)
	} else {
		fmt.Println("CruzarResult parseado correctamente")
	}

	fmt.Printf("Opcode: 0x07 (Cruzar Result)\n")
	fmt.Printf("Cruzar Result ID: %s\n", string(readCruzarResult.ID))
	fmt.Printf("Status: %d\n", readCruzarResult.Status)
	fmt.Printf("CellIndex: %d\n", readCruzarResult.CellIndex)
	fmt.Println("--------------------------------------------------")

}
