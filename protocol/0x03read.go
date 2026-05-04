package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

/*
READ (0x23)
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
		return rm, ErrMessageTooShort
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
		return rm, ErrMessageIncomplete
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

func (p *ProtocolParser) ReadResultBytes(id []byte, status int32, cellIndex uint32, value []byte) []byte {
	result := make([]byte, 16+4+4+len(value))
	offset := 0
	copy(result[offset:offset+16], id)
	offset += 16
	binary.BigEndian.PutUint32(result[offset:offset+4], uint32(status))
	offset += 4
	binary.BigEndian.PutUint32(result[offset:offset+4], cellIndex)
	offset += 4
	copy(result[offset:], value)
	return result
}

func (p *ProtocolParser) ReadReqBytes(cellIndex uint32, dbName, key, secret []byte) []byte {

	// Generar ID aleatorio de 16 bytes
	ID := make([]byte, 16)
	rand.Read(ID)
	msg := make([]byte, 4+16+4+4+4+4+len(dbName)+len(key)+len(secret))
	binary.BigEndian.PutUint32(msg[0:4], 0x23)
	copy(msg[4:20], ID)
	binary.BigEndian.PutUint32(msg[20:24], cellIndex)
	binary.BigEndian.PutUint32(msg[24:28], uint32(len(dbName)))
	binary.BigEndian.PutUint32(msg[28:32], uint32(len(key)))
	binary.BigEndian.PutUint32(msg[32:36], uint32(len(secret)))
	copy(msg[36:36+len(dbName)], dbName)
	copy(msg[36+len(dbName):36+len(dbName)+len(key)], key)
	copy(msg[36+len(dbName)+len(key):36+len(dbName)+len(key)+len(secret)], secret)
	return msg
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

func (parser *ProtocolParser) testRead() {

	rawReadReqMsg := []byte{
		// Opcode: 0x23
		0x00, 0x00, 0x00, 0x23,
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
	fmt.Printf("Opcode: 0x23 (READ)\n")
	fmt.Printf("ID: %s\n", string(result.ID))
	fmt.Printf("CellIndex: %d\n", result.CellIndex)
	fmt.Printf("DBName: %s\n", string(result.DBName))
	fmt.Printf("Key: %s\n", string(result.Key))
	fmt.Printf("Secret: %s\n", string(result.Secret))
	fmt.Println("--------------------------------------------------")

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
	fmt.Printf("Opcode: 0x23 (READ Result)\n")
	fmt.Printf("Read Result ID: %s\n", string(readResult.ID))
	fmt.Printf("Status: %d\n", readResult.Status)
	fmt.Printf("CellIndex: %d\n", readResult.CellIndex)
	fmt.Printf("Value: %s\n", string(readResult.Value))
	fmt.Println("--------------------------------------------------")
}
