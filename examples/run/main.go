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

	//msg, _ := parser.CreateDBReqBytes("real", "secretaria", 10)
	//msg := parser.DeleteDBReqBytes("PRIMODB", "secretaria", 0)
	//msg := parser.WriteReqBytes(0, []byte("real"), []byte("persona1"), []byte("esto es el valor de la persona"), []byte("secretaria"))
	//msg := parser.ReadReqBytes(0, []byte("real"), []byte("persona1"), []byte("secretaria"))
	msg := parser.ReadFreeReqBytes([]byte("real"), []byte("persona1"))
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
			CruzarSelf:      true,
			CloneSelf:       false,
			DominanceSelf:   false,
			FreeRead:        false,
			Migradable:      false,
		}

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

	//msg := parser.DiferirReqBytes([]byte("real"), []byte("secretaria"), []byte("celularas"), 0, parser.GenomaDetailFromBytes(GenomaDetailFromBytes), 0, 0, 0)
	//msg := parser.DiferirReqBytes([]byte("PRIMODB"), []byte("secretaria"), []byte("celular"), 0, parser.GenomaDetailFromBytes(GenomaDetailFromBytes), 0, 0, 0)

	//msg := parser.CruzarReqBytes([]byte("real"), 7, 8, []byte("celulares"), []byte("celularas"), 0, 0, 0, []byte("cruzado"))

	resp := samsara.ProcessRequest(
		msg,
		parser,
		centralHandler,
	)

	//r, _ := parser.CreateDBResult(resp)
	//r, _ := parser.CruzarResult(resp)
	//r, _ := parser.DeleteDBResult(resp)

	//r, _ := parser.WriteResult(resp)
	//r, _ := parser.ReadResult(resp)
	//r, _ := parser.ReadFreeResult(resp)
	//r, _ := parser.DeleteResult(resp)
	//r, _ := parser.ReadCellResult(resp)
	r, _ := parser.DiferirResult(resp)

	// Verificación del resultado de creación de base de datos
	fmt.Printf("Result ID: %x\n", string(r.ID))
	fmt.Printf("Status: %d\n", r.Status)
	fmt.Printf("CellIndex: %d\n", r.CellIndex)
	//fmt.Printf("Value: %s\n", string(r.Value))
	fmt.Println("--------------------------------------------------")

}
