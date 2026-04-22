package main

import (
	"log"

	protocol "github.com/DiegoSandival/samsara-go/protocol"

	//ouroboros "github.com/DiegoSandival/ouroboros-go"
	samsara "github.com/DiegoSandival/samsara-go"
)

var Opcode byte

func main() {

	parser := &protocol.ProtocolParser{}

	centralHandler := samsara.NewCentralHandler("data")

	resp := samsara.ProcessRequest(
		[]byte{0x00 /* mensaje real del protocolo */},
		parser,
		centralHandler,
	)

	log.Printf("Respuesta: %s", string(resp))

}
