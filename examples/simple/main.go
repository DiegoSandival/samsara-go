package main

import (
	"fmt"
	"log"
	"os"

	ouroboros "github.com/DiegoSandival/ouroboros-go"
	samsara "github.com/DiegoSandival/samsara-go"
)

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