package main

import (
	"fmt"

	protocol "github.com/DiegoSandival/samsara-go/protocol"

	//ouroboros "github.com/DiegoSandival/ouroboros-go"
	samsara "github.com/DiegoSandival/samsara-go/handler"
)

var Opcode byte

func main() {

	parser := &protocol.ProtocolParser{}

	centralHandler := samsara.NewCentralHandler()

	//msg, _ := parser.CreateDBReqBytes("PRIMODB", "secretaria", 10)
	//msg := parser.DeleteDBReqBytes("aDB", "secreto", 0)
	//msg := parser.WriteReqBytes(0, []byte("private2DB"), []byte("clave"), []byte("valorando"), []byte("secreto"))
	//msg := parser.ReadReqBytes(0, []byte("private2DB"), []byte("clave"), []byte("secreto"))
	//msg := parser.ReadFreeReqBytes([]byte("private2DB"), []byte("clave"))
	//msg := parser.DeleteReqBytes([]byte("private2DB"), []byte("clave"), []byte("secreto"), 1)
	//msg := parser.ReadCellReqBytes([]byte("PRIMODB"), []byte("secretaria"), 0)

	/*
		GenomaDetailFromBytes := protocol.GenomaDetail{
			ReadSelf:        true,
			ReadNeighbors:   false,
			WriteSelf:       true,
			WriteNeighbors:  false,
			DeleteSelf:      true,
			DeleteNeighbors: false,
			DiferirSelf:     true,
			CruzarSelf:      false,
			CloneSelf:       false,
			DominanceSelf:   false,
			FreeRead:        false,
			Migradable:      false,
		}

			GenomaDetailFromBytesB := protocol.GenomaDetail{
				ReadSelf:        true,
				ReadNeighbors:   false,
				WriteSelf:       true,
				WriteNeighbors:  false,
				DeleteSelf:      true,
				DeleteNeighbors: false,
				DiferirSelf:     true,
				CruzarSelf:      false,
				CloneSelf:       false,
				DominanceSelf:   false,
				FreeRead:        false,
				Migradable:      false,
			}*/

	//msg := parser.DiferirReqBytes([]byte("PRIMODB"), []byte("secretaria"), []byte("celular"), 0, parser.GenomaDetailFromBytes(GenomaDetailFromBytes), 0, 0, 0)
	//msg := parser.DiferirReqBytes([]byte("PRIMODB"), []byte("secretaria"), []byte("celular"), 0, parser.GenomaDetailFromBytes(GenomaDetailFromBytes), 0, 0, 0)

	msg := parser.CruzarReqBytes(1, 3, 0, 0, 0, []byte("PRIMODB"), []byte("secretaria"), []byte("celular"), []byte("secretocito"))
	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)

	r, _ := parser.CruzarResult(resp)
	//r, _ := parser.DeleteDBResult(resp)

	// Verificación del resultado de creación de base de datos
	fmt.Printf(" Result ID: %s\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Printf("CellIndex: %d\n", r.CellIndex)
	//fmt.Printf("Value: %s\n", string(r.Value))
	fmt.Println("--------------------------------------------------")

}
