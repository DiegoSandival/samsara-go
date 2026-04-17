package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"path/filepath"

	handler "github.com/DiegoSandival/samsara-go/handler"
)

func main() {
	baseDir := filepath.Join(".", "examples_data", "dispatcher_demo")
	h := handler.NewCentralHandler(baseDir)
	defer h.Close()

	createReq := handler.MarshalMessage(
		handler.OpcodeCreateDB,
		requestID("request-00000001"),
		handler.BuildCreateDBPayload(64, "manual", []byte("manual-secret")),
	)
	createResp := h.HandleRaw(createReq)
	printEnvelope("CREATE_DB", createResp)

	writeReq := handler.MarshalMessage(
		handler.OpcodeWrite,
		requestID("request-00000002"),
		handler.BuildWritePayload("manual", "greeting", []byte("hello world"), 0, []byte("manual-secret")),
	)
	writeResp := h.HandleRaw(writeReq)
	printEnvelope("WRITE", writeResp)
}

func requestID(s string) handler.RequestID {
	var id handler.RequestID
	copy(id[:], []byte(s))
	return id
}

func printEnvelope(label string, raw []byte) {
	status, payload, err := handler.UnmarshalEnvelope(raw)
	if err != nil {
		log.Fatalf("%s: invalid envelope: %v", label, err)
	}

	fmt.Printf("%s status=%d payloadLen=%d\n", label, status, len(payload))
	if len(payload) >= 4 {
		ln := binary.LittleEndian.Uint32(payload[:4])
		if len(payload) == int(4+ln) {
			fmt.Printf("%s payload(text)=%q\n", label, string(payload[4:]))
		}
	}
}
