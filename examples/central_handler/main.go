package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"

	handler "github.com/DiegoSandival/samsara-go/handler"
)

func main() {
	baseDir := filepath.Join(".", "central-handler-data")
	_ = os.RemoveAll(baseDir)
	defer os.RemoveAll(baseDir)

	h := handler.NewCentralHandler(baseDir)
	defer h.Close()

	var reqID handler.RequestID
	copy(reqID[:], []byte("request-00000001"))

	fmt.Println("1) READ antes de CREATE_DB (debe fallar)")
	readBefore := send(h, reqID, handler.OpcodeReadFree, handler.BuildReadFreePayload("manual", "alpha"))
	printResponse(readBefore)

	fmt.Println("2) CREATE_DB manual")
	created := send(h, reqID, handler.OpcodeCreateDB, handler.BuildCreateDBPayload(60, "manual", []byte("root-secret")))
	printResponse(created)

	fmt.Println("3) READ_FREE en DB vacia (debe ser undefined)")
	readEmpty := send(h, reqID, handler.OpcodeReadFree, handler.BuildReadFreePayload("manual", "alpha"))
	printResponse(readEmpty)

	fmt.Println("4) DELETE_DB manual")
	deleted := send(h, reqID, handler.OpcodeDeleteDB, handler.BuildManageDBPayload("manual", 0, nil))
	printResponse(deleted)

	fmt.Println("central handler raw-bytes demo completed")
}

func send(h *handler.CentralHandler, id handler.RequestID, opcode byte, payload []byte) []byte {
	rawReq := handler.MarshalMessage(opcode, id, payload)
	return h.HandleRaw(rawReq)
}

func printResponse(raw []byte) {
	status, payload, err := handler.UnmarshalEnvelope(raw)
	if err != nil {
		log.Fatalf("invalid response envelope: %v", err)
	}
	fmt.Printf("status=%d payloadLen=%d\n", status, len(payload))
	if len(payload) >= 4 {
		ln := binary.LittleEndian.Uint32(payload[:4])
		if len(payload) == int(4+ln) {
			fmt.Printf("payload(text)=%q\n", string(payload[4:]))
		}
	}
}
