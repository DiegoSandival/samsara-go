package main

import (
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

		//payload [size (int32), DB name (string)]

		//exatract DB name and size from payload
		//extract size from payload int32
		var size int32
		if len(msg.Payload) >= 4 {
			size = int32(msg.Payload[0]) | int32(msg.Payload[1])<<8 | int32(msg.Payload[2])<<16 | int32(msg.Payload[3])<<24
		}

		store, err := samsara.New(string(msg.Payload[4:]), uint32(size))
		if err != nil {
			log.Fatal(err)
		}
		defer store.Close()

		h.stores[string(msg.Payload[4:])] = store
		return []byte{0x00} // Success status
	case OpcodeDeleteDB:
		// Handle delete DB opcode
		delete(h.stores, string(msg.Payload[4:]))
		return []byte{0x00} // Success status
	}
	return []byte{0x01} // Unknown opcode or error status
}
