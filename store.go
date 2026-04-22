// store.go
package samsara

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"os"
	"time"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
	bbolt "go.etcd.io/bbolt"

	"lukechampine.com/blake3"
)

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

func DeleteDB(path string) error {

	// Primero intentamos eliminar el archivo de la base de datos principal
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Luego intentamos eliminar el archivo de la base de datos de membranas
	err = os.Remove(path + ".bolt")
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Se renombra de New a NewStore para evitar confusiones en el paquete main
func NewStore(path string, maxRecords uint32) (*Store, error) {
	db, err := ouroboros.OpenOuroborosDB(path, maxRecords)
	if err != nil {
		return nil, err
	}

	kv, err := bbolt.Open(path+".bolt", 0600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	store, err := NewWithDBAndKV(db, kv)
	if err != nil {
		_ = db.Close()
		_ = kv.Close()
		return nil, err
	}

	return store, nil
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

func (s *Store) NewCellWithSecret(salt [16]byte, secret []byte, genome, x, y, z uint32) ouroboros.Celula {
	data := append(salt[:], secret...)
	hash := blake3.Sum256(data)

	return ouroboros.Celula{
		Hash:   hash,
		Salt:   salt,
		Genoma: genome,
		X:      x,
		Y:      y,
		Z:      z,
	}
}

func (s *Store) readCellAuth(cellIndex uint32, secret []byte) (ouroboros.Celula, bool) {
	cell, err := s.db.ReadAuth(cellIndex, secret)
	if err != nil {
		return ouroboros.Celula{}, false
	}

	return cell, true
}

func (s *Store) readSystemCell(cellIndex uint32) (ouroboros.Celula, bool) {
	cell, err := s.db.Read(cellIndex)
	if err != nil {
		return ouroboros.Celula{}, false
	}

	return cell, true
}

type resolvedCell struct {
	cell          ouroboros.Celula
	index         uint32
	originalIndex uint32
}

func (s *Store) resolveCell(cellIndex uint32, secret []byte) (resolvedCell, bool) {
	return s.resolveCellWithReader(cellIndex, func(index uint32) (ouroboros.Celula, bool) {
		return s.readCellAuth(index, secret)
	})
}

func (s *Store) resolveCellWithReader(cellIndex uint32, reader func(uint32) (ouroboros.Celula, bool)) (resolvedCell, bool) {
	index := cellIndex
	cell, ok := reader(index)
	if !ok {
		return resolvedCell{}, false
	}

	for cell.Genoma&ouroboros.Migrada != 0 {
		index = cell.X
		cell, ok = reader(index)
		if !ok {
			return resolvedCell{}, false
		}
	}

	return resolvedCell{
		cell:          cell,
		index:         index,
		originalIndex: cellIndex,
	}, true
}

type Membrane struct {
	OwnerIndex uint32
	Value      []byte
}

func (s *Store) getMembrane(key string) (Membrane, bool, error) {
	if s == nil || s.kv == nil {
		return Membrane{}, false, ErrNilKV
	}

	var membrane Membrane
	var exists bool
	err := s.kv.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(membraneBucket)
		if bucket == nil {
			return ErrNilKV
		}

		value := bucket.Get([]byte(key))
		if value == nil {
			return nil
		}

		decoded, err := decodeMembrane(value)
		if err != nil {
			return err
		}
		membrane = decoded
		exists = true
		return nil
	})
	if err != nil {
		return Membrane{}, false, err
	}

	return membrane, exists, nil
}

func encodeMembrane(membrane Membrane) []byte {
	encoded := make([]byte, 4+len(membrane.Value))
	binary.LittleEndian.PutUint32(encoded[:4], membrane.OwnerIndex)
	copy(encoded[4:], membrane.Value)
	return encoded
}

func decodeMembrane(data []byte) (Membrane, error) {
	if len(data) < 4 {
		return Membrane{}, errors.New("invalid membrane payload")
	}

	return Membrane{
		OwnerIndex: binary.LittleEndian.Uint32(data[:4]),
		Value:      cloneBytes(data[4:]),
	}, nil
}

func cloneBytes(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	return bytes.Clone(data)
}

func (s *Store) refresh(cellIndex uint32, secret []byte) (uint32, bool) {
	original, ok := s.readSystemCell(cellIndex)
	if !ok {
		return 0, false
	}

	renewed, ok := refreshCellWithSecret(original, secret)
	if !ok {
		return 0, false
	}

	renewedIndex, err := s.db.Append(renewed)
	if err != nil {
		return 0, false
	}

	migratedGenome := original.Genoma | ouroboros.Migrada
	if err := s.db.Update(cellIndex, migratedGenome, renewedIndex, original.Y, original.Z); err != nil {
		return 0, false
	}

	return renewedIndex, true
}

func refreshCellWithSecret(cell ouroboros.Celula, secret []byte) (ouroboros.Celula, bool) {
	var salt [16]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return ouroboros.Celula{}, false
	}

	return NewCellWithSecret(salt, secret, cell.Genoma, cell.X, cell.Y, cell.Z), true
}
func NewCellWithSecret(salt [16]byte, secret []byte, genome, x, y, z uint32) ouroboros.Celula {
	data := append(salt[:], secret...)
	hash := blake3.Sum256(data)

	return ouroboros.Celula{
		Hash:   hash,
		Salt:   salt,
		Genoma: genome,
		X:      x,
		Y:      y,
		Z:      z,
	}
}

func (s *Store) permissionFlag(ownerIndex uint32, activeIndex uint32, selfFlag uint32, anyFlag uint32) uint32 {
	if s.isOwner(ownerIndex, activeIndex) {
		return selfFlag
	}

	return anyFlag
}

func (s *Store) isOwner(ownerIndex uint32, activeIndex uint32) bool {
	owner, ok := s.resolveSystemCell(ownerIndex)
	if !ok {
		return false
	}

	return owner.index == activeIndex
}

func (s *Store) resolveSystemCell(cellIndex uint32) (resolvedCell, bool) {
	return s.resolveCellWithReader(cellIndex, s.readSystemCell)
}

func (s *Store) putMembrane(key string, membrane Membrane) error {
	if s == nil || s.kv == nil {
		return ErrNilKV
	}

	encoded := encodeMembrane(membrane)
	return s.kv.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(membraneBucket)
		if bucket == nil {
			return ErrNilKV
		}

		return bucket.Put([]byte(key), encoded)
	})
}
func (s *Store) Close() error {
	if s == nil {
		return nil
	}

	var closeErr error
	if s.kv != nil {
		closeErr = errors.Join(closeErr, s.kv.Close())
	}
	if s.db != nil {
		closeErr = errors.Join(closeErr, s.db.Close())
	}
	if s.tempKVPath != "" {
		closeErr = errors.Join(closeErr, os.Remove(s.tempKVPath))
	}

	return closeErr
}
