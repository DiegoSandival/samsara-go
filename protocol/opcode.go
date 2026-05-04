package protocol

const (
	OpcodeCreateDB byte = 0x20
	OpcodeDeleteDB byte = 0x21
	OpcodeWrite    byte = 0x22
	OpcodeRead     byte = 0x23
	OpcodeReadFree byte = 0x24
	OpcodeDelete   byte = 0x25
	OpcodeReadCell byte = 0x26
	OpcodeDiferir  byte = 0x27
	OpcodeCruzar   byte = 0x28
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
type OpcodeReqMessage struct {
	Opcode []byte
}

func (p *ProtocolParser) Opcode(msg []byte) byte {
	if len(msg) < 4 {
		return 0
	}

	//extrael el ultimo byte del opcode
	//0x00, 0x00, 0x00, 0x23,
	//a   , b   , c   , d   ,
	//opcode = d

	return msg[3]
}
