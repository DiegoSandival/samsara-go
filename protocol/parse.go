package main

// opcodes:
// 0x01: READ
// 0x02: READ_FREE
// 0x03: WRITE
// 0x04: DELETE
// 0x05: READ_CELL
// 0x06: DIFERIR
// 0x07: CRUZAR
// 0x08: CREATE_DB
// 0x09: DELETE_DB
// 0x010: CREATE_GENESIS

func main() {

	parser := &ProtocolParser{}

	parser.testRead()
	parser.testReadFree()
}

type ProtocolParser struct{}
