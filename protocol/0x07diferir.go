package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

/*DIFERIR (0x27)
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
	ID          []byte
	DBName      []byte
	CellIndex   uint32
	Secret      []byte
	ChildGenome uint32
	X           uint32
	Y           uint32
	Z           uint32
	ChildSecret []byte
}

func (p *ProtocolParser) DiferirReq(msg []byte) (DiferirReqMessage, error) {
	var dr DiferirReqMessage

	if len(msg) < 68 {
		return dr, fmt.Errorf("mensaje demasiado corto")
	}
	offset := 0
	offset += 4 // opcode

	dr.ID = make([]byte, 16)
	copy(dr.ID, msg[offset:offset+16])
	offset += 16

	dbNameLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4
	dr.CellIndex = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4
	secretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4
	dr.ChildGenome = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4
	dr.X = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4
	dr.Y = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4
	dr.Z = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4
	if len(msg) < offset+16 {
		return dr, fmt.Errorf("mensaje demasiado corto para ChildSalt")
	}
	offset += 16
	childSecretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4
	totalVariableLength := int(dbNameLen + secretLen + childSecretLen)
	if len(msg) < offset+totalVariableLength {
		return dr, fmt.Errorf("mensaje incompleto")
	}

	dr.DBName = make([]byte, dbNameLen)
	copy(dr.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	dr.Secret = make([]byte, secretLen)
	copy(dr.Secret, msg[offset:offset+int(secretLen)])
	offset += int(secretLen)

	dr.ChildSecret = make([]byte, childSecretLen)
	copy(dr.ChildSecret, msg[offset:offset+int(childSecretLen)])

	return dr, nil
}

func (p *ProtocolParser) DiferirReqBytes(DBName, Secret, ChildSecret []byte, CellIndex uint32, ChildGenome uint32, X, Y, Z uint32) []byte {
	ID := make([]byte, 16)
	rand.Read(ID)
	dbNameLen := uint32(len(DBName))
	secretLen := uint32(len(Secret))
	childSecretLen := uint32(len(ChildSecret))
	totalLen := 4 + 16 + 4 + 4 + 4 + 4 + 4 + 4 + 4 + 16 + 4 + dbNameLen + secretLen + childSecretLen
	msg := make([]byte, totalLen)
	offset := 0
	binary.BigEndian.PutUint32(msg[offset:offset+4], 0x27) // Opcode
	offset += 4

	copy(msg[offset:offset+16], ID)
	offset += 16
	binary.BigEndian.PutUint32(msg[offset:offset+4], dbNameLen)
	offset += 4
	binary.BigEndian.PutUint32(msg[offset:offset+4], CellIndex)
	offset += 4
	binary.BigEndian.PutUint32(msg[offset:offset+4], secretLen)
	offset += 4
	binary.BigEndian.PutUint32(msg[offset:offset+4], ChildGenome)
	offset += 4
	binary.BigEndian.PutUint32(msg[offset:offset+4], X)
	offset += 4
	binary.BigEndian.PutUint32(msg[offset:offset+4], Y)
	offset += 4
	binary.BigEndian.PutUint32(msg[offset:offset+4], Z)
	offset += 4
	childSalt := make([]byte, 16)
	rand.Read(childSalt)
	copy(msg[offset:offset+16], childSalt)
	offset += 16
	binary.BigEndian.PutUint32(msg[offset:offset+4], childSecretLen)
	offset += 4
	copy(msg[offset:offset+int(dbNameLen)], DBName)
	offset += int(dbNameLen)
	copy(msg[offset:offset+int(secretLen)], Secret)
	offset += int(secretLen)
	copy(msg[offset:offset+int(childSecretLen)], ChildSecret)
	offset += int(childSecretLen)

	return msg
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

func (p *ProtocolParser) DiferirResultBytes(ID []byte, status int32, cellIndex uint32) []byte {
	totalLen := 16 + 4 + 4
	msg := make([]byte, totalLen)
	offset := 0

	copy(msg[offset:offset+16], ID)
	offset += 16

	binary.BigEndian.PutUint32(msg[offset:offset+4], uint32(status))
	offset += 4

	binary.BigEndian.PutUint32(msg[offset:offset+4], cellIndex)

	return msg
}

func (parser *ProtocolParser) testDiferir() {

	rawDiferirReqMsg := []byte{
		0x00, 0x00, 0x00, 0x27, // Opcode
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

	fmt.Printf("Opcode: 0x27 (Diferir)\n")
	fmt.Printf("Read Cell Req ID: %s\n", string(readDiferirReq.ID))
	fmt.Printf("CellIndex: %d\n", readDiferirReq.CellIndex)
	fmt.Printf("DBName: %s\n", string(readDiferirReq.DBName))
	fmt.Printf("Secret: %s\n", string(readDiferirReq.Secret))
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

	fmt.Printf("Opcode: 0x27 (Diferir Result)\n")
	fmt.Printf("Diferir Result ID: %s\n", string(readDiferirResult.ID))
	fmt.Printf("Status: %d\n", readDiferirResult.Status)
	fmt.Printf("CellIndex: %d\n", readDiferirResult.CellIndex)
	fmt.Println("--------------------------------------------------")
}
