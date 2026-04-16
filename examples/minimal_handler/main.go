package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"sync"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
	samsara "github.com/DiegoSandival/samsara-go"
)

const (
	OpcodeRead     byte = 0x01
	OpcodeReadFree byte = 0x02
	OpcodeWrite    byte = 0x03
	OpcodeDelete   byte = 0x04
	OpcodeReadCell byte = 0x05
	OpcodeDiferir  byte = 0x06
	OpcodeCruzar   byte = 0x07
	OpcodeCreateDB byte = 0x08
	OpcodeDeleteDB byte = 0x09
)

type RequestID [16]byte

type Message struct {
	Opcode  byte
	ID      RequestID
	Payload []byte
}

const (
	requestHeaderSize = 17
)

type CentralHandler struct {
	baseDir string

	mu     sync.Mutex
	stores map[string]*samsara.Store
}

func Unmarshal(data []byte) (Message, error) {
	if len(data) < requestHeaderSize {
		return Message{}, fmt.Errorf("message too short: got %d, want at least %d", len(data), requestHeaderSize)
	}

	var msg Message
	msg.Opcode = data[0]
	copy(msg.ID[:], data[1:17])
	msg.Payload = append([]byte(nil), data[17:]...)

	return msg, nil
}

func main() {
	path := "./demo.db"
	defer os.Remove(path)
	defer os.Remove(path + ".bolt")

	store, err := samsara.New(path, 64)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	var salt [16]byte
	copy(salt[:], []byte("demo-salt-123456"))

	secret := []byte("demo-secret")
	cell := samsara.NewCellWithSecret(
		salt,
		secret,
		ouroboros.LeerSelf|ouroboros.EscribirSelf,
		10,
		20,
		30,
	)

	index, err := store.DB().Append(cell)
	if err != nil {
		log.Fatal(err)
	}

	write := store.Write("saludo", []byte("hola samsara"), index, secret)
	if write.Status != samsara.StatusOK {
		log.Fatalf("write failed: %+v", write)
	}

	read := store.Read("saludo", write.NewCellIndex, secret)
	if read.Status != samsara.StatusOK {
		log.Fatalf("read failed: %+v", read)
	}

	fmt.Printf("valor=%s\n", string(read.Value))
	fmt.Printf("nuevo_indice=%d\n", read.NewCellIndex)
	fmt.Println("example completed")
}

func NewCentralHandler(baseDir string) *CentralHandler {

	return &CentralHandler{
		stores: make(map[string]*samsara.Store),
	}
}

func (h *CentralHandler) Handle(msg Message) []byte {

	switch msg.Opcode {
	case OpcodeCreateDB:
		// Handle create DB opcode

		//payload [size (int32), name len (int32), DB name (string), secret (variable)]

		//exatract DB name and size from payload
		//extract size from payload int32
		var size uint32
		if len(msg.Payload) >= 4 {
			size = uint32(msg.Payload[0]) | uint32(msg.Payload[1])<<8 | uint32(msg.Payload[2])<<16 | uint32(msg.Payload[3])<<24
		}

		//extract DB name from payload
		var nameLen uint32
		if len(msg.Payload) >= 8 {
			nameLen = uint32(msg.Payload[4]) | uint32(msg.Payload[5])<<8 | uint32(msg.Payload[6])<<16 | uint32(msg.Payload[7])<<24
		}
		var name string
		if len(msg.Payload) > 8 {
			name = string(msg.Payload[8 : 8+nameLen])
		} else {
			return []byte{0x01} // Error status: missing DB name
		}

		store, err := samsara.New(name, size)
		if err != nil {
			log.Fatal(err)
		}
		defer store.Close()

		h.stores[name] = store

		//extract secret from payload (after size and name)
		var secret []byte
		if len(msg.Payload) > int(8+nameLen) {
			secret = msg.Payload[8+nameLen:]
		} else {
			return []byte{0x01} // Error status: missing secret
		}

		var salt [16]byte
		// Read llena el slice con bytes aleatorios seguros
		_, err = rand.Read(salt[:])
		if err != nil {
			// Este error es extremadamente raro, pero debe manejarse
			log.Fatal(err)
		}

		cell := samsara.NewCellWithSecret(
			salt,
			secret,
			ouroboros.LeerSelf|ouroboros.EscribirSelf,
			0,
			0,
			0,
		)

		_, _ = store.DB().Append(cell)

		return []byte{0x00} // Success status

	case OpcodeDeleteDB:

		//payload [name len (int32), DB name (string), secret (variable)]

		//extract DB name from payload
		var nameLen uint32
		if len(msg.Payload) >= 8 {
			nameLen = uint32(msg.Payload[4]) | uint32(msg.Payload[5])<<8 | uint32(msg.Payload[6])<<16 | uint32(msg.Payload[7])<<24
		}
		var name string
		if len(msg.Payload) > 8 {
			name = string(msg.Payload[8 : 8+nameLen])
		} else {
			return []byte{0x01} // Error status: missing DB name
		}

		//extract secret from payload (after size and name)
		var secret []byte
		if len(msg.Payload) > int(8+nameLen) {
			secret = msg.Payload[8+nameLen:]
		} else {
			return []byte{0x01} // Error status: missing secret
		}

		store, exists := h.stores[name]
		if !exists {
			return []byte{0x01} // Error status: DB not found
		}

		_, err := store.DB().ReadAuth(0, secret)
		if err != nil {
			return []byte{0x01} // Error status: invalid secret
		}

		delete(h.stores, name)
		return []byte{0x00} // Success status

	case OpcodeDiferir:
		// Handle append cell opcode
		// Payload: [DB name (string), salt (16 bytes), secret (variable), genome (uint32)]

	default:
		return []byte{0x01} // Unknown opcode or error status
		// 	// Handle unknown opcode
	}
	return []byte{0x01} // Unknown opcode or error status
}
