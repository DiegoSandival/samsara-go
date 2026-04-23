package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

/*DELETE_DB (0x01)
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

func (p *ProtocolParser) DeleteDBReqBytes(dbName string, secret string, cellIndex uint32) []byte {
	dbNameBytes := []byte(dbName)
	secretBytes := []byte(secret)
	dbNameLen := uint32(len(dbNameBytes))
	secretLen := uint32(len(secretBytes))
	msg := make([]byte, 4+16+4+4+4+len(dbNameBytes)+len(secretBytes))
	offset := 0
	// Opcode
	binary.BigEndian.PutUint32(msg[offset:offset+4], 0x01)
	offset += 4
	// ID (16 bytes aleatorios)
	randomID := make([]byte, 16)
	_, err := rand.Read(randomID)
	if err != nil {
		// En caso de error, podemos usar un ID fijo o manejarlo de otra forma
		copy(randomID, []byte("default-id-1234"))
	}
	copy(msg[offset:offset+16], randomID)
	offset += 16
	// CellIndex
	binary.BigEndian.PutUint32(msg[offset:offset+4], cellIndex)
	offset += 4
	// DB Name Len
	binary.BigEndian.PutUint32(msg[offset:offset+4], dbNameLen)
	offset += 4
	// Secret Len
	binary.BigEndian.PutUint32(msg[offset:offset+4], secretLen)
	offset += 4
	// DB Name
	copy(msg[offset:offset+int(dbNameLen)], dbNameBytes)
	offset += int(dbNameLen)
	// Secret
	copy(msg[offset:offset+int(secretLen)], secretBytes)
	offset += int(secretLen)

	return msg
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

func (p *ProtocolParser) DeleteDBResultBytes(id []byte, status int32) []byte {
	msg := make([]byte, 20)
	offset := 0

	copy(msg[offset:offset+16], id)
	offset += 16

	binary.BigEndian.PutUint32(msg[offset:offset+4], uint32(status))

	return msg
}

func (parser *ProtocolParser) testDeleteDB() {
	rawDeleteDBReqMsg := []byte{
		0x00, 0x00, 0x00, 0x01, // Opcode
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // ID
		0x00, 0x00, 0x00, 0x01, // CellIndex
		0x00, 0x00, 0x00, 0x04, // DB Name Len
		0x00, 0x00, 0x00, 0x02, // Secret Len
		0x00, 0x00, 0x00, 0x03, // X
		0x00, 0x00, 0x00, 0x04, // Y
		0x00, 0x00, 0x00, 0x05, // Z
		0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x09, // ChildSalt
		0x00, 0x00, 0x00, 0x04, // ChildSecret Len
		0x0A, 0x0B, 0x0C, 0x0D, // DB Name
		0x0E, 0x0F, 0x10, 0x11, // Secret
		0x12, 0x13, 0x14, 0x15, // ChildSecret
	}

	readDeleteDBReq, err := parser.DeleteDBReq(rawDeleteDBReqMsg)
	if err != nil {
		fmt.Printf("Error al parsear DeleteDBReq: %v\n", err)
	} else {
		fmt.Println("DeleteDBReq parseado correctamente")
	}

	fmt.Printf("Opcode: 0x01 (DeleteDB)\n")
	fmt.Printf("DeleteDB Req ID: %s\n", string(readDeleteDBReq.ID))
	fmt.Printf("CellIndex: %d\n", readDeleteDBReq.CellIndex)
	fmt.Printf("DBName: %s\n", string(readDeleteDBReq.DBName))
	fmt.Printf("Secret: %s\n", string(readDeleteDBReq.Secret))
	fmt.Println("--------------------------------------------------")

	rawDeleteDBResultMsg := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // ID
		0x00, 0x00, 0x00, 0x01, // Status
		0x00, 0x00, 0x00, 0x2A, // CellIndex
	}

	readDeleteDBResult, err := parser.DeleteDBResult(rawDeleteDBResultMsg)
	if err != nil {
		fmt.Printf("Error al parsear DeleteDBResult: %v\n", err)
	} else {
		fmt.Println("DeleteDBResult parseado correctamente")
	}

	fmt.Printf("Opcode: 0x01 (DeleteDB Result)\n")
	fmt.Printf("DeleteDB Result ID: %s\n", string(readDeleteDBResult.ID))
	fmt.Printf("Status: %d\n", readDeleteDBResult.Status)
	fmt.Println("--------------------------------------------------")
}
