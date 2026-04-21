// store.go
package main

import (
	"errors"
	ouroboros "github.com/DiegoSandival/ouroboros-go"
	bbolt "go.etcd.io/bbolt"
	"time"
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

