package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

/*READ_CELL (0x06)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Secret: M]*/

type ReadCellReqMessage struct {
	ID        []byte
	CellIndex uint32
	DBName    []byte
	Secret    []byte
}

func (p *ProtocolParser) ReadCellReqBytes(dbName, secret []byte, cellIndex uint32) []byte {

	// Generar ID aleatorio de 16 bytes
	ID := make([]byte, 16)
	rand.Read(ID)

	dbNameLen := uint32(len(dbName))
	secretLen := uint32(len(secret))
	totalLen := 4 + 16 + 4 + 4 + 4 + dbNameLen + secretLen
	msg := make([]byte, totalLen)
	// Opcode
	binary.BigEndian.PutUint32(msg[0:4], 0x06)
	// ID
	copy(msg[4:20], ID)
	// CellIndex
	binary.BigEndian.PutUint32(msg[20:24], cellIndex)
	// DB Name Len
	binary.BigEndian.PutUint32(msg[24:28], dbNameLen)
	// Secret Len
	binary.BigEndian.PutUint32(msg[28:32], secretLen)
	// DB Name
	copy(msg[32:32+dbNameLen], dbName)
	// Secret
	copy(msg[32+dbNameLen:], secret)
	return msg
}

func (p *ProtocolParser) ReadCellReq(msg []byte) (ReadCellReqMessage, error) {
	var rm ReadCellReqMessage

	// Opcode(4) + ID(16) + CellIndex(4) + DBLen(4) + SecretLen(4) = 32 bytes
	if len(msg) < 32 {
		return rm, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0
	offset += 4 // opcode

	rm.ID = make([]byte, 16)
	copy(rm.ID, msg[offset:offset+16])
	offset += 16

	rm.CellIndex = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	dbNameLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	secretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + secretLen)
	if len(msg) < offset+totalVariableLength {
		return rm, fmt.Errorf("mensaje incompleto")
	}

	rm.DBName = make([]byte, dbNameLen)
	copy(rm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	rm.Secret = make([]byte, secretLen)
	copy(rm.Secret, msg[offset:offset+int(secretLen)])

	return rm, nil
}

/*READ_CELL result
[ID: 16]
[Status: 4]
[Value: 4]*/

type ReadCellResultMessage struct {
	ID     []byte
	Status int32
	Value  []byte
}

func (p *ProtocolParser) ReadCellResultBytes(id []byte, status int32, value []byte) []byte {

	valueLen := uint32(len(value))
	totalLen := 16 + 4 + 4 + valueLen
	msg := make([]byte, totalLen)
	// ID
	copy(msg[0:16], id)
	// Status
	binary.BigEndian.PutUint32(msg[16:20], uint32(status))
	// Value Len
	binary.BigEndian.PutUint32(msg[20:24], valueLen)
	// Value
	copy(msg[24:], value)
	return msg
}

func (p *ProtocolParser) ReadCellResult(msg []byte) (ReadCellResultMessage, error) {
	var rr ReadCellResultMessage

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

func (parser *ProtocolParser) testReadCell() {

	rawReadCellReqMsg := []byte{
		// Opcode: 6
		0x00, 0x00, 0x00, 0x06,
		// ID: 16 bytes (16 letras 'I')
		0x49, 0x49, 0x49, 0x49, 0x49, 0x49, 0x49, 0x49,
		0x49, 0x49, 0x49, 0x49, 0x49, 0x49, 0x49, 0x49,
		// CellIndex: 42
		0x00, 0x00, 0x00, 0x2A,
		// DB Name Len: 5
		0x00, 0x00, 0x00, 0x05,
		// Secret Len: 4
		0x00, 0x00, 0x00, 0x04,
		// DB Name: "redis"
		0x72, 0x65, 0x64, 0x69, 0x73,
		// Secret
		0x73, 0x65, 0x63, 0x72,
	}

	readCellReq, err := parser.ReadCellReq(rawReadCellReqMsg)
	if err != nil {
		fmt.Printf("Error parsing read cell request: %v\n", err)
		return
	}

	// Verificación del resultado de lectura de celda
	fmt.Printf("Opcode: 6 (READ_CELL)\n")
	fmt.Printf("Read Cell Req ID: %s\n", string(readCellReq.ID))
	fmt.Printf("CellIndex: %d\n", readCellReq.CellIndex)
	fmt.Printf("DBName: %s\n", string(readCellReq.DBName))
	fmt.Printf("Secret: %s\n", string(readCellReq.Secret))
	fmt.Println("--------------------------------------------------")

	rawReadResultMsg := []byte{
		// ID: 16 bytes (16 letras 'J')
		0x4A, 0x4A, 0x4A, 0x4A, 0x4A, 0x4A, 0x4A, 0x4A,
		0x4A, 0x4A, 0x4A, 0x4A, 0x4A, 0x4A, 0x4A, 0x4A,
		// Status: 1 (éxito)
		0x00, 0x00, 0x00, 0x01,
		// Value: "cellvalue"
		0x63, 0x65, 0x6C, 0x6C, 0x76, 0x61, 0x6C, 0x75, 0x65,
	}

	readCellResult, err := parser.ReadCellResult(rawReadResultMsg)
	if err != nil {
		fmt.Printf("Error parsing read cell result: %v\n", err)
		return
	}
	// Verificación del resultado de lectura de celda
	fmt.Printf("Opcode: 6 (READ_CELL Result)\n")
	fmt.Printf("Read Cell Result ID: %s\n", string(readCellResult.ID))
	fmt.Printf("Status: %d\n", readCellResult.Status)
	fmt.Printf("Value: %s\n", string(readCellResult.Value))
	fmt.Println("--------------------------------------------------")

}
