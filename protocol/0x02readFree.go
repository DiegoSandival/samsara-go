package main

import (
	"encoding/binary"
	"fmt"
)

/*READ_FREE (0x02)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[Key Len: 4] |
[DB Name: N]
[Key: M]*/

type ReadFreeReqMessage struct {
	ID     []byte
	DBName []byte
	Key    []byte
}

func (p *ProtocolParser) ReadFreeReq(msg []byte) (ReadFreeReqMessage, error) {
	var rm ReadFreeReqMessage

	// Opcode(4) + ID(16) + DBLen(4) + KeyLen(4) = 28 bytes
	if len(msg) < 28 {
		return rm, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0
	offset += 4 // opcode

	rm.ID = make([]byte, 16)
	copy(rm.ID, msg[offset:offset+16])
	offset += 16

	dbNameLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	keyLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + keyLen)
	if len(msg) < offset+totalVariableLength {
		return rm, fmt.Errorf("mensaje incompleto")
	}

	rm.DBName = make([]byte, dbNameLen)
	copy(rm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	rm.Key = make([]byte, keyLen)
	copy(rm.Key, msg[offset:offset+int(keyLen)])

	return rm, nil
}

/*READ_FREE result
[ID: 16]
[Status: 4]
[Value: 4]*/

type ReadFreeResult struct {
	ID     []byte
	Status int32
	Value  []byte
}

func (p *ProtocolParser) ReadFreeResult(msg []byte) (ReadFreeResult, error) {
	var rr ReadFreeResult

	// ID(16) + Status(4) = 20 bytes
	if len(msg) < 20 {
		return rr, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0

	rr.ID = make([]byte, 16)
	copy(rr.ID, msg[offset:offset+16])
	offset += 16

	rr.Status = int32(binary.BigEndian.Uint32(msg[offset : offset+4]))
	offset += 4

	valueLen := len(msg) - offset
	if valueLen < 0 {
		return rr, fmt.Errorf("mensaje incompleto para el valor")
	}

	rr.Value = make([]byte, valueLen)
	copy(rr.Value, msg[offset:])

	return rr, nil
}

func (parser *ProtocolParser) testReadFree() {

	rawReadFreeReqMsg := []byte{
		// Opcode: 2
		0x00, 0x00, 0x00, 0x02,
		// ID: 16 bytes (16 letras 'C')
		0x43, 0x43, 0x43, 0x43, 0x43, 0x43, 0x43, 0x43,
		0x43, 0x43, 0x43, 0x43, 0x43, 0x43, 0x43, 0x43,
		// DB Name Len: 5
		0x00, 0x00, 0x00, 0x05,
		// Key Len: 4
		0x00, 0x00, 0x00, 0x04,
		// DB Name: "redis"
		0x72, 0x65, 0x64, 0x69, 0x73,
		// Key: "user"
		0x75, 0x73, 0x65, 0x72,
	}

	readFreeReq, err := parser.ReadFreeReq(rawReadFreeReqMsg)
	if err != nil {
		fmt.Printf("Error parsing read free request: %v\n", err)
		return
	}

	// Verificación del resultado de lectura libre
	fmt.Printf("Opcode: 2 (READ FREE)\n")
	fmt.Printf("Read Free Req ID: %s\n", string(readFreeReq.ID))
	fmt.Printf("DBName: %s\n", string(readFreeReq.DBName))
	fmt.Printf("Key: %s\n", string(readFreeReq.Key))
	fmt.Println("--------------------------------------------------")

	rawReadFreeResultMsg := []byte{
		// ID: 16 bytes (16 letras 'D')
		0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44,
		0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44,
		// Status: 0 (error)
		0x00, 0x00, 0x00, 0x00,
		// Value: "error"
		0x65, 0x72, 0x72, 0x6F, 0x72,
	}

	readFreeResult, err := parser.ReadFreeResult(rawReadFreeResultMsg)
	if err != nil {
		fmt.Printf("Error parsing read free result: %v\n", err)
		return
	}

	// Verificación del resultado de lectura libre
	fmt.Printf("Opcode: 2 (READ FREE Result)\n")
	fmt.Printf("Read Free Result ID: %s\n", string(readFreeResult.ID))
	fmt.Printf("Status: %d\n", readFreeResult.Status)
	fmt.Printf("Value: %s\n", string(readFreeResult.Value))
	fmt.Println("--------------------------------------------------")
}
