package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
	samsara "github.com/DiegoSandival/samsara-go"

	handler "github.com/DiegoSandival/samsara-go/handler"
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

func (h *CentralHandler) RegisterStore(name string, store *samsara.Store) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stores[name] = store
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

	baseDir := filepath.Join(".", "central-handler-data-minimal")

	h := handler.NewCentralHandler(baseDir)
	defer h.Close()

	//crate or load list of DBs from disk

	//create or load file for each DB in list

	db_names, err := os.OpenFile(baseDir+string(os.PathSeparator)+"db_names.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer db_names.Close()

	db_name_list, err := io.ReadAll(db_names)
	if err != nil {
		panic(err)
	}

	//pasamos la lista de DBs a memoria un slice de strings
	var dbs_list []string
	for _, name := range bytes.Split(db_name_list, []byte{'\n'}) {
		if len(name) > 0 {
			dbs_list = append(dbs_list, string(name))
		}
	}

	//cargamos cada DB en memoria
	for _, name := range dbs_list {
		store, err := samsara.Open(name)
		if err != nil {
			log.Fatal(err)
		}

		h.RegisterStore(name, store)
	}

	//OPCODE, ID,  payload [size (int32), name len (int32), DB name (string), secret (variable)]
	var createDBreq []byte
	createDBreq = append(createDBreq, OpcodeCreateDB)
	createDBreq = append(createDBreq, []byte("request-00000001")...)
	size := uint32(64)
	name := "manual"
	nameLen := uint32(len(name))
	createDBreq = append(createDBreq, byte(size), byte(size>>8), byte(size>>16), byte(size>>24))
	createDBreq = append(createDBreq, byte(nameLen), byte(nameLen>>8), byte(nameLen>>16), byte(nameLen>>24))
	createDBreq = append(createDBreq, []byte(name)...)
	createDBreq = append(createDBreq, []byte("manual-secret")...)

	h.HandleRaw(createDBreq)

	//OPCODE, ID, payload [name len (int32), DB name (string), secret (variable)]
	var deleteDBreq []byte
	deleteDBreq = append(deleteDBreq, OpcodeDeleteDB)
	deleteDBreq = append(deleteDBreq, []byte("request-00000004")...)
	nameLen = uint32(len(name))
	deleteDBreq = append(deleteDBreq, byte(nameLen), byte(nameLen>>8), byte(nameLen>>16), byte(nameLen>>24))
	deleteDBreq = append(deleteDBreq, []byte(name)...)
	deleteDBreq = append(deleteDBreq, []byte("manual-secret")...)

	h.HandleRaw(deleteDBreq)

	//opcode, ID, payload [cel_index, secrete_len, secrete, dbname_len, DB name, len_key, key, value]
	var writeReq []byte
	writeReq = append(writeReq, OpcodeWrite)
	writeReq = append(writeReq, []byte("request-00000002")...)
	cellIndex := uint32(0)
	secret := []byte("manual-secret")
	dbName := "manual"
	key := "greeting"
	value := []byte("hello world")
	writeReq = append(writeReq, byte(cellIndex), byte(cellIndex>>8), byte(cellIndex>>16), byte(cellIndex>>24))
	secretLen := uint32(len(secret))
	writeReq = append(writeReq, byte(secretLen), byte(secretLen>>8), byte(secretLen>>16), byte(secretLen>>24))
	writeReq = append(writeReq, secret...)
	dbNameLen := uint32(len(dbName))
	writeReq = append(writeReq, byte(dbNameLen), byte(dbNameLen>>8), byte(dbNameLen>>16), byte(dbNameLen>>24))
	writeReq = append(writeReq, []byte(dbName)...)
	keyLen := uint32(len(key))
	writeReq = append(writeReq, byte(keyLen), byte(keyLen>>8), byte(keyLen>>16), byte(keyLen>>24))
	writeReq = append(writeReq, []byte(key)...)
	writeReq = append(writeReq, value...)

	h.HandleRaw(writeReq)

	//opcode, ID, payload [cel_index, secrete_len, secrete, key]
	var readReq []byte
	readReq = append(readReq, OpcodeRead)
	readReq = append(readReq, []byte("request-00000003")...)
	cellIndex = uint32(0)
	secret = []byte("manual-secret")
	readReq = append(readReq, byte(cellIndex), byte(cellIndex>>8), byte(cellIndex>>16), byte(cellIndex>>24))
	secretLen = uint32(len(secret))
	readReq = append(readReq, byte(secretLen), byte(secretLen>>8), byte(secretLen>>16), byte(secretLen>>24))
	readReq = append(readReq, secret...)
	key = "greeting"
	keyLen = uint32(len(key))
	readReq = append(readReq, byte(keyLen), byte(keyLen>>8), byte(keyLen>>16), byte(keyLen>>24))
	readReq = append(readReq, []byte(key)...)

	h.HandleRaw(readReq)

}

func NewCentralHandler(baseDir string) *CentralHandler {

	return &CentralHandler{
		baseDir: baseDir,
		stores:  make(map[string]*samsara.Store),
	}
}

