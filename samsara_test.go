package samsara

import (
	"bytes"
	"testing"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
)

func TestWriteReadDeleteFlow(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	secret := []byte("owner-secret")
	index := appendTestCell(t, store, secret, ouroboros.LeerSelf|ouroboros.EscribirSelf|ouroboros.BorrarSelf)

	write := store.Write("alpha", []byte("uno"), index, secret)
	if write.Status != StatusOK || !write.HasNewCell {
		t.Fatalf("write failed: %+v", write)
	}

	read := store.Read("alpha", write.NewCellIndex, secret)
	if read.Status != StatusOK || !read.HasValue || !bytes.Equal(read.Value, []byte("uno")) {
		t.Fatalf("read failed: %+v", read)
	}

	deleted := store.Delete("alpha", read.NewCellIndex, secret)
	if deleted.Status != StatusOK || !deleted.HasNewCell {
		t.Fatalf("delete failed: %+v", deleted)
	}

	missing := store.Read("alpha", deleted.NewCellIndex, secret)
	if missing.Status != StatusUndefined || !missing.HasCellIndex {
		t.Fatalf("expected undefined read, got: %+v", missing)
	}
}

func TestReadFreeUsesOwnerGenome(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	secret := []byte("free-secret")
	index := appendTestCell(t, store, secret, ouroboros.LeerLibre|ouroboros.EscribirSelf)

	write := store.Write("public", []byte("visible"), index, secret)
	if write.Status != StatusOK {
		t.Fatalf("write failed: %+v", write)
	}

	read := store.ReadFree("public")
	if read.Status != StatusOK || !read.HasValue || !bytes.Equal(read.Value, []byte("visible")) {
		t.Fatalf("free read failed: %+v", read)
	}
}

func TestDiferirAndReadCell(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()

	parentSecret := []byte("parent-secret")
	parentGenome := ouroboros.Diferir | ouroboros.LeerSelf
	parentIndex := appendTestCell(t, store, parentSecret, parentGenome)

	var childSalt [16]byte
	copy(childSalt[:], []byte("child-salt-00001"))

	result := store.Diferir(parentIndex, parentSecret, []byte("child-secret"), childSalt, ouroboros.LeerSelf, 7, 8, 9)
	if result.Status != StatusOK || !result.HasDeferred || !result.HasNewCell {
		t.Fatalf("diferir failed: %+v", result)
	}

	child := store.ReadCell(result.DeferredIndex, []byte("child-secret"))
	if child.Status != StatusOK || !child.HasCell || child.Cell.X != 7 || child.Cell.Y != 8 || child.Cell.Z != 9 {
		t.Fatalf("child read failed: %+v", child)
	}
}

func TestMembranesPersistAcrossReopen(t *testing.T) {
	path := t.TempDir() + "/samsara.db"

	store, err := New(path, 64)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	secret := []byte("persist-secret")
	index := appendTestCell(t, store, secret, ouroboros.LeerSelf|ouroboros.EscribirSelf)

	write := store.Write("persisted", []byte("value"), index, secret)
	if write.Status != StatusOK || !write.HasNewCell {
		t.Fatalf("write failed: %+v", write)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	reopened, err := New(path, 64)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer reopened.Close()

	read := reopened.Read("persisted", write.NewCellIndex, secret)
	if read.Status != StatusOK || !read.HasValue || !bytes.Equal(read.Value, []byte("value")) {
		t.Fatalf("reopened read failed: %+v", read)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := New(t.TempDir()+"/samsara.db", 64)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	return store
}

func appendTestCell(t *testing.T, store *Store, secret []byte, genome uint32) uint32 {
	t.Helper()

	var salt [16]byte
	copy(salt[:], []byte("test-salt-000000"))

	index, err := store.DB().Append(NewCellWithSecret(salt, secret, genome, 1, 2, 3))
	if err != nil {
		t.Fatalf("append cell: %v", err)
	}

	return index
}
