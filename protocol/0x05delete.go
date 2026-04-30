package protocol

import (
	"encoding/binary"
	"fmt"
)

/*DELETE (0x25)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Key Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Key: M]
[Secret: P]*/

type DeleteReqMessage struct {
	ID        []byte
	CellIndex uint32
	DBName    []byte
	Key       []byte
	Secret    []byte
}

func (p *ProtocolParser) DeleteReqBytes(dbName, key, secret []byte, cellIndex uint32) []byte {
	dbNameLen := uint32(len(dbName))
	keyLen := uint32(len(key))
	secretLen := uint32(len(secret))
	totalLen := 4 + 16 + 4 + 4 + 4 + 4 + dbNameLen + keyLen + secretLen

	msg := make([]byte, totalLen)
	offset := 0
	// Opcode
	binary.BigEndian.PutUint32(msg[offset:offset+4], 0x25)
	offset += 4
	// ID (16 bytes)
	copy(msg[offset:offset+16], make([]byte, 16)) // ID vacío
	offset += 16
	// CellIndex
	binary.BigEndian.PutUint32(msg[offset:offset+4], cellIndex)
	offset += 4
	// DB Name Len
	binary.BigEndian.PutUint32(msg[offset:offset+4], dbNameLen)
	offset += 4
	// Key Len
	binary.BigEndian.PutUint32(msg[offset:offset+4], keyLen)
	offset += 4
	// Secret Len
	binary.BigEndian.PutUint32(msg[offset:offset+4], secretLen)
	offset += 4
	// DB Name
	copy(msg[offset:offset+int(dbNameLen)], dbName)
	offset += int(dbNameLen)
	// Key
	copy(msg[offset:offset+int(keyLen)], key)
	offset += int(keyLen)
	// Secret
	copy(msg[offset:offset+int(secretLen)], secret)
	offset += int(secretLen)

	return msg
}

func (p *ProtocolParser) DeleteReq(msg []byte) (DeleteReqMessage, error) {
	var dm DeleteReqMessage

	// Opcode(4) + ID(16) + CellIndex(4) + DBLen(4) + KeyLen(4) + SecretLen(4) = 36 bytes
	if len(msg) < 36 {
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

	keyLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	secretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + keyLen + secretLen)
	if len(msg) < offset+totalVariableLength {
		return dm, fmt.Errorf("mensaje incompleto")
	}

	dm.DBName = make([]byte, dbNameLen)
	copy(dm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	dm.Key = make([]byte, keyLen)
	copy(dm.Key, msg[offset:offset+int(keyLen)])
	offset += int(keyLen)

	dm.Secret = make([]byte, secretLen)
	copy(dm.Secret, msg[offset:offset+int(secretLen)])

	return dm, nil
}

/*DELETE result
[ID: 16]
[Status: 4]*/

type DeleteResult struct {
	ID     []byte
	Status int32
}

func (p *ProtocolParser) DeleteResultBytes(id []byte, status int32) []byte {
	msg := make([]byte, 16+4)
	offset := 0
	// ID
	copy(msg[offset:offset+16], id)
	offset += 16
	// Status
	binary.BigEndian.PutUint32(msg[offset:offset+4], uint32(status))
	return msg
}

func (p *ProtocolParser) DeleteResult(msg []byte) (DeleteResult, error) {
	var dr DeleteResult

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

func (parser *ProtocolParser) testDelete() {

	rawDeleteReqMsg := []byte{
		// Opcode: 0x25
		0x00, 0x00, 0x00, 0x25,
		// ID: 16 bytes (16 letras 'G')
		0x47, 0x47, 0x47, 0x47, 0x47, 0x47, 0x47, 0x47,
		0x47, 0x47, 0x47, 0x47, 0x47, 0x47, 0x47, 0x47,
		// CellIndex: 42
		0x00, 0x00, 0x00, 0x2A,
		// DB Name Len: 4
		0x00, 0x00, 0x00, 0x04,
		// Key Len: 7
		0x00, 0x00, 0x00, 0x07,
		// Secret Len: 6
		0x00, 0x00, 0x00, 0x06,
		// DB Name: "test"
		0x74, 0x65, 0x73, 0x74,
		// Key: "mykey01"
		0x6D, 0x79, 0x6B, 0x65, 0x79, 0x30, 0x31,
		// Secret: "secret"
		0x73, 0x65, 0x63, 0x72, 0x65, 0x74,
	}

	deleteReq, err := parser.DeleteReq(rawDeleteReqMsg)
	if err != nil {
		fmt.Printf("Error parsing delete request: %v\n", err)
		return
	}

	// Verificación del resultado de eliminación
	fmt.Printf("Opcode: 0x25 (DELETE)\n")
	fmt.Printf("Delete Req ID: %s\n", string(deleteReq.ID))
	fmt.Printf("CellIndex: %d\n", deleteReq.CellIndex)
	fmt.Printf("DBName: %s\n", string(deleteReq.DBName))
	fmt.Printf("Key: %s\n", string(deleteReq.Key))
	fmt.Printf("Secret: %s\n", string(deleteReq.Secret))
	fmt.Println("--------------------------------------------------")

	rawDeleteResultMsg := []byte{
		// ID: 16 bytes (16 letras 'H')
		0x48, 0x48, 0x48, 0x48, 0x48, 0x48, 0x48, 0x48,
		0x48, 0x48, 0x48, 0x48, 0x48, 0x48, 0x48, 0x48,
		// Status: 1 (éxito)
		0x00, 0x00, 0x00, 0x01,
	}

	deleteResult, err := parser.DeleteResult(rawDeleteResultMsg)
	if err != nil {
		fmt.Printf("Error parsing delete result: %v\n", err)
		return
	}

	// Verificación del resultado de eliminación
	fmt.Printf("Opcode: 0x25 (DELETE Result)\n")
	fmt.Printf("Delete Result ID: %s\n", string(deleteResult.ID))
	fmt.Printf("Status: %d\n", deleteResult.Status)
	fmt.Println("--------------------------------------------------")
}
