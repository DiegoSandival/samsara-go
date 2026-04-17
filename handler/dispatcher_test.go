package handler

import (
	"encoding/binary"
	"os"
	"strings"
	"testing"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
	samsara "github.com/DiegoSandival/samsara-go"
)

func TestCentralHandler_ManualFlowAllOpcodes(t *testing.T) {
	h := NewCentralHandler(t.TempDir())
	defer h.Close()

	id := makeRequestID("handler-test-0001")

	missingWrite := send(t, h, id, OpcodeWrite, BuildWritePayload("main", "alpha", []byte("v1"), 1, []byte("x")))
	assertStatus(t, missingWrite, StatusCodeUndefined)

	created := send(t, h, id, OpcodeCreateDB, BuildCreateDBPayload(60, "main", []byte("root-secret")))
	assertStatus(t, created, StatusCodeOK)

	store, ok := h.getStore("main")
	if !ok {
		t.Fatal("expected created db to be available")
	}

	secretA := []byte("secret-a")
	secretB := []byte("secret-b")
	indexA := appendCell(t, store, secretA, ouroboros.LeerSelf|ouroboros.LeerAny|ouroboros.LeerLibre|ouroboros.EscribirSelf|ouroboros.BorrarSelf|ouroboros.Diferir|ouroboros.Fucionar, [16]byte{'a'})
	indexB := appendCell(t, store, secretB, ouroboros.LeerSelf|ouroboros.Fucionar, [16]byte{'b'})

	writeResp := send(t, h, id, OpcodeWrite, BuildWritePayload("main", "alpha", []byte("v1"), indexA, secretA))
	assertStatus(t, writeResp, StatusCodeOK)
	newIndex, flags := decodeWriteLikePayload(t, writeResp.payload)
	if flags&0x02 == 0 {
		t.Fatal("expected write new cell flag")
	}
	indexA = newIndex

	readResp := send(t, h, id, OpcodeRead, BuildReadPayload("main", "alpha", indexA, secretA))
	assertStatus(t, readResp, StatusCodeOK)
	value, readNewIndex, readFlags := decodeReadPayload(t, readResp.payload)
	if string(value) != "v1" || readFlags&0x01 == 0 || readFlags&0x04 == 0 {
		t.Fatalf("unexpected read payload: value=%q flags=%08b", string(value), readFlags)
	}
	indexA = readNewIndex

	readFreeResp := send(t, h, id, OpcodeReadFree, BuildReadFreePayload("main", "alpha"))
	assertStatus(t, readFreeResp, StatusCodeOK)
	value, freeFlags := decodeReadFreePayload(t, readFreeResp.payload)
	if string(value) != "v1" || freeFlags&0x01 == 0 {
		t.Fatalf("unexpected read_free payload: value=%q flags=%08b", string(value), freeFlags)
	}

	readCellResp := send(t, h, id, OpcodeReadCell, BuildReadCellPayload("main", indexA, secretA))
	assertStatus(t, readCellResp, StatusCodeOK)
	if len(readCellResp.payload) != 69 {
		t.Fatalf("unexpected read_cell payload len: %d", len(readCellResp.payload))
	}

	var childSalt [16]byte
	copy(childSalt[:], []byte("child-salt-000001"))
	diferirResp := send(t, h, id, OpcodeDiferir, BuildDiferirPayload("main", indexA, secretA, []byte("child-secret"), childSalt, ouroboros.LeerSelf, 7, 8, 9))
	assertStatus(t, diferirResp, StatusCodeOK)
	indexA = decodeDiferirNewIndex(t, diferirResp.payload)

	var cruzarSalt [16]byte
	copy(cruzarSalt[:], []byte("cross-salt-000001"))
	cruzarResp := send(t, h, id, OpcodeCruzar, BuildCruzarPayload("main", indexA, secretA, indexB, secretB, []byte("cross-child"), cruzarSalt, 1, 2, 3))
	assertStatus(t, cruzarResp, StatusCodeOK)
	indexA = decodeCruzarNewIndexA(t, cruzarResp.payload)

	deleteResp := send(t, h, id, OpcodeDelete, BuildDeletePayload("main", "alpha", indexA, secretA))
	assertStatus(t, deleteResp, StatusCodeOK)

	deleted := send(t, h, id, OpcodeDeleteDB, BuildManageDBPayload("main", 0, nil))
	assertStatus(t, deleted, StatusCodeOK)

	missingRead := send(t, h, id, OpcodeRead, BuildReadPayload("main", "alpha", indexA, secretA))
	assertStatus(t, missingRead, StatusCodeUndefined)
}

func TestUnmarshalValidation(t *testing.T) {
	_, err := Unmarshal([]byte{OpcodeRead})
	if err == nil {
		t.Fatal("expected error for short message")
	}
}

func TestInvalidPayloadAndUnknownOpcode(t *testing.T) {
	h := NewCentralHandler(t.TempDir())
	defer h.Close()

	id := makeRequestID("invalid-000000001")

	badResp := send(t, h, id, OpcodeRead, []byte{0x01, 0x02})
	assertStatus(t, badResp, StatusCodeUndefined)
	if !strings.Contains(decodeErrorString(t, badResp.payload), "invalid payload") {
		t.Fatalf("unexpected payload error: %q", decodeErrorString(t, badResp.payload))
	}

	unknownResp := send(t, h, id, 0xFF, nil)
	assertStatus(t, unknownResp, StatusCodeUndefined)
	if !strings.Contains(decodeErrorString(t, unknownResp.payload), "unknown opcode") {
		t.Fatalf("unexpected unknown opcode payload: %q", decodeErrorString(t, unknownResp.payload))
	}
}

