package main

import (
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

var Opcode byte

func main() {

	parser := &protocol.ProtocolParser{}

	CentralHandler := &CentralHandler{
		baseDir: "data",
		stores:  make(map[string]*Store),
	}

	ProcessRequest([]byte{0x00 /* Aquí irían los datos del mensaje */}, parser, CentralHandler)

}
