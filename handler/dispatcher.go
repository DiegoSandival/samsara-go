package handler

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

const (
	requestHeaderSize = 17 // 1 byte opcode + 16 bytes request ID
	maxStringLen      = 1024
	maxSecretLen      = 4096
	maxValueLen       = 8 * 1024 * 1024
	defaultMaxRecords = 60
)

const (
	StatusCodeOK byte = iota
	StatusCodeUnauthorized
	StatusCodeUndefined
	StatusCodeErrorDB
)

type RequestID [16]byte

type Message struct {
	Opcode  byte
	ID      RequestID
	Payload []byte
}

type CentralHandler struct {
	baseDir string

	mu     sync.Mutex
	stores map[string]*samsara.Store
}

func NewCentralHandler(baseDir string) *CentralHandler {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "."
	}
	_ = os.MkdirAll(baseDir, 0755)

	return &CentralHandler{
		baseDir: baseDir,
		stores:  make(map[string]*samsara.Store),
	}
}

func (h *CentralHandler) Close() error {
	if h == nil {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	var closeErr error
	for name, store := range h.stores {
		if store == nil {
			continue
		}
		if err := store.Close(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("close db %q: %w", name, err))
		}
	}
	h.stores = make(map[string]*samsara.Store)
	return closeErr
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

func MarshalMessage(opcode byte, id RequestID, payload []byte) []byte {
	out := make([]byte, requestHeaderSize+len(payload))
	out[0] = opcode
	copy(out[1:17], id[:])
	copy(out[17:], payload)
	return out
}

func MarshalEnvelope(statusCode byte, payload []byte) []byte {
	out := make([]byte, 1+4+len(payload))
	out[0] = statusCode
	binary.LittleEndian.PutUint32(out[1:5], uint32(len(payload)))
	copy(out[5:], payload)
	return out
}

func UnmarshalEnvelope(data []byte) (byte, []byte, error) {
	if len(data) < 5 {
		return 0, nil, fmt.Errorf("envelope too short: %d", len(data))
	}
	status := data[0]
	ln := binary.LittleEndian.Uint32(data[1:5])
	if len(data)-5 != int(ln) {
		return 0, nil, fmt.Errorf("invalid envelope length: declared=%d actual=%d", ln, len(data)-5)
	}
	payload := append([]byte(nil), data[5:]...)
	return status, payload, nil
}

func (h *CentralHandler) HandleRaw(data []byte) []byte {
	msg, err := Unmarshal(data)
	if err != nil {
		return errorEnvelope(StatusCodeUndefined, "invalid message: "+err.Error())
	}
	return h.Handle(msg)
}

func (h *CentralHandler) Handle(msg Message) []byte {
	if h == nil {
		return errorEnvelope(StatusCodeErrorDB, "nil handler")
	}

	switch msg.Opcode {
	case OpcodeRead:
		p, err := parseReadPayload(msg.Payload)
		if err != nil {
			return errorEnvelope(StatusCodeUndefined, "invalid payload: "+err.Error())
		}
		store, ok := h.getStore(p.DBName)
		if !ok {
			return errorEnvelope(StatusCodeUndefined, "undefined db: "+p.DBName)
		}
		result := store.Read(p.Key, p.CellIndex, p.Secret)
		return MarshalEnvelope(statusCodeFromStatus(result.Status), encodeReadResult(result))
	case OpcodeReadFree:
		p, err := parseReadFreePayload(msg.Payload)
		if err != nil {
			return errorEnvelope(StatusCodeUndefined, "invalid payload: "+err.Error())
		}
		store, ok := h.getStore(p.DBName)
		if !ok {
			return errorEnvelope(StatusCodeUndefined, "undefined db: "+p.DBName)
		}
		result := store.ReadFree(p.Key)
		return MarshalEnvelope(statusCodeFromStatus(result.Status), encodeReadFreeResult(result))
	case OpcodeWrite:
		p, err := parseWritePayload(msg.Payload)
		if err != nil {
			return errorEnvelope(StatusCodeUndefined, "invalid payload: "+err.Error())
		}
		store, ok := h.getStore(p.DBName)
		if !ok {
			return errorEnvelope(StatusCodeUndefined, "undefined db: "+p.DBName)
		}
		result := store.Write(p.Key, p.Value, p.CellIndex, p.Secret)
		return MarshalEnvelope(statusCodeFromStatus(result.Status), encodeWriteResult(result))
	case OpcodeDelete:
		p, err := parseDeletePayload(msg.Payload)
		if err != nil {
			return errorEnvelope(StatusCodeUndefined, "invalid payload: "+err.Error())
		}
		store, ok := h.getStore(p.DBName)
		if !ok {
			return errorEnvelope(StatusCodeUndefined, "undefined db: "+p.DBName)
		}
		result := store.Delete(p.Key, p.CellIndex, p.Secret)
		return MarshalEnvelope(statusCodeFromStatus(result.Status), encodeDeleteResult(result))
	case OpcodeReadCell:
		p, err := parseReadCellPayload(msg.Payload)
		if err != nil {
			return errorEnvelope(StatusCodeUndefined, "invalid payload: "+err.Error())
		}
		store, ok := h.getStore(p.DBName)
		if !ok {
			return errorEnvelope(StatusCodeUndefined, "undefined db: "+p.DBName)
		}
		result := store.ReadCell(p.CellIndex, p.Secret)
		return MarshalEnvelope(statusCodeFromStatus(result.Status), encodeReadCellResult(result))
	case OpcodeDiferir:
		p, err := parseDiferirPayload(msg.Payload)
		if err != nil {
			return errorEnvelope(StatusCodeUndefined, "invalid payload: "+err.Error())
		}
		store, ok := h.getStore(p.DBName)
		if !ok {
			return errorEnvelope(StatusCodeUndefined, "undefined db: "+p.DBName)
		}
		result := store.Diferir(p.CellIndex, p.ParentSecret, p.ChildSecret, p.ChildSalt, p.ChildGenome, p.X, p.Y, p.Z)
		return MarshalEnvelope(statusCodeFromStatus(result.Status), encodeDiferirResult(result))
	case OpcodeCruzar:
		p, err := parseCruzarPayload(msg.Payload)
		if err != nil {
			return errorEnvelope(StatusCodeUndefined, "invalid payload: "+err.Error())
		}
		store, ok := h.getStore(p.DBName)
		if !ok {
			return errorEnvelope(StatusCodeUndefined, "undefined db: "+p.DBName)
		}
		result := store.Cruzar(p.CellIndexA, p.SecretA, p.CellIndexB, p.SecretB, p.ChildSecret, p.ChildSalt, p.X, p.Y, p.Z)
		return MarshalEnvelope(statusCodeFromStatus(result.Status), encodeCruzarResult(result))
	case OpcodeCreateDB:
		p, err := parseManageDBPayload(msg.Payload)
		if err != nil {
			return errorEnvelope(StatusCodeUndefined, "invalid payload: "+err.Error())
		}
		if err := h.createDB(p.DBName); err != nil {
			if errors.Is(err, os.ErrExist) {
				return errorEnvelope(StatusCodeUndefined, err.Error())
			}
			return errorEnvelope(StatusCodeErrorDB, err.Error())
		}
		return MarshalEnvelope(StatusCodeOK, encodeString(p.DBName))
	case OpcodeDeleteDB:
		p, err := parseManageDBPayload(msg.Payload)
		if err != nil {
			return errorEnvelope(StatusCodeUndefined, "invalid payload: "+err.Error())
		}
		if err := h.deleteDB(p.DBName); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return errorEnvelope(StatusCodeUndefined, err.Error())
			}
			return errorEnvelope(StatusCodeErrorDB, err.Error())
		}
		return MarshalEnvelope(StatusCodeOK, encodeString(p.DBName))
	default:
		return errorEnvelope(StatusCodeUndefined, fmt.Sprintf("unknown opcode: 0x%02x", msg.Opcode))
	}
}

