package main

import (
	"encoding/binary"
	"fmt"
)

/*DIFERIR (0x06)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[CellIndex: 4]
[ParentSecret Len: 4]
[ChildGenome: 4]
[X: 4]
[Y: 4]
[Z: 4]
[ChildSalt: 16]
[ChildSecret Len: 4] |
[DB Name: N]
[ParentSecret: M]
[ChildSecret: P]*/

type DiferirReqMessage struct {
	ID           []byte
	DBName       []byte
	CellIndex    uint32
	ParentSecret []byte
	ChildGenome  uint32
	X            uint32
	Y            uint32
	Z            uint32
	ChildSalt    [16]byte
	ChildSecret  []byte
}

func (p *ProtocolParser) DiferirReq(msg []byte) (DiferirReqMessage, error) {
	var dm DiferirReqMessage

	// Opcode(4) + ID(16) + DBLen(4) + CellIndex(4) + ParentSecretLen(4) + ChildGenome(4) + X(4) + Y(4) + Z(4) + ChildSalt(16) + ChildSecretLen(4) = 68 bytes
	if len(msg) < 68 {
		return dm, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0
	offset += 4 // opcode

	dm.ID = make([]byte, 16)
	copy(dm.ID, msg[offset:offset+16])
	offset += 16

	dbNameLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	dm.CellIndex = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	parentSecretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	dm.ChildGenome = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	dm.X = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	dm.Y = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	dm.Z = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	copy(dm.ChildSalt[:], msg[offset:offset+16])
	offset += 16

	childSecretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + parentSecretLen + childSecretLen)
	if len(msg) < offset+totalVariableLength {
		return dm, fmt.Errorf("mensaje incompleto")
	}

	dm.DBName = make([]byte, dbNameLen)
	copy(dm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	dm.ParentSecret = make([]byte, parentSecretLen)
	copy(dm.ParentSecret, msg[offset:offset+int(parentSecretLen)])
	offset += int(parentSecretLen)

	dm.ChildSecret = make([]byte, childSecretLen)
	copy(dm.ChildSecret, msg[offset:offset+int(childSecretLen)])

	return dm, nil
}

/*DIFERIR result
[ID: 16]
[Status: 4]
[CellIndex: 4]*/

type DiferirResult struct {
	ID        []byte
	Status    int32
	CellIndex uint32
}

func (p *ProtocolParser) DiferirResult(msg []byte) (DiferirResult, error) {
	var dr DiferirResult

	if len(msg) < 24 {
		return dr, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0

	dr.ID = make([]byte, 16)
	copy(dr.ID, msg[offset:offset+16])
	offset += 16

	dr.Status = int32(binary.BigEndian.Uint32(msg[offset : offset+4]))
	offset += 4

	dr.CellIndex = binary.BigEndian.Uint32(msg[offset : offset+4])

	return dr, nil
}

func (parser *ProtocolParser) testDiferir() {

	rawDiferirReqMsg := []byte{
		0x00, 0x00, 0x00, 0x06, // Opcode
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

	readDiferirReq, err := parser.DiferirReq(rawDiferirReqMsg)
	if err != nil {
		fmt.Printf("Error al parsear DiferirReq: %v\n", err)
	} else {
		fmt.Println("DiferirReq parseado correctamente")
	}

	fmt.Printf("Opcode: 0x06 (Diferir)\n")
	fmt.Printf("Read Cell Req ID: %s\n", string(readDiferirReq.ID))
	fmt.Printf("CellIndex: %d\n", readDiferirReq.CellIndex)
	fmt.Printf("DBName: %s\n", string(readDiferirReq.DBName))
	fmt.Printf("ParentSecret: %s\n", string(readDiferirReq.ParentSecret))
	fmt.Printf("ChildSecret: %s\n", string(readDiferirReq.ChildSecret))
	fmt.Println("--------------------------------------------------")

	rawDiferirResultMsg := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // ID
		0x00, 0x00, 0x00, 0x01, // Status
		0x00, 0x00, 0x00, 0x2A, // CellIndex
	}

	readDiferirResult, err := parser.DiferirResult(rawDiferirResultMsg)
	if err != nil {
		fmt.Printf("Error al parsear DiferirResult: %v\n", err)
	} else {
		fmt.Println("DiferirResult parseado correctamente")
	}

	fmt.Printf("Opcode: 0x06 (Diferir Result)\n")
	fmt.Printf("Diferir Result ID: %s\n", string(readDiferirResult.ID))
	fmt.Printf("Status: %d\n", readDiferirResult.Status)
	fmt.Printf("CellIndex: %d\n", readDiferirResult.CellIndex)
	fmt.Println("--------------------------------------------------")
}
