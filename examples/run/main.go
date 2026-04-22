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

	//msg, _ := parser.CreateDBReqBytes("privateDB", "secreto", 100)
	//msg := parser.WriteReqBytes(0, []byte("privateDB"), []byte("clave"), []byte("valor"), []byte("secreto"))
	msg := parser.ReadReqBytes(0, []byte("privateDB"), []byte("clave"), []byte("secreto"))
	//msg := parser.ReadFreeReqBytes([]byte("privateDB"), []byte("clave"))

	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)

	//r, _ := parser.ReadFreeResult(resp)
	r, _ := parser.ReadResult(resp)

	// Verificación del resultado de creación de base de datos
	fmt.Printf(" Result ID: %s\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Printf("Value: %s\n", string(r.Value))
	fmt.Println("--------------------------------------------------")

}
