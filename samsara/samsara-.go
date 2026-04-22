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

type Status string

const (
	StatusOK           Status = "ok"
	StatusUnauthorized Status = "unauthorized"
	StatusUndefined    Status = "undefined"
	StatusErrorDB      Status = "error_db"
)

var (
	ErrNilDB       = errors.New("nil ouroboros db")
	ErrNilKV       = errors.New("nil membrane db")
	membraneBucket = []byte("membranes")
)

const tempMembranePattern = "samsara-*.bolt"

type Membrane struct {
	OwnerIndex uint32
	Value      []byte
}

type Store struct {
	db         *ouroboros.OuroborosDB
	kv         *bbolt.DB
	tempKVPath string
}

type resolvedCell struct {
	cell          ouroboros.Celula
	index         uint32
	originalIndex uint32
}

type ReadResult struct {
	Status       Status
	Value        []byte
	CellIndex    uint32
	NewCellIndex uint32
	HasCellIndex bool
	HasNewCell   bool
	HasValue     bool
}

type FreeReadResult struct {
	Status   Status
	Value    []byte
	HasValue bool
}

type WriteResult struct {
	Status       Status
	CellIndex    uint32
	NewCellIndex uint32
	HasCellIndex bool
	HasNewCell   bool
}

type DeleteResult struct {
	Status       Status
	CellIndex    uint32
	NewCellIndex uint32
	HasCellIndex bool
	HasNewCell   bool
}

type DiferirResult struct {
	Status        Status
	CellIndex     uint32
	DeferredIndex uint32
	NewCellIndex  uint32
	HasCellIndex  bool
	HasDeferred   bool
	HasNewCell    bool
}

type CruzarResult struct {
	Status        Status
	CellIndexA    uint32
	CellIndexB    uint32
	ChildIndex    uint32
	NewCellIndexA uint32
	NewCellIndexB uint32
	HasCellIndexA bool
	HasCellIndexB bool
	HasChild      bool
	HasNewCellA   bool
	HasNewCellB   bool
}

type CellReadResult struct {
	Status       Status
	Cell         ouroboros.Celula
	CellIndex    uint32
	HasCell      bool
	HasCellIndex bool
}

func New(path string, maxRecords uint32) (*Store, error) {
	db, err := ouroboros.OpenOuroborosDB(path, maxRecords)

	if err != nil {
		return nil, err
	}

	kv, err := openMembraneDB(path + ".bolt")
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return NewWithDBAndKV(db, kv)
}

func Open(path string) (*Store, error) {
	db, err := ouroboros.OpenOuroborosDB(path, 0)
	if err != nil {
		return nil, err
	}

	kv, err := openMembraneDB(path + ".bolt")
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return NewWithDBAndKV(db, kv)
}

func NewWithDB(db *ouroboros.OuroborosDB) (*Store, error) {
	if db == nil {
		return nil, ErrNilDB
	}

	kv, tempPath, err := openTempMembraneDB()
	if err != nil {
		return nil, err
	}

	store, err := NewWithDBAndKV(db, kv)
	if err != nil {
		_ = kv.Close()
		_ = os.Remove(tempPath)
		return nil, err
	}

	store.tempKVPath = tempPath
	return store, nil
}

func NewWithDBAndKV(db *ouroboros.OuroborosDB, kv *bbolt.DB) (*Store, error) {
	if db == nil {
		return nil, ErrNilDB
	}
	if kv == nil {
		return nil, ErrNilKV
	}
	if err := ensureMembraneBucket(kv); err != nil {
		return nil, err
	}

	return &Store{db: db, kv: kv}, nil
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

func (s *Store) Read(key string, cellIndex uint32, secret []byte) ReadResult {
	active, ok := s.resolveCell(cellIndex, secret)
	if !ok {
		return ReadResult{Status: StatusUnauthorized}
	}

	membrane, exists, err := s.getMembrane(key)
	if err != nil {
		return ReadResult{Status: StatusErrorDB}
	}
	if !exists {
		return ReadResult{
			Status:       StatusUndefined,
			CellIndex:    active.index,
			HasCellIndex: true,
		}
	}

	requiredFlag := s.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.LeerSelf, ouroboros.LeerAny)

	if active.cell.Genoma&requiredFlag == 0 {
		return ReadResult{
			Status:       StatusUnauthorized,
			CellIndex:    active.index,
			HasCellIndex: true,
		}
	}

	newIndex, refreshed := s.refresh(active.index, secret)
	if !refreshed {
		return ReadResult{Status: StatusErrorDB}
	}

	return ReadResult{
		Status:       StatusOK,
		Value:        cloneBytes(membrane.Value),
		NewCellIndex: newIndex,
		HasValue:     true,
		HasNewCell:   true,
	}
}

