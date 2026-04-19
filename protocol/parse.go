package main

import (
	"encoding/binary"
	"fmt"
)

func main() {

	parser := &ProtocolParser{}

	rawReadReqMsg := []byte{
		// Opcode: 1
		0x00, 0x00, 0x00, 0x01,
		// ID: 16 bytes (puedes usar cualquier cosa, aquí son 16 letras 'A')
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41, 0x41,
		// CellIndex: 42
		0x00, 0x00, 0x00, 0x2A,
		// DB Name Len: 5
		0x00, 0x00, 0x00, 0x05,
		// Key Len: 4
		0x00, 0x00, 0x00, 0x04,
		// Secret Len: 6
		0x00, 0x00, 0x00, 0x06,
		// DB Name: "redis"
		0x72, 0x65, 0x64, 0x69, 0x73,
		// Key: "user"
		0x75, 0x73, 0x65, 0x72,
		// Secret: "secret"
		0x73, 0x65, 0x63, 0x72, 0x65, 0x74,
	}

	result, err := parser.ReadReq(rawReadReqMsg)
	if err != nil {
		fmt.Printf("Error parsing message: %v\n", err)
		return
	}

	// Verificación
	fmt.Printf("ID: %s\n", string(result.ID))
	fmt.Printf("CellIndex: %d\n", result.CellIndex)
	fmt.Printf("DBName: %s\n", string(result.DBName))
	fmt.Printf("Key: %s\n", string(result.Key))
	fmt.Printf("Secret: %s\n", string(result.Secret))

	rawReadResultMsg := []byte{
		// ID: 16 bytes (16 letras 'B')
		0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42,
		0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42,
		// Status: 1 (éxito)
		0x00, 0x00, 0x00, 0x01,
		// CellIndex: 42
		0x00, 0x00, 0x00, 0x2A,
		// Value: "value"
		0x76, 0x61, 0x6C, 0x75, 0x65,
	}

	readResult, err := parser.ReadResult(rawReadResultMsg)
	if err != nil {
		fmt.Printf("Error parsing read result: %v\n", err)
		return
	}

	// Verificación del resultado de lectura
	fmt.Printf("Read Result ID: %s\n", string(readResult.ID))
	fmt.Printf("Status: %d\n", readResult.Status)
	fmt.Printf("CellIndex: %d\n", readResult.CellIndex)
	fmt.Printf("Value: %s\n", string(readResult.Value))

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
	fmt.Printf("Read Free Req ID: %s\n", string(readFreeReq.ID))
	fmt.Printf("DBName: %s\n", string(readFreeReq.DBName))
	fmt.Printf("Key: %s\n", string(readFreeReq.Key))

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
	fmt.Printf("Read Free Result ID: %s\n", string(readFreeResult.ID))
	fmt.Printf("Status: %d\n", readFreeResult.Status)
	fmt.Printf("Value: %s\n", string(readFreeResult.Value))

	rawWriteReqMsg := []byte{
		// Opcode: 3
		0x00, 0x00, 0x00, 0x03,
		// ID: 16 bytes (16 letras 'E')
		0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45,
		0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45,
	}

	writeReq, err := parser.WriteReq(rawWriteReqMsg)
	if err != nil {
		fmt.Printf("Error parsing write request: %v\n", err)
		return
	}
	// Verificación del resultado de escritura
	fmt.Printf("Write Req ID: %s\n", string(writeReq.ID))

	rawWriteResultMsg := []byte{
		// ID: 16 bytes (16 letras 'F')
		0x46, 0x46, 0x46, 0x46, 0x46, 0x46, 0x46, 0x46,
		0x46, 0x46, 0x46, 0x46, 0x46, 0x46, 0x46, 0x46,
	}

	writeResult, err := parser.WriteResult(rawWriteResultMsg)
	if err != nil {
		fmt.Printf("Error parsing write result: %v\n", err)
		return
	}
	// Verificación del resultado de escritura
	fmt.Printf("Write Result ID: %s\n", string(writeResult.ID))

}

type ProtocolParser struct{}

/*
READ (0x01)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Key Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Key: M]
[Secret: P]
*/
type ReadReqMessage struct {
	ID        []byte
	CellIndex uint32
	DBName    []byte
	Key       []byte
	Secret    []byte
}

