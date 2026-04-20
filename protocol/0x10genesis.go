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