func (s *Store) ReadFree(key string) FreeReadResult {
	membrane, exists, err := s.getMembrane(key)
	if err != nil {
		return FreeReadResult{Status: StatusErrorDB}
	}
	if !exists {
		return FreeReadResult{Status: StatusUndefined}
	}

	owner, ok := s.resolveSystemCell(membrane.OwnerIndex)
	if !ok || owner.cell.Genoma&ouroboros.LeerLibre == 0 {
		return FreeReadResult{Status: StatusUnauthorized}
	}

	return FreeReadResult{
		Status:   StatusOK,
		Value:    cloneBytes(membrane.Value),
		HasValue: true,
	}
}

func (s *Store) Write(key string, value []byte, cellIndex uint32, secret []byte) WriteResult {
	active, ok := s.resolveCell(cellIndex, secret)
	if !ok {
		return WriteResult{Status: StatusUnauthorized}
	}

	membrane, exists, err := s.getMembrane(key)
	if err != nil {
		return WriteResult{Status: StatusErrorDB}
	}
	if !exists {
		err = s.putMembrane(key, Membrane{
			OwnerIndex: active.originalIndex,
			Value:      cloneBytes(value),
		})
		if err != nil {
			return WriteResult{Status: StatusErrorDB}
		}

		newIndex, refreshed := s.refresh(active.index, secret)
		if !refreshed {
			return WriteResult{Status: StatusErrorDB}
		}

		return WriteResult{
			Status:       StatusOK,
			NewCellIndex: newIndex,
			HasNewCell:   true,
		}
	}

	requiredFlag := s.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.EscribirSelf, ouroboros.EscribirAny)

	if active.cell.Genoma&requiredFlag == 0 {
		return WriteResult{
			Status:       StatusUnauthorized,
			CellIndex:    active.index,
			HasCellIndex: true,
		}
	}

	membrane.Value = cloneBytes(value)
	if err := s.putMembrane(key, membrane); err != nil {
		return WriteResult{Status: StatusErrorDB}
	}

	newIndex, refreshed := s.refresh(active.index, secret)
	if !refreshed {
		return WriteResult{Status: StatusErrorDB}
	}

	return WriteResult{
		Status:       StatusOK,
		NewCellIndex: newIndex,
		HasNewCell:   true,
	}
}

func (s *Store) Delete(key string, cellIndex uint32, secret []byte) DeleteResult {
	active, ok := s.resolveCell(cellIndex, secret)
	if !ok {
		return DeleteResult{Status: StatusUnauthorized}
	}

	membrane, exists, err := s.getMembrane(key)
	if err != nil {
		return DeleteResult{Status: StatusErrorDB}
	}
	if !exists {
		return DeleteResult{
			Status:       StatusUndefined,
			CellIndex:    active.index,
			HasCellIndex: true,
		}
	}

	requiredFlag := s.permissionFlag(membrane.OwnerIndex, active.index, ouroboros.BorrarSelf, ouroboros.BorrarAny)

	if active.cell.Genoma&requiredFlag == 0 {
		return DeleteResult{
			Status:       StatusUnauthorized,
			CellIndex:    active.index,
			HasCellIndex: true,
		}
	}

	if err := s.deleteMembrane(key); err != nil {
		return DeleteResult{Status: StatusErrorDB}
	}

	newIndex, refreshed := s.refresh(active.index, secret)
	if !refreshed {
		return DeleteResult{Status: StatusErrorDB}
	}

	return DeleteResult{
		Status:       StatusOK,
		NewCellIndex: newIndex,
		HasNewCell:   true,
	}
}

func (s *Store) Diferir(cellIndex uint32, parentSecret []byte, childSecret []byte, childSalt [16]byte, childGenome, x, y, z uint32) DiferirResult {
	active, ok := s.resolveCell(cellIndex, parentSecret)
	if !ok {
		return DiferirResult{Status: StatusUnauthorized}
	}

	if active.cell.Genoma&ouroboros.Diferir == 0 {
		return DiferirResult{
			Status:       StatusUnauthorized,
			CellIndex:    active.index,
			HasCellIndex: true,
		}
	}

	if active.cell.Genoma&childGenome != childGenome {
		return DiferirResult{
			Status:       StatusUnauthorized,
			CellIndex:    active.index,
			HasCellIndex: true,
		}
	}

	child := NewCellWithSecret(childSalt, childSecret, childGenome, x, y, z)
	childIndex, err := s.db.Append(child)
	if err != nil {
		return DiferirResult{Status: StatusErrorDB}
	}

	newParentIndex, refreshed := s.refresh(active.index, parentSecret)
	if !refreshed {
		return DiferirResult{Status: StatusErrorDB}
	}

	return DiferirResult{
		Status:        StatusOK,
		DeferredIndex: childIndex,
		NewCellIndex:  newParentIndex,
		HasDeferred:   true,
		HasNewCell:    true,
	}
}

