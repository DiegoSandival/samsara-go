package main

import (
	"fmt"
	"log"

	protocol "github.com/DiegoSandival/samsara-go/protocol"

	//ouroboros "github.com/DiegoSandival/ouroboros-go"
	samsara "github.com/DiegoSandival/samsara-go/handler"
)

var Opcode byte

func main() {

	parser := &protocol.ProtocolParser{}

	centralHandler := samsara.NewCentralHandler()

	createDBTest(parser, centralHandler)
	//deleteDBTest(parser, centralHandler)
	//writeTest(parser, centralHandler)
	//readTest(parser, centralHandler)
	//readFreeTest(parser, centralHandler)

	//msg := parser.ReadFreeReqBytes([]byte("real"), []byte("persona1"))
	//msg := parser.DeleteReqBytes([]byte("private2DB"), []byte("clave"), []byte("secreto"), 1)
	//msg := parser.ReadCellReqBytes([]byte("PRIMODB"), []byte("secretaria"), 0)

	//msg := parser.DiferirReqBytes([]byte("PRIMODB"), []byte("secretaria"), []byte("celular"), 0, parser.GenomaDetailFromBytes(GenomaDetailFromBytes), 0, 0, 0)
	//msg := parser.CruzarReqBytes([]byte("real"), 7, 8, []byte("celulares"), []byte("celularas"), 0, 0, 0, []byte("cruzado"))

	//r, _ := parser.CruzarResult(resp)
	//r, _ := parser.DeleteDBResult(resp)

}

func createDBTest(parser *protocol.ProtocolParser, centralHandler *samsara.CentralHandler) {
	GenomaGenesis := protocol.GenomaDetail{
		ReadSelf:        true,
		ReadNeighbors:   true,
		WriteSelf:       true,
		WriteNeighbors:  true,
		DeleteSelf:      true,
		DeleteNeighbors: true,
		DiferirSelf:     true,
		CruzarSelf:      true,
		CloneSelf:       true,
		DominanceSelf:   true,
		FreeRead:        true,
		IsMigrated:      false,
	}

	msg, _ := parser.CreateDBReqBytes("example", "secretaria", 10, parser.GenomaDetailFromBytes(GenomaGenesis))

	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)

	r, _ := parser.CreateDBResult(resp)

	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Println("--------------------------------------------------")

}

func deleteDBTest(parser *protocol.ProtocolParser, centralHandler *samsara.CentralHandler) {

	msg := parser.DeleteDBReqBytes("example", "secretaria", 0)

	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)

	r, _ := parser.DeleteDBResult(resp)
	log.Println("d")
	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Println("--------------------------------------------------")

}

func writeTest(parser *protocol.ProtocolParser, centralHandler *samsara.CentralHandler) {

	msg := parser.WriteReqBytes(0, []byte("real"), []byte("persona1"), []byte("esto es el valor de la persona"), []byte("secretaria"))

	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)

	r, _ := parser.WriteResult(resp)

	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Println("--------------------------------------------------")
}

func readTest(parser *protocol.ProtocolParser, centralHandler *samsara.CentralHandler) {

	msg := parser.ReadReqBytes(0, []byte("real"), []byte("persona1"), []byte("secretaria"))
	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)
	r, _ := parser.ReadResult(resp)

	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Printf("CellIndex: %d\n", r.CellIndex)
	fmt.Printf("Value: %s\n", string(r.Value))
	fmt.Println("--------------------------------------------------")
}

func readFreeTest(parser *protocol.ProtocolParser, centralHandler *samsara.CentralHandler) {

	msg := parser.ReadFreeReqBytes([]byte("real"), []byte("persona1"))
	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)
	r, _ := parser.ReadFreeResult(resp)

	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Printf("Value: %s\n", string(r.Value))
	fmt.Println("--------------------------------------------------")
}