func (h *CentralHandler) HandleRaw(data []byte) []byte {
	msg, err := Unmarshal(data)
	if err != nil {
		return []byte{0x01} // Error status: invalid message format
	}
	return h.Handle(msg)
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
		var secret []byte
		if len(msg.Payload) > int(8+nameLen) {
			secret = msg.Payload[8+nameLen:]
		} else {
			return []byte{0x01} // Error status: missing secret
		}

		store, err := samsara.New(name, size)
		if err != nil {
			log.Fatal(err)
		}
		defer store.Close()

		h.stores[name] = store

		//extract secret from payload (after size and name)

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

	case OpcodeWrite:
		// Handle write opcode
		// Payload: [cel_index, secrete_len, secrete, dbname_len, DB name, len_key, key, value]

		//extrac cell index from payload
		var cellIndex uint32
		if len(msg.Payload) >= 4 {
			cellIndex = uint32(msg.Payload[0]) | uint32(msg.Payload[1])<<8 | uint32(msg.Payload[2])<<16 | uint32(msg.Payload[3])<<24
		}

		//extract secret len from payload
		var secretLen uint32
		if len(msg.Payload) >= 8 {
			secretLen = uint32(msg.Payload[4]) | uint32(msg.Payload[5])<<8 | uint32(msg.Payload[6])<<16 | uint32(msg.Payload[7])<<24
		}

		//extract secret from payload
		var secret []byte
		if len(msg.Payload) > int(8+secretLen) {
			secret = msg.Payload[8 : 8+secretLen]
		} else {
			return []byte{0x01} // Error status: missing secret
		}

		//extract DB name len from payload
		var dbNameLen uint32
		if len(msg.Payload) >= int(8+secretLen+4) {
			dbNameLen = uint32(msg.Payload[8+secretLen]) | uint32(msg.Payload[8+secretLen+1])<<8 | uint32(msg.Payload[8+secretLen+2])<<16 | uint32(msg.Payload[8+secretLen+3])<<24
		}
		//extract DB name from payload
		var dbName string
		if len(msg.Payload) > int(8+secretLen+4) {
			dbName = string(msg.Payload[8+secretLen+4 : 8+secretLen+4+dbNameLen])
		} else {
			return []byte{0x01} // Error status: missing DB name
		}

		//extract key len from payload
		var keyLen uint32
		if len(msg.Payload) >= int(8+secretLen+4+dbNameLen+4) {
			keyLen = uint32(msg.Payload[8+secretLen+4+dbNameLen]) | uint32(msg.Payload[8+secretLen+4+dbNameLen+1])<<8 | uint32(msg.Payload[8+secretLen+4+dbNameLen+2])<<16 | uint32(msg.Payload[8+secretLen+4+dbNameLen+3])<<24
		}
		//extract key from payload
		var key []byte
		if len(msg.Payload) > int(8+secretLen+4+dbNameLen+4) {
			key = msg.Payload[8+secretLen+4+dbNameLen+4 : 8+secretLen+4+dbNameLen+4+keyLen]
		} else {
			return []byte{0x01} // Error status: missing key
		}
		//extract value from payload (after key)
		var value []byte
		if len(msg.Payload) > int(8+secretLen+4+dbNameLen+4+keyLen) {
			value = msg.Payload[8+secretLen+4+dbNameLen+4+keyLen:]
		} else {
			return []byte{0x01} // Error status: missing value
		}

		h.stores[dbName].Write(string(key), value, cellIndex, secret)

		return []byte{0x00} // Success status

	case OpcodeRead:
		// Handle read opcode
		// Payload: [cel_index, secrete_len, secrete, key]

		//extrac cell index from payload
		var cellIndex uint32
		if len(msg.Payload) >= 4 {
			cellIndex = uint32(msg.Payload[0]) | uint32(msg.Payload[1])<<8 | uint32(msg.Payload[2])<<16 | uint32(msg.Payload[3])<<24
		}

		//extract secret len from payload
		var secretLen uint32
		if len(msg.Payload) >= 8 {
			secretLen = uint32(msg.Payload[4]) | uint32(msg.Payload[5])<<8 | uint32(msg.Payload[6])<<16 | uint32(msg.Payload[7])<<24
		}

		//extract secret from payload
		var secret []byte
		if len(msg.Payload) > int(8+secretLen) {
			secret = msg.Payload[8 : 8+secretLen]
		} else {
			return []byte{0x01} // Error status: missing secret
		}

		//extract key from payload
		var key []byte
		if len(msg.Payload) > int(8+secretLen) {
			key = msg.Payload[8+secretLen:]
		} else {
			return []byte{0x01} // Error status: missing key
		}

		result := h.stores["manual"].Read(string(key), cellIndex, secret)

		return encodeReadResult(result) // Success: return the value

	default:
		return []byte{0x01} // Unknown opcode or error status
		// 	// Handle unknown opcode
	}
}

func encodeReadResult(r samsara.ReadResult) []byte {
	var b bytes.Buffer
	writeBytes(&b, r.Value)
	writeU32(&b, r.CellIndex)
	writeU32(&b, r.NewCellIndex)
	flags := byte(0)
	if r.HasValue {
		flags |= 0x01
	}
	if r.HasCellIndex {
		flags |= 0x02
	}
	if r.HasNewCell {
		flags |= 0x04
	}
	b.WriteByte(flags)
	return b.Bytes()
}

func writeU32(b *bytes.Buffer, v uint32) {
	var temp [4]byte
	binary.LittleEndian.PutUint32(temp[:], v)
	b.Write(temp[:])
}

func writeString(b *bytes.Buffer, v string) {
	writeBytes(b, []byte(v))
}

func writeBytes(b *bytes.Buffer, v []byte) {
	writeU32(b, uint32(len(v)))
	b.Write(v)
}