func (s *Store) Cruzar(cellIndexA uint32, secretA []byte, cellIndexB uint32, secretB []byte, childSecret []byte, childSalt [16]byte, x, y, z uint32) CruzarResult {
	activeA, okA := s.resolveCell(cellIndexA, secretA)
	activeB, okB := s.resolveCell(cellIndexB, secretB)
	if !okA || !okB {
		return CruzarResult{Status: StatusUnauthorized}
	}

	if activeA.cell.Genoma&ouroboros.Fucionar == 0 || activeB.cell.Genoma&ouroboros.Fucionar == 0 {
		return CruzarResult{
			Status:        StatusUnauthorized,
			CellIndexA:    activeA.index,
			CellIndexB:    activeB.index,
			HasCellIndexA: true,
			HasCellIndexB: true,
		}
	}

	childGenome := activeA.cell.Genoma | activeB.cell.Genoma
	child := NewCellWithSecret(childSalt, childSecret, childGenome, x, y, z)
	childIndex, err := s.db.Append(child)
	if err != nil {
		return CruzarResult{Status: StatusErrorDB}
	}

	newIndexA, okNewA := s.refresh(activeA.index, secretA)
	newIndexB, okNewB := s.refresh(activeB.index, secretB)
	if !okNewA || !okNewB {
		return CruzarResult{Status: StatusErrorDB}
	}

	return CruzarResult{
		Status:        StatusOK,
		ChildIndex:    childIndex,
		NewCellIndexA: newIndexA,
		NewCellIndexB: newIndexB,
		HasChild:      true,
		HasNewCellA:   true,
		HasNewCellB:   true,
	}
}

func (s *Store) ReadCell(cellIndex uint32, secret []byte) CellReadResult {
	active, ok := s.resolveCell(cellIndex, secret)
	if !ok {
		return CellReadResult{Status: StatusUnauthorized}
	}

	return CellReadResult{
		Status:       StatusOK,
		Cell:         active.cell,
		CellIndex:    active.index,
		HasCell:      true,
		HasCellIndex: true,
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

func (s *Store) resolveCell(cellIndex uint32, secret []byte) (resolvedCell, bool) {
	return s.resolveCellWithReader(cellIndex, func(index uint32) (ouroboros.Celula, bool) {
		return s.readCellAuth(index, secret)
	})
}

func (s *Store) resolveSystemCell(cellIndex uint32) (resolvedCell, bool) {
	return s.resolveCellWithReader(cellIndex, s.readSystemCell)
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

func openMembraneDB(path string) (*bbolt.DB, error) {
	kv, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	return kv, nil
}

func openTempMembraneDB() (*bbolt.DB, string, error) {
	tempFile, err := os.CreateTemp("", tempMembranePattern)
	if err != nil {
		return nil, "", err
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return nil, "", err
	}

	kv, err := openMembraneDB(tempPath)
	if err != nil {
		_ = os.Remove(tempPath)
		return nil, "", err
	}

	return kv, tempPath, nil
}

func ensureMembraneBucket(kv *bbolt.DB) error {
	return kv.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(membraneBucket)
		return err
	})
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

func (s *Store) deleteMembrane(key string) error {
	if s == nil || s.kv == nil {
		return ErrNilKV
	}

	return s.kv.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(membraneBucket)
		if bucket == nil {
			return ErrNilKV
		}

		return bucket.Delete([]byte(key))
	})
}

func (s *Store) isOwner(ownerIndex uint32, activeIndex uint32) bool {
	owner, ok := s.resolveSystemCell(ownerIndex)
	if !ok {
		return false
	}

	return owner.index == activeIndex
}

func (s *Store) permissionFlag(ownerIndex uint32, activeIndex uint32, selfFlag uint32, anyFlag uint32) uint32 {
	if s.isOwner(ownerIndex, activeIndex) {
		return selfFlag
	}

	return anyFlag
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