func TestCentralHandler_PersistAndReloadDBNames(t *testing.T) {
	baseDir := t.TempDir()
	h := NewCentralHandler(baseDir)
	defer h.Close()

	id := makeRequestID("persist-db-names")
	secret := []byte("root-secret")

	created := send(t, h, id, OpcodeCreateDB, BuildCreateDBPayload(64, "main", secret))
	assertStatus(t, created, StatusCodeOK)

	namesData, err := os.ReadFile(baseDir + "/db_names.txt")
	if err != nil {
		t.Fatalf("read db_names.txt: %v", err)
	}
	if string(namesData) != "main\n" {
		t.Fatalf("unexpected db_names.txt content: %q", string(namesData))
	}

	if err := h.Close(); err != nil {
		t.Fatalf("close handler: %v", err)
	}

	h2 := NewCentralHandler(baseDir)
	defer h2.Close()

	h2.mu.Lock()
	_, loaded := h2.stores["main"]
	h2.mu.Unlock()
	if !loaded {
		t.Fatal("expected main db to be loaded from db_names.txt")
	}

	deleted := send(t, h2, id, OpcodeDeleteDB, BuildManageDBPayload("main", 0, nil))
	assertStatus(t, deleted, StatusCodeOK)

	namesAfterDelete, err := os.ReadFile(baseDir + "/db_names.txt")
	if err != nil {
		t.Fatalf("read db_names.txt after delete: %v", err)
	}
	if len(namesAfterDelete) != 0 {
		t.Fatalf("expected empty db_names.txt after delete, got: %q", string(namesAfterDelete))
	}
}

type envelope struct {
	status  byte
	payload []byte
}

func send(t *testing.T, h *CentralHandler, id RequestID, opcode byte, payload []byte) envelope {
	t.Helper()
	raw := MarshalMessage(opcode, id, payload)
	resp := h.HandleRaw(raw)
	status, body, err := UnmarshalEnvelope(resp)
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return envelope{status: status, payload: body}
}

func assertStatus(t *testing.T, got envelope, want byte) {
	t.Helper()
	if got.status != want {
		t.Fatalf("unexpected status: got=%d want=%d payload=%v", got.status, want, got.payload)
	}
}

func makeRequestID(s string) RequestID {
	var id RequestID
	copy(id[:], []byte(s))
	return id
}

func appendCell(t *testing.T, store *samsara.Store, secret []byte, genome uint32, salt [16]byte) uint32 {
	t.Helper()
	idx, err := store.DB().Append(samsara.NewCellWithSecret(salt, secret, genome, 1, 2, 3))
	if err != nil {
		t.Fatalf("append cell: %v", err)
	}
	return idx
}

func decodeWriteLikePayload(t *testing.T, data []byte) (uint32, byte) {
	t.Helper()
	if len(data) != 9 {
		t.Fatalf("invalid write-like payload len: %d", len(data))
	}
	newIndex := binary.LittleEndian.Uint32(data[4:8])
	return newIndex, data[8]
}

func decodeReadPayload(t *testing.T, data []byte) ([]byte, uint32, byte) {
	t.Helper()
	if len(data) < 13 {
		t.Fatalf("invalid read payload len: %d", len(data))
	}
	valueLen := binary.LittleEndian.Uint32(data[:4])
	if len(data) < int(4+valueLen+9) {
		t.Fatalf("invalid read payload value len: %d", valueLen)
	}
	value := data[4 : 4+valueLen]
	newOffset := 4 + valueLen + 4
	newIndex := binary.LittleEndian.Uint32(data[newOffset : newOffset+4])
	flags := data[newOffset+4]
	return value, newIndex, flags
}

func decodeReadFreePayload(t *testing.T, data []byte) ([]byte, byte) {
	t.Helper()
	if len(data) < 5 {
		t.Fatalf("invalid read_free payload len: %d", len(data))
	}
	valueLen := binary.LittleEndian.Uint32(data[:4])
	if len(data) != int(4+valueLen+1) {
		t.Fatalf("invalid read_free payload value len: %d", valueLen)
	}
	return data[4 : 4+valueLen], data[len(data)-1]
}

func decodeDiferirNewIndex(t *testing.T, data []byte) uint32 {
	t.Helper()
	if len(data) != 13 {
		t.Fatalf("invalid diferir payload len: %d", len(data))
	}
	return binary.LittleEndian.Uint32(data[8:12])
}

func decodeCruzarNewIndexA(t *testing.T, data []byte) uint32 {
	t.Helper()
	if len(data) != 21 {
		t.Fatalf("invalid cruzar payload len: %d", len(data))
	}
	return binary.LittleEndian.Uint32(data[12:16])
}

func decodeErrorString(t *testing.T, data []byte) string {
	t.Helper()
	if len(data) < 4 {
		t.Fatalf("invalid error payload len: %d", len(data))
	}
	ln := binary.LittleEndian.Uint32(data[:4])
	if len(data) != int(4+ln) {
		t.Fatalf("invalid error payload string len: %d", ln)
	}
	return string(data[4:])
}
