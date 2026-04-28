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

	//createDBTest(parser, centralHandler)
	//deleteDBTest(parser, centralHandler)
	//writeTest(parser, centralHandler)
	//readTest(parser, centralHandler)
	//readFreeTest(parser, centralHandler)
	//deleteTest(parser, centralHandler)
	//readCellTest(parser, centralHandler)
	//diferirTest(parser, centralHandler)
	cruzarTest(parser, centralHandler)

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
		DominanceSelf:   true,
		FreeRead:        true,
	}

	msg, _ := parser.CreateDBReqBytes("pruebas", "comotellamas", 10, parser.GenomaDetailFromBytes(GenomaGenesis))

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

	msg := parser.WriteReqBytes(0, []byte("pruebas"), []byte("papeleria"), []byte("prefuero tu belleza y juventud"), []byte("comotellamas"))

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

	msg := parser.ReadReqBytes(0, []byte("pruebas"), []byte("papeleria"), []byte("comotellamas"))
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

	msg := parser.ReadFreeReqBytes([]byte("pruebas"), []byte("papeleria"))
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

func deleteTest(parser *protocol.ProtocolParser, centralHandler *samsara.CentralHandler) {

	msg := parser.DeleteReqBytes([]byte("pruebas"), []byte("papeleria"), []byte("comotellamas"), 0)
	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)
	r, _ := parser.DeleteResult(resp)

	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Println("--------------------------------------------------")
}

func readCellTest(parser *protocol.ProtocolParser, centralHandler *samsara.CentralHandler) {

	msg := parser.ReadCellReqBytes([]byte("pruebas"), []byte("comotellamas"), 0)

	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)

	r, _ := parser.ReadCellResult(resp)

	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Printf("Cell Genoma: %d\n", r.Value)

	fmt.Println("--------------------------------------------------")
}

func diferirTest(parser *protocol.ProtocolParser, centralHandler *samsara.CentralHandler) {

	msg := parser.DiferirReqBytes([]byte("pruebas"), []byte("comotellamas"), []byte("celular"), 0, parser.GenomaDetailFromBytes(protocol.GenomaDetail{
		ReadSelf:        true,
		ReadNeighbors:   false,
		WriteSelf:       true,
		WriteNeighbors:  false,
		DeleteSelf:      true,
		DeleteNeighbors: false,
		DiferirSelf:     true,
		CruzarSelf:      true,
		DominanceSelf:   false,
		FreeRead:        false,
	}), 0, 0, 0)

	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)
	r, _ := parser.DiferirResult(resp)

	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Printf("New Cell Index: %d\n", r.CellIndex)
	fmt.Println("--------------------------------------------------")
}

func cruzarTest(parser *protocol.ProtocolParser, centralHandler *samsara.CentralHandler) {
	msg := parser.CruzarReqBytes([]byte("pruebas"), 0, 4, []byte("comotellamas"), []byte("celular"), 0, 0, 0, []byte("cruzado"))

	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)
	r, _ := parser.CruzarResult(resp)

	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Printf("New Cell Index: %d\n", r.CellIndex)
	fmt.Println("--------------------------------------------------")
}
