package main

import (
	"fmt"

	protocol "github.com/DiegoSandival/samsara-go/protocol"

	//ouroboros "github.com/DiegoSandival/ouroboros-go"
	samsara "github.com/DiegoSandival/samsara-go"
)

var Opcode byte

func main() {

	parser := &protocol.ProtocolParser{}

	centralHandler := samsara.NewCentralHandler()

	rawCreateDBReqMsg := []byte{
		0x00, 0x00, 0x00, 0x08, // Opcode
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, // ID
		0x00, 0x00, 0x00, 0x04, // DB Name Len
		0x00, 0x00, 0x00, 0x04, // Secret Len
		0x00, 0x00, 0x00, 0x04, // Genesis DB Len
		0x00, 0x00, 0x00, 0x01, // Genesis Index
		0x00, 0x00, 0x00, 0x64, // DB Size (100)
		0x64, 0x62, 0x31, 0x32, // DB Name: "db12"
		0x73, 0x65, 0x63, 0x72, // Secret: "secr"
		0x67, 0x65, 0x6E, 0x31, // Genesis DB: "gen1"
	}

	resp := samsara.ProcessRequest(
		rawCreateDBReqMsg,
		parser,
		centralHandler,
	)

	r, _ := parser.CreateDBResult(resp)

	// Verificación del resultado de escritura
	fmt.Printf("Opcode: 3 (WRITE Result)\n")
	fmt.Printf("Write Result ID: %s\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Println("--------------------------------------------------")

	/*
		rawWriteReqMsg := []byte{
			// Opcode: 3
			0x00, 0x00, 0x00, 0x03,
			// ID: 16 bytes (16 letras 'E')
			0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45,
			0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45, 0x45,
			// CellIndex: 0
			0x00, 0x00, 0x00, 0x00,
			// DB Name Len: 4
			0x00, 0x00, 0x00, 0x04,
			// Key Len: 4
			0x00, 0x00, 0x00, 0x04,
			// Value Len: 5
			0x00, 0x00, 0x00, 0x05,
			// Secret Len: 4
			0x00, 0x00, 0x00, 0x04,
			// DB Name: "db12"
			0x64, 0x62, 0x31, 0x32, 0x00,
			// Key: "user"
			0x75, 0x73, 0x65, 0x72,
			// Value: "value"
			0x76, 0x61, 0x6C, 0x75, 0x65,
			// Secret: secr
			0x73, 0x65, 0x63, 0x72,
		}

		resp := samsara.ProcessRequest(
			rawWriteReqMsg,
			parser,
			centralHandler,
		)

		r, _ := parser.WriteResult(resp)

		// Verificación del resultado de escritura
		fmt.Printf("Opcode: 3 (WRITE Result)\n")
		fmt.Printf("Write Result ID: %s\n", string(r.ID))
		fmt.Printf("Status: %d\n", r.Status)
		fmt.Println("--------------------------------------------------")
	*/
}
