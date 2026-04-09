# samsara-go

Capa minima sobre `github.com/DiegoSandival/ouroboros-go` que implementa el pseudocodigo de `psudocodigo.txt`.

Expone una estructura `Store` con estas operaciones:

- `Read`
- `ReadFree`
- `Write`
- `Delete`
- `Diferir`
- `Cruzar`
- `ReadCell`

Las celulas viven en `ouroboros-go` y las membranas se guardan en bbolt. Si abres el store con `New("./samsara.db", ...)`, las membranas se persisten en `./samsara.db.bolt`.

Ejemplo minimo:

```go
store, err := samsara.New("./samsara.db", 128)
if err != nil {
	panic(err)
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
	panic(err)
}

write := store.Write("foo", []byte("bar"), index, secret)
read := store.Read("foo", write.NewCellIndex, secret)
```
