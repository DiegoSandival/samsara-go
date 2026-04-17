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

	readReq := handler.MarshalMessage(
		handler.OpcodeRead,
		requestID("request-00000003"),
		handler.BuildReadPayload("manual", "greeting", 0, []byte("manual-secret")),
	)

	readResp := h.HandleRaw(readReq)
	status, payload, err := handler.UnmarshalEnvelope(readResp)
	if err != nil {
		log.Fatalf("READ: invalid envelope: %v", err)
	}

	if status != handler.StatusCodeOK {
		fmt.Printf("READ status=%d payload=%q\n", status, decodeError(payload))
		return
	}

	value, err := decodeReadValue(payload)
	if err != nil {
		log.Fatalf("READ: invalid payload: %v", err)
	}

	fmt.Printf("READ status=%d value=%q\n", status, string(value))
}

func requestID(s string) handler.RequestID {
	var id handler.RequestID
	copy(id[:], []byte(s))
	return id
}

func decodeReadValue(payload []byte) ([]byte, error) {
	if len(payload) < 13 {
		return nil, fmt.Errorf("payload too short: %d", len(payload))
	}

	valueLen := binary.LittleEndian.Uint32(payload[:4])
	if int(4+valueLen) > len(payload) {
		return nil, fmt.Errorf("invalid value length: %d", valueLen)
	}

	value := make([]byte, valueLen)
	copy(value, payload[4:4+valueLen])
	return value, nil
}

func decodeError(payload []byte) string {
	if len(payload) < 4 {
		return ""
	}
	ln := binary.LittleEndian.Uint32(payload[:4])
	if len(payload) != int(4+ln) {
		return ""
	}
	return string(payload[4:])
}
