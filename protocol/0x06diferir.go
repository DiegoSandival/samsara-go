package main

import (
	"encoding/binary"
	"fmt"
)

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

func (parser *ProtocolParser) testDiferir() {
	// Aquí puedes agregar un mensaje de prueba para DIFERIR y verificar su análisis
}
