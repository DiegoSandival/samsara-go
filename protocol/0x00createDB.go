package protocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

/*CREATE_DB (0x00)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[Secret Len: 4]
[DB Size: 4] |
[DB Name: N]
[Secret: M]*/

type CreateDBReqMessage struct {
	ID     []byte
	DBName []byte
	Secret []byte
	DBSize uint32
}

func (p *ProtocolParser) CreateDBReqBytes(DBName string, Secret string, DBSize uint32) ([]byte, error) {
	oPcode := uint32(0x00)
	// Generar ID aleatorio de 16 bytes
	ID := make([]byte, 16)
	_, err := rand.Read(ID)
	if err != nil {
		return nil, fmt.Errorf("error generando ID: %v", err)
	}

	dbNameBytes := []byte(DBName)
	secretBytes := []byte(Secret)
	dbNameLen := uint32(len(dbNameBytes))
	secretLen := uint32(len(secretBytes))
	// Crear el mensaje concatenando todos los campos
	message := make([]byte, 4+16+4+4+4+dbNameLen+secretLen)
	binary.BigEndian.PutUint32(message[0:4], oPcode) // Opcode
	copy(message[4:20], ID)
	binary.BigEndian.PutUint32(message[20:24], dbNameLen)
	binary.BigEndian.PutUint32(message[24:28], secretLen)
	binary.BigEndian.PutUint32(message[28:32], DBSize)
	copy(message[32:32+dbNameLen], dbNameBytes)
	copy(message[32+dbNameLen:32+dbNameLen+secretLen], secretBytes)

	return message, nil
}

func (p *ProtocolParser) CreateDBReq(msg []byte) (CreateDBReqMessage, error) {
	var cm CreateDBReqMessage

	// Opcode(4) + ID(16) + DBLen(4) + SecretLen(4) + DBSize(4) = 32 bytes
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

	cm.DBSize = binary.BigEndian.Uint32(msg[offset : offset+4])
	offset += 4

	cm.DBName = make([]byte, dbNameLen)
	copy(cm.DBName, msg[offset:offset+int(dbNameLen)])
	offset += int(dbNameLen)

	cm.Secret = make([]byte, secretLen)
	copy(cm.Secret, msg[offset:offset+int(secretLen)])
	offset += int(secretLen)

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

func (parser *ProtocolParser) CreateDBResultBytes(Id []byte, Status int32) []byte {

	//convert status to bytes
	statusBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(statusBytes, uint32(Status))

	//concatenate ID and status
	result := append(Id, statusBytes...)

	return result
}

func (parser *ProtocolParser) testCreateDB() {
	rawCreateDBReqMsg := []byte{
		0x00, 0x00, 0x00, 0x00, // Opcode
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // ID
		0x00, 0x00, 0x00, 0x04, // DB Name Len
		0x00, 0x00, 0x00, 0x04, // Secret Len
		0x00, 0x00, 0x00, 0x64, // DB Size (100)
		0x64, 0x62, 0x31, 0x32, // DB Name: "db12"
		0x73, 0x65, 0x63, 0x72, // Secret: "secr"
	}

	readCreateDBReq, err := parser.CreateDBReq(rawCreateDBReqMsg)
	if err != nil {
		fmt.Printf("Error al parsear CreateDBReq: %v\n", err)
	} else {
		fmt.Println("CreateDBReq parseado correctamente")
	}

	fmt.Printf("Opcode: 0x00 (CreateDB)\n")
	fmt.Printf("CreateDB Req ID: %s\n", string(readCreateDBReq.ID))
	fmt.Printf("DBName: %s\n", string(readCreateDBReq.DBName))
	fmt.Printf("Secret: %s\n", string(readCreateDBReq.Secret))
	fmt.Println("--------------------------------------------------")

	rawCreateDBResultMsg := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // ID
		0x00, 0x00, 0x00, 0x01, // Status
	}

	readCreateDBResult, err := parser.CreateDBResult(rawCreateDBResultMsg)
	if err != nil {
		fmt.Printf("Error al parsear CreateDBResult: %v\n", err)
	} else {
		fmt.Println("CreateDBResult parseado correctamente")
	}

	fmt.Printf("Opcode: 0x00 (CreateDB Result)\n")
	fmt.Printf("CreateDB Result ID: %s\n", string(readCreateDBResult.ID))
	fmt.Printf("Status: %d\n", readCreateDBResult.Status)
	fmt.Println("--------------------------------------------------")
}
