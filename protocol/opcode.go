package protocol

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
	//extrael el ultimo byte del opcode
	//0x00, 0x00, 0x00, 0x23,
	//a   , b   , c   , d   ,
	//opcode = d

	return msg[3]
}