func (h *CentralHandler) createDB(dbName string) error {
	if err := validateDBName(dbName); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.stores[dbName]; exists {
		return fmt.Errorf("db already exists: %w", os.ErrExist)
	}

	path := h.dbPath(dbName)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("db already exists: %w", os.ErrExist)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	store, err := samsara.New(path, defaultMaxRecords)
	if err != nil {
		return err
	}
	h.stores[dbName] = store
	return nil
}

func (h *CentralHandler) deleteDB(dbName string) error {
	if err := validateDBName(dbName); err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	path := h.dbPath(dbName)
	boltPath := path + ".bolt"

	store, open := h.stores[dbName]
	if open {
		if err := store.Close(); err != nil {
			return err
		}
		delete(h.stores, dbName)
	}

	hadFile := false
	if err := os.Remove(path); err == nil {
		hadFile = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.Remove(boltPath); err == nil {
		hadFile = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if !open && !hadFile {
		return fmt.Errorf("db does not exist: %w", os.ErrNotExist)
	}

	return nil
}

func (h *CentralHandler) getStore(dbName string) (*samsara.Store, bool) {
	if err := validateDBName(dbName); err != nil {
		return nil, false
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if store, ok := h.stores[dbName]; ok {
		return store, true
	}

	path := h.dbPath(dbName)
	if _, err := os.Stat(path); err != nil {
		return nil, false
	}

	store, err := samsara.New(path, defaultMaxRecords)
	if err != nil {
		return nil, false
	}
	h.stores[dbName] = store
	return store, true
}

func (h *CentralHandler) dbPath(dbName string) string {
	return filepath.Join(h.baseDir, dbName+".db")
}

func statusCodeFromStatus(status samsara.Status) byte {
	switch status {
	case samsara.StatusOK:
		return StatusCodeOK
	case samsara.StatusUnauthorized:
		return StatusCodeUnauthorized
	case samsara.StatusUndefined:
		return StatusCodeUndefined
	default:
		return StatusCodeErrorDB
	}
}

func errorEnvelope(statusCode byte, msg string) []byte {
	return MarshalEnvelope(statusCode, encodeString(msg))
}

func encodeString(v string) []byte {
	var b bytes.Buffer
	writeString(&b, v)
	return b.Bytes()
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

func encodeReadFreeResult(r samsara.FreeReadResult) []byte {
	var b bytes.Buffer
	writeBytes(&b, r.Value)
	flags := byte(0)
	if r.HasValue {
		flags |= 0x01
	}
	b.WriteByte(flags)
	return b.Bytes()
}

func encodeWriteResult(r samsara.WriteResult) []byte {
	var b bytes.Buffer
	writeU32(&b, r.CellIndex)
	writeU32(&b, r.NewCellIndex)
	flags := byte(0)
	if r.HasCellIndex {
		flags |= 0x01
	}
	if r.HasNewCell {
		flags |= 0x02
	}
	b.WriteByte(flags)
	return b.Bytes()
}

func encodeDeleteResult(r samsara.DeleteResult) []byte {
	var b bytes.Buffer
	writeU32(&b, r.CellIndex)
	writeU32(&b, r.NewCellIndex)
	flags := byte(0)
	if r.HasCellIndex {
		flags |= 0x01
	}
	if r.HasNewCell {
		flags |= 0x02
	}
	b.WriteByte(flags)
	return b.Bytes()
}

func encodeDiferirResult(r samsara.DiferirResult) []byte {
	var b bytes.Buffer
	writeU32(&b, r.CellIndex)
	writeU32(&b, r.DeferredIndex)
	writeU32(&b, r.NewCellIndex)
	flags := byte(0)
	if r.HasCellIndex {
		flags |= 0x01
	}
	if r.HasDeferred {
		flags |= 0x02
	}
	if r.HasNewCell {
		flags |= 0x04
	}
	b.WriteByte(flags)
	return b.Bytes()
}

func encodeCruzarResult(r samsara.CruzarResult) []byte {
	var b bytes.Buffer
	writeU32(&b, r.CellIndexA)
	writeU32(&b, r.CellIndexB)
	writeU32(&b, r.ChildIndex)
	writeU32(&b, r.NewCellIndexA)
	writeU32(&b, r.NewCellIndexB)
	flags := byte(0)
	if r.HasCellIndexA {
		flags |= 0x01
	}
	if r.HasCellIndexB {
		flags |= 0x02
	}
	if r.HasChild {
		flags |= 0x04
	}
	if r.HasNewCellA {
		flags |= 0x08
	}
	if r.HasNewCellB {
		flags |= 0x10
	}
	b.WriteByte(flags)
	return b.Bytes()
}

func encodeReadCellResult(r samsara.CellReadResult) []byte {
	var b bytes.Buffer
	b.Write(r.Cell.Hash[:])
	b.Write(r.Cell.Salt[:])
	writeU32(&b, r.Cell.Genoma)
	writeU32(&b, r.Cell.X)
	writeU32(&b, r.Cell.Y)
	writeU32(&b, r.Cell.Z)
	writeU32(&b, r.CellIndex)
	flags := byte(0)
	if r.HasCell {
		flags |= 0x01
	}
	if r.HasCellIndex {
		flags |= 0x02
	}
	b.WriteByte(flags)
	return b.Bytes()
}

func validateDBName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("dbName is required")
	}
	if len(name) > 128 {
		return errors.New("dbName is too long")
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return fmt.Errorf("dbName has invalid character: %q", r)
	}
	return nil
}

type payloadReader struct {
	data []byte
	off  int
}

func newPayloadReader(data []byte) *payloadReader {
	return &payloadReader{data: data}
}

func (r *payloadReader) readU32() (uint32, error) {
	if len(r.data)-r.off < 4 {
		return 0, errors.New("not enough bytes for uint32")
	}
	v := binary.LittleEndian.Uint32(r.data[r.off : r.off+4])
	r.off += 4
	return v, nil
}

func (r *payloadReader) readFixed(n int) ([]byte, error) {
	if n < 0 || len(r.data)-r.off < n {
		return nil, fmt.Errorf("not enough bytes for fixed field (%d)", n)
	}
	v := append([]byte(nil), r.data[r.off:r.off+n]...)
	r.off += n
	return v, nil
}

func (r *payloadReader) readLenBytes(max int, field string) ([]byte, error) {
	ln, err := r.readU32()
	if err != nil {
		return nil, fmt.Errorf("%s length: %w", field, err)
	}
	if ln > uint32(max) {
		return nil, fmt.Errorf("%s too large: %d > %d", field, ln, max)
	}
	if uint32(len(r.data)-r.off) < ln {
		return nil, fmt.Errorf("%s payload truncated", field)
	}
	v := append([]byte(nil), r.data[r.off:r.off+int(ln)]...)
	r.off += int(ln)
	return v, nil
}

func (r *payloadReader) readLenString(max int, field string) (string, error) {
	b, err := r.readLenBytes(max, field)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (r *payloadReader) ensureEOF() error {
	if r.off != len(r.data) {
		return fmt.Errorf("unexpected trailing bytes: %d", len(r.data)-r.off)
	}
	return nil
}

type readPayload struct {
	DBName    string
	Key       string
	CellIndex uint32
	Secret    []byte
}

type readFreePayload struct {
	DBName string
	Key    string
}

type writePayload struct {
	DBName    string
	Key       string
	Value     []byte
	CellIndex uint32
	Secret    []byte
}

type deletePayload struct {
	DBName    string
	Key       string
	CellIndex uint32
	Secret    []byte
}

type readCellPayload struct {
	DBName    string
	CellIndex uint32
	Secret    []byte
}

type diferirPayload struct {
	DBName       string
	CellIndex    uint32
	ParentSecret []byte
	ChildSecret  []byte
	ChildSalt    [16]byte
	ChildGenome  uint32
	X            uint32
	Y            uint32
	Z            uint32
}

type cruzarPayload struct {
	DBName      string
	CellIndexA  uint32
	SecretA     []byte
	CellIndexB  uint32
	SecretB     []byte
	ChildSecret []byte
	ChildSalt   [16]byte
	X           uint32
	Y           uint32
	Z           uint32
}

type manageDBPayload struct {
	DBName    string
	CellIndex uint32
	Secret    []byte
}

func parseReadPayload(data []byte) (readPayload, error) {
	r := newPayloadReader(data)
	dbName, err := r.readLenString(maxStringLen, "dbName")
	if err != nil {
		return readPayload{}, err
	}
	key, err := r.readLenString(maxStringLen, "key")
	if err != nil {
		return readPayload{}, err
	}
	cellIndex, err := r.readU32()
	if err != nil {
		return readPayload{}, err
	}
	secret, err := r.readLenBytes(maxSecretLen, "secret")
	if err != nil {
		return readPayload{}, err
	}
	if err := r.ensureEOF(); err != nil {
		return readPayload{}, err
	}
	return readPayload{DBName: dbName, Key: key, CellIndex: cellIndex, Secret: secret}, nil
}

func parseReadFreePayload(data []byte) (readFreePayload, error) {
	r := newPayloadReader(data)
	dbName, err := r.readLenString(maxStringLen, "dbName")
	if err != nil {
		return readFreePayload{}, err
	}
	key, err := r.readLenString(maxStringLen, "key")
	if err != nil {
		return readFreePayload{}, err
	}
	if err := r.ensureEOF(); err != nil {
		return readFreePayload{}, err
	}
	return readFreePayload{DBName: dbName, Key: key}, nil
}

func parseWritePayload(data []byte) (writePayload, error) {
	r := newPayloadReader(data)
	dbName, err := r.readLenString(maxStringLen, "dbName")
	if err != nil {
		return writePayload{}, err
	}
	key, err := r.readLenString(maxStringLen, "key")
	if err != nil {
		return writePayload{}, err
	}
	value, err := r.readLenBytes(maxValueLen, "value")
	if err != nil {
		return writePayload{}, err
	}
	cellIndex, err := r.readU32()
	if err != nil {
		return writePayload{}, err
	}
	secret, err := r.readLenBytes(maxSecretLen, "secret")
	if err != nil {
		return writePayload{}, err
	}
	if err := r.ensureEOF(); err != nil {
		return writePayload{}, err
	}
	return writePayload{DBName: dbName, Key: key, Value: value, CellIndex: cellIndex, Secret: secret}, nil
}

func parseDeletePayload(data []byte) (deletePayload, error) {
	r := newPayloadReader(data)
	dbName, err := r.readLenString(maxStringLen, "dbName")
	if err != nil {
		return deletePayload{}, err
	}
	key, err := r.readLenString(maxStringLen, "key")
	if err != nil {
		return deletePayload{}, err
	}
	cellIndex, err := r.readU32()
	if err != nil {
		return deletePayload{}, err
	}
	secret, err := r.readLenBytes(maxSecretLen, "secret")
	if err != nil {
		return deletePayload{}, err
	}
	if err := r.ensureEOF(); err != nil {
		return deletePayload{}, err
	}
	return deletePayload{DBName: dbName, Key: key, CellIndex: cellIndex, Secret: secret}, nil
}

func parseReadCellPayload(data []byte) (readCellPayload, error) {
	r := newPayloadReader(data)
	dbName, err := r.readLenString(maxStringLen, "dbName")
	if err != nil {
		return readCellPayload{}, err
	}
	cellIndex, err := r.readU32()
	if err != nil {
		return readCellPayload{}, err
	}
	secret, err := r.readLenBytes(maxSecretLen, "secret")
	if err != nil {
		return readCellPayload{}, err
	}
	if err := r.ensureEOF(); err != nil {
		return readCellPayload{}, err
	}
	return readCellPayload{DBName: dbName, CellIndex: cellIndex, Secret: secret}, nil
}

func parseDiferirPayload(data []byte) (diferirPayload, error) {
	r := newPayloadReader(data)
	dbName, err := r.readLenString(maxStringLen, "dbName")
	if err != nil {
		return diferirPayload{}, err
	}
	cellIndex, err := r.readU32()
	if err != nil {
		return diferirPayload{}, err
	}
	parentSecret, err := r.readLenBytes(maxSecretLen, "parentSecret")
	if err != nil {
		return diferirPayload{}, err
	}
	childSecret, err := r.readLenBytes(maxSecretLen, "childSecret")
	if err != nil {
		return diferirPayload{}, err
	}
	saltBytes, err := r.readFixed(16)
	if err != nil {
		return diferirPayload{}, err
	}
	var childSalt [16]byte
	copy(childSalt[:], saltBytes)
	childGenome, err := r.readU32()
	if err != nil {
		return diferirPayload{}, err
	}
	x, err := r.readU32()
	if err != nil {
		return diferirPayload{}, err
	}
	y, err := r.readU32()
	if err != nil {
		return diferirPayload{}, err
	}
	z, err := r.readU32()
	if err != nil {
		return diferirPayload{}, err
	}
	if err := r.ensureEOF(); err != nil {
		return diferirPayload{}, err
	}
	return diferirPayload{
		DBName:       dbName,
		CellIndex:    cellIndex,
		ParentSecret: parentSecret,
		ChildSecret:  childSecret,
		ChildSalt:    childSalt,
		ChildGenome:  childGenome,
		X:            x,
		Y:            y,
		Z:            z,
	}, nil
}

func parseCruzarPayload(data []byte) (cruzarPayload, error) {
	r := newPayloadReader(data)
	dbName, err := r.readLenString(maxStringLen, "dbName")
	if err != nil {
		return cruzarPayload{}, err
	}
	cellIndexA, err := r.readU32()
	if err != nil {
		return cruzarPayload{}, err
	}
	secretA, err := r.readLenBytes(maxSecretLen, "secretA")
	if err != nil {
		return cruzarPayload{}, err
	}
	cellIndexB, err := r.readU32()
	if err != nil {
		return cruzarPayload{}, err
	}
	secretB, err := r.readLenBytes(maxSecretLen, "secretB")
	if err != nil {
		return cruzarPayload{}, err
	}
	childSecret, err := r.readLenBytes(maxSecretLen, "childSecret")
	if err != nil {
		return cruzarPayload{}, err
	}
	saltBytes, err := r.readFixed(16)
	if err != nil {
		return cruzarPayload{}, err
	}
	var childSalt [16]byte
	copy(childSalt[:], saltBytes)
	x, err := r.readU32()
	if err != nil {
		return cruzarPayload{}, err
	}
	y, err := r.readU32()
	if err != nil {
		return cruzarPayload{}, err
	}
	z, err := r.readU32()
	if err != nil {
		return cruzarPayload{}, err
	}
	if err := r.ensureEOF(); err != nil {
		return cruzarPayload{}, err
	}
	return cruzarPayload{
		DBName:      dbName,
		CellIndexA:  cellIndexA,
		SecretA:     secretA,
		CellIndexB:  cellIndexB,
		SecretB:     secretB,
		ChildSecret: childSecret,
		ChildSalt:   childSalt,
		X:           x,
		Y:           y,
		Z:           z,
	}, nil
}

func parseManageDBPayload(data []byte) (manageDBPayload, error) {
	r := newPayloadReader(data)
	dbName, err := r.readLenString(maxStringLen, "dbName")
	if err != nil {
		return manageDBPayload{}, err
	}
	cellIndex, err := r.readU32()
	if err != nil {
		return manageDBPayload{}, err
	}
	secret, err := r.readLenBytes(maxSecretLen, "secret")
	if err != nil {
		return manageDBPayload{}, err
	}
	if err := r.ensureEOF(); err != nil {
		return manageDBPayload{}, err
	}
	return manageDBPayload{DBName: dbName, CellIndex: cellIndex, Secret: secret}, nil
}

func BuildReadPayload(dbName, key string, cellIndex uint32, secret []byte) []byte {
	var b bytes.Buffer
	writeString(&b, dbName)
	writeString(&b, key)
	writeU32(&b, cellIndex)
	writeBytes(&b, secret)
	return b.Bytes()
}

func BuildReadFreePayload(dbName, key string) []byte {
	var b bytes.Buffer
	writeString(&b, dbName)
	writeString(&b, key)
	return b.Bytes()
}

func BuildWritePayload(dbName, key string, value []byte, cellIndex uint32, secret []byte) []byte {
	var b bytes.Buffer
	writeString(&b, dbName)
	writeString(&b, key)
	writeBytes(&b, value)
	writeU32(&b, cellIndex)
	writeBytes(&b, secret)
	return b.Bytes()
}

func BuildDeletePayload(dbName, key string, cellIndex uint32, secret []byte) []byte {
	var b bytes.Buffer
	writeString(&b, dbName)
	writeString(&b, key)
	writeU32(&b, cellIndex)
	writeBytes(&b, secret)
	return b.Bytes()
}

func BuildReadCellPayload(dbName string, cellIndex uint32, secret []byte) []byte {
	var b bytes.Buffer
	writeString(&b, dbName)
	writeU32(&b, cellIndex)
	writeBytes(&b, secret)
	return b.Bytes()
}

func BuildDiferirPayload(dbName string, cellIndex uint32, parentSecret, childSecret []byte, childSalt [16]byte, childGenome, x, y, z uint32) []byte {
	var b bytes.Buffer
	writeString(&b, dbName)
	writeU32(&b, cellIndex)
	writeBytes(&b, parentSecret)
	writeBytes(&b, childSecret)
	b.Write(childSalt[:])
	writeU32(&b, childGenome)
	writeU32(&b, x)
	writeU32(&b, y)
	writeU32(&b, z)
	return b.Bytes()
}

func BuildCruzarPayload(dbName string, cellIndexA uint32, secretA []byte, cellIndexB uint32, secretB []byte, childSecret []byte, childSalt [16]byte, x, y, z uint32) []byte {
	var b bytes.Buffer
	writeString(&b, dbName)
	writeU32(&b, cellIndexA)
	writeBytes(&b, secretA)
	writeU32(&b, cellIndexB)
	writeBytes(&b, secretB)
	writeBytes(&b, childSecret)
	b.Write(childSalt[:])
	writeU32(&b, x)
	writeU32(&b, y)
	writeU32(&b, z)
	return b.Bytes()
}

func BuildManageDBPayload(dbName string, cellIndex uint32, secret []byte) []byte {
	var b bytes.Buffer
	writeString(&b, dbName)
	writeU32(&b, cellIndex)
	writeBytes(&b, secret)
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
