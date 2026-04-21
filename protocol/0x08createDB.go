package protocol

import (
	"encoding/binary"
	"fmt"
)

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

func (parser *ProtocolParser) CreateDBResultByte(Id []byte, Status int32) []byte {

	//convert status to bytes
	statusBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(statusBytes, uint32(Status))

	//concatenate ID and status
	result := append(Id, statusBytes...)

	return result
}

func (parser *ProtocolParser) testCreateDB() {
	rawCreateDBReqMsg := []byte{
		0x00, 0x00, 0x00, 0x08, // Opcode
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // ID
		0x00, 0x00, 0x00, 0x04, // DB Name Len
		0x00, 0x00, 0x00, 0x04, // Secret Len
		0x00, 0x00, 0x00, 0x04, // Genesis DB Len
	}

	readCreateDBReq, err := parser.CreateDBReq(rawCreateDBReqMsg)
	if err != nil {
		fmt.Printf("Error al parsear CreateDBReq: %v\n", err)
	} else {
		fmt.Println("CreateDBReq parseado correctamente")
	}

	fmt.Printf("Opcode: 0x08 (CreateDB)\n")
	fmt.Printf("CreateDB Req ID: %s\n", string(readCreateDBReq.ID))
	fmt.Printf("DBName: %s\n", string(readCreateDBReq.DBName))
	fmt.Printf("Secret: %s\n", string(readCreateDBReq.Secret))
	fmt.Printf("GenesisDB: %s\n", string(readCreateDBReq.GenesisDB))
	fmt.Printf("GenesisIndex: %d\n", readCreateDBReq.GenesisIndex)
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

	fmt.Printf("Opcode: 0x08 (CreateDB Result)\n")
	fmt.Printf("CreateDB Result ID: %s\n", string(readCreateDBResult.ID))
	fmt.Printf("Status: %d\n", readCreateDBResult.Status)
	fmt.Println("--------------------------------------------------")
}
