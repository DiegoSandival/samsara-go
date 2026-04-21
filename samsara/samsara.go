package main

import (
	"errors"
	"sync"
	"time"

	protocol "samsara/protocol"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
	bbolt "go.etcd.io/bbolt"
)

var Opcode byte

func main() {

	parser := &protocol.ProtocolParser{}

	CentralHandler := &CentralHandler{
		baseDir: "data",
		stores:  make(map[string]*Store),
	}

	handleOpcode([]byte{0x08}, parser, CentralHandler)

}

func handleOpcode(msg []byte, parser *protocol.ProtocolParser, handler *CentralHandler) []byte {
	switch Opcode {
	case 0x00:
		// Aquí se llamaría a CentralHandler.CreateDB, pasando el parser y los datos recibidos.
		// Ejemplo:
		// response := CentralHandler.CreateDB(parser, receivedData)
		// Luego se enviaría 'response' de vuelta al cliente.

		handler.CreateDB(parser, msg[4:]) // Pasamos el mensaje sin el opcode (primeros 4 bytes)
	default:
		// Manejar otros opcodes o enviar un error de opcode desconocido.
	}
}

var (
	ErrNilDB       = errors.New("nil ouroboros db")
	ErrNilKV       = errors.New("nil membrane db")
	membraneBucket = []byte("membranes")
)

type Store struct {
	db         *ouroboros.OuroborosDB
	kv         *bbolt.DB
	tempKVPath string
}

// New inicializa el Store creando una nueva DB con un límite de registros.
func New(path string, maxRecords uint32) (*Store, error) {
	db, err := ouroboros.OpenOuroborosDB(path, maxRecords)

	if err != nil {
		return nil, err
	}

	// Hacemos inline de la apertura de bbolt para evitar funciones extra.
	kv, err := bbolt.Open(path+".bolt", 0600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	store, err := NewWithDBAndKV(db, kv)
	if err != nil {
		// FIX: Ahora cerramos correctamente ambas conexiones si falla la creación del bucket.
		_ = db.Close()
		_ = kv.Close()
		return nil, err
	}

	return store, nil
}

// Open abre un Store existente (pasando 0 a maxRecords).
func Open(path string) (*Store, error) {
	db, err := ouroboros.OpenExistingOuroborosDB(path)

	if err != nil {
		return nil, err
	}

	// Hacemos inline de la apertura de bbolt para evitar funciones extra.
	kv, err := bbolt.Open(path+".bolt", 0600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	store, err := NewWithDBAndKV(db, kv)
	if err != nil {
		// FIX: Ahora cerramos correctamente ambas conexiones si falla la creación del bucket.
		_ = db.Close()
		_ = kv.Close()
		return nil, err
	}

	return store, nil
}

// NewWithDBAndKV inicializa el Store a partir de instancias ya abiertas.
func NewWithDBAndKV(db *ouroboros.OuroborosDB, kv *bbolt.DB) (*Store, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	if kv == nil {
		return nil, ErrNilKV
	}

	// Hacemos inline de ensureMembraneBucket para reducir la verbosidad.
	err := kv.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(membraneBucket)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &Store{db: db, kv: kv}, nil
}

func (s *Store) DB() *ouroboros.OuroborosDB {
	if s == nil {
		return nil
	}

	return s.db
}

func (s *Store) KV() *bbolt.DB {
	if s == nil {
		return nil
	}

	return s.kv
}

type CentralHandler struct {
	baseDir string

	mu     sync.Mutex
	stores map[string]*Store
}

func (h *CentralHandler) RegisterStore(name string, store *Store) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stores[name] = store
}

func (h *CentralHandler) CreateDB(parser *protocol.ProtocolParser, data []byte) []byte {

	protocol, err := parser.CreateDBReq(data)
	if err != nil {
		return protocol.CreateDBResponse{Success: false, Error: err.Error()}.Serialize()
	}

	store, err := New(protocol.DBName, protocol.MaxRecords)
	if err != nil {
		return protocol.CreateDBResponse{Success: false, Error: err.Error()}.Serialize()
	}

	h.RegisterStore(protocol.DBName, store)

	return parser.CreateDBResultByte(protocol.ID, 1)

}