func (p *ProtocolParser) ReadReq(msg []byte) (ReadReqMessage, error) {
	var rm ReadReqMessage

	// 1. Validar el tamaño mínimo de la cabecera fija
	// Opcode(4) + ID(16) + CellIndex(4) + DBLen(4) + KeyLen(4) + SecretLen(4) = 36 bytes
	if len(msg) < 36 {
		// Retornamos el struct vacío si el mensaje es demasiado corto
		// En un entorno de producción, sería mejor cambiar la firma de la
		// función para retornar (ReadMessage, error)
		return rm, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0

	// Saltamos el Opcode (4 bytes) ya que no está en el struct
	offset += 4

	// Extraemos el ID (16 bytes)
	// Usamos copy para no mantener la referencia al array original (evita fugas de memoria)
	rm.ID = make([]byte, 16)
	copy(rm.ID, msg[offset:offset+16])
	offset += 16

	// Extraemos el CellIndex (4 bytes)
	// Nota: Usamos BigEndian porque es el estándar para protocolos de red.
	// Si tu sistema origen usa LittleEndian, cámbialo a binary.LittleEndian.
	rm.CellIndex = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	// Extraemos las longitudes (4 bytes cada una)
	dbNameLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	keyLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	secretLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	// 2. Validar que el mensaje contiene todos los bytes variables (payload)
	totalVariableLength := int(dbNameLen + keyLen + secretLen)
	if len(msg) < offset+totalVariableLength {
		// Mensaje incompleto (los datos reales no coinciden con las longitudes dadas)
		return rm, fmt.Errorf("mensaje incompleto")
	}

	// Extraemos DB Name (N bytes)
	rm.DBName = make([]byte, dbNameLen)
	copy(rm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	// Extraemos Key (M bytes)
	rm.Key = make([]byte, keyLen)
	copy(rm.Key, msg[offset:offset+int(keyLen)])
	offset += int(keyLen)

	// Extraemos Secret (P bytes)
	rm.Secret = make([]byte, secretLen)
	copy(rm.Secret, msg[offset:offset+int(secretLen)])

	return rm, nil
}

/*
READ Result
[ID: 16]
[Status: 4]
[CellIndex: 4]
[Value: 4]
*/
type ReadResult struct {
	ID        []byte
	Status    int32
	CellIndex uint32
	Value     []byte
}

func (p *ProtocolParser) ReadResult(msg []byte) (ReadResult, error) {
	var rr ReadResult

	// Validar tamaño mínimo (ID(16) + Status(4) + CellIndex(4) + ValueLen(4)) = 28 bytes
	if len(msg) < 28 {
		return rr, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0

	// Extraemos el ID (16 bytes)
	rr.ID = make([]byte, 16)
	copy(rr.ID, msg[offset:offset+16])
	offset += 16

	// Extraemos el Status (4 bytes)
	rr.Status = int32(binary.BigEndian.Uint32(msg[offset : offset+4]))
	offset += 4

	// Extraemos el CellIndex (4 bytes)
	rr.CellIndex = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	// Extraemos el Value (resto del mensaje)
	valueLen := len(msg) - offset
	if valueLen < 0 {
		return rr, fmt.Errorf("mensaje incompleto para el valor")
	}
	rr.Value = make([]byte, valueLen)
	copy(rr.Value, msg[offset:])

	return rr, nil
}

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

/*DELETE (0x04)
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

/*CREATE_DB (0x08)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[Secret Len: 4]
[Genesis DB len: 4]
[Genesis index: 4] |
[DB Name: N]
[Secret: M]
[Genesis DB: P]*/

type CreateDBReqMessage struct {
	ID           []byte
	DBName       []byte
	Secret       []byte
	GenesisDB    []byte
	GenesisIndex uint32
}

func (p *ProtocolParser) CreateDBReq(msg []byte) (CreateDBReqMessage, error) {
	var cm CreateDBReqMessage

	// Opcode(4) + ID(16) + DBLen(4) + SecretLen(4) + GenesisDBLen(4) + GenesisIndex(4) = 36 bytes
	if len(msg) < 36 {
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

	genesisDBLen := binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	cm.GenesisIndex = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	totalVariableLength := int(dbNameLen + secretLen + genesisDBLen)
	if len(msg) < offset+totalVariableLength {
		return cm, fmt.Errorf("mensaje incompleto")
	}

	cm.DBName = make([]byte, dbNameLen)
	copy(cm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	cm.Secret = make([]byte, secretLen)
	copy(cm.Secret, msg[offset:offset+int(secretLen)])
	offset += int(secretLen)

	cm.GenesisDB = make([]byte, genesisDBLen)
	copy(cm.GenesisDB, msg[offset:offset+int(genesisDBLen)])

	return cm, nil
}

/*CREATE_DB result
[ID: 16]
[Status: 4]*/

type CreateDBResult struct {
	ID     []byte
	Status int32
}

func (p *ProtocolParser) CreateDBResult(msg []byte) (CreateDBResult, error) {
	var cr CreateDBResult

	if len(msg) < 20 {
		return cr, fmt.Errorf("mensaje demasiado corto")
	}

	offset := 0

	cr.ID = make([]byte, 16)
	copy(cr.ID, msg[offset:offset+16])
	offset += 16

	cr.Status = int32(binary.BigEndian.Uint32(msg[offset : offset+4]))

	return cr, nil
}

/*DELETE_DB (0x09)
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

/*READ_CELL (0x05)
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

type ReadCellResult struct {
	ID     []byte
	Status int32
	Value  []byte
}

func (p *ProtocolParser) ReadCellResult(msg []byte) (ReadCellResult, error) {
	var rr ReadCellResult

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
