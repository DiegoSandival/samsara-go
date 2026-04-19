package main

import (
	"encoding/binary"
	"fmt"
)

/*WRITE (0x03)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Key Len: 4]
[Value Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Key: M]
[Value: P]
[Secret: Q]*/

type WriteReqMessage struct {
	ID        []byte
	CellIndex uint32
	DBName    []byte
	Key       []byte
	Value     []byte
	Secret    []byte
}

func (p *ProtocolParser) WriteReq(msg []byte) (WriteReqMessage, error) {
	var wm WriteReqMessage

	// Opcode(4) + ID(16) + CellIndex(4) + DBLen(4) + KeyLen(4) + ValueLen(4) + SecretLen(4) = 40 bytes
	if len(msg) < 40 {
		return wm, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0
	offset += 4 // opcode

	wm.ID = make([]byte, 16)
	copy(wm.ID, msg[offset:offset+16])
	offset += 16

	wm.CellIndex = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	dbNameLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	keyLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	valueLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	secretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + keyLen + valueLen + secretLen)
	if len(msg) < offset+totalVariableLength {
		return wm, fmt.Errorf("mensaje incompleto")
	}

	wm.DBName = make([]byte, dbNameLen)
	copy(wm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	wm.Key = make([]byte, keyLen)
	copy(wm.Key, msg[offset:offset+int(keyLen)])
	offset += int(keyLen)

	wm.Value = make([]byte, valueLen)
	copy(wm.Value, msg[offset:offset+int(valueLen)])
	offset += int(valueLen)

	wm.Secret = make([]byte, secretLen)
	copy(wm.Secret, msg[offset:offset+int(secretLen)])

	return wm, nil
}

/*WRITE result
[ID: 16]
[Status: 4]*/

type WriteResult struct {
	ID     []byte
	Status int32
}

func (p *ProtocolParser) WriteResult(msg []byte) (WriteResult, error) {
	var wr WriteResult

	if len(msg) < 20 {
		return wr, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0

	wr.ID = make([]byte, 16)
	copy(wr.ID, msg[offset:offset+16])
	offset += 16

	wr.Status = int32(binary.BigEndian.Uint32(msg[offset : offset+4]))

	return wr, nil
}

func (parser *ProtocolParser) testWrite() {

	rawWriteReqMsg := []byte{
		// Opcode: 3
		0x00, 0x00, 0x00, 0x03,
		// ID: 16 bytes (16 letras 'E')
		0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45,
		0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45,
		// CellIndex: 42
		0x00, 0x00, 0x00, 0x2A,
		// DB Name Len: 5
		0x00, 0x00, 0x00, 0x05,
		// Key Len: 4
		0x00, 0x00, 0x00, 0x04,
		// Value Len: 5
		0x00, 0x00, 0x00, 0x05,
		// Secret Len: 6
		0x00, 0x00, 0x00, 0x06,
		// DB Name: "redis"
		0x72, 0x65, 0x64, 0x69, 0x73,
		// Key: "user"
		0x75, 0x73, 0x65, 0x72,
		// Value: "value"
		0x76, 0x61, 0x6C, 0x75, 0x65,
		// Secret
		0x73, 0x65, 0x63, 0x72, 0x65, 0x74,
	}

	writeReq, err := parser.WriteReq(rawWriteReqMsg)
	if err != nil {
		fmt.Printf("Error parsing write request: %v\n", err)
		return
	}
	// Verificación del resultado de escritura
	fmt.Printf("Opcode: 3 (WRITE)\n")
	fmt.Printf("Write Req ID: %s\n", string(writeReq.ID))
	fmt.Printf("CellIndex: %d\n", writeReq.CellIndex)
	fmt.Printf("DBName: %s\n", string(writeReq.DBName))
	fmt.Printf("Key: %s\n", string(writeReq.Key))
	fmt.Printf("Value: %s\n", string(writeReq.Value))
	fmt.Printf("Secret: %s\n", string(writeReq.Secret))
	fmt.Println("--------------------------------------------------")

	rawWriteResultMsg := []byte{
		// ID: 16 bytes (16 letras 'F')
		0x46, 0x46, 0x46, 0x46, 0x46, 0x46, 0x46, 0x46,
		0x46, 0x46, 0x46, 0x46, 0x46, 0x46, 0x46, 0x46,
		// Status: 1 (éxito)
		0x00, 0x00, 0x00, 0x01,
	}

	writeResult, err := parser.WriteResult(rawWriteResultMsg)
	if err != nil {
		fmt.Printf("Error parsing write result: %v\n", err)
		return
	}
	// Verificación del resultado de escritura
	fmt.Printf("Opcode: 3 (WRITE Result)\n")
	fmt.Printf("Write Result ID: %s\n", string(writeResult.ID))
	fmt.Printf("Status: %d\n", writeResult.Status)
	fmt.Println("--------------------------------------------------")
}
