# Quick Start - Guía Rápida de Opcodes

## Referencia Rápida de Opcodes

```
LECTURA BÁSICA:
├─ READ (0x01)       → Leer datos privados [Requiere Auth]
└─ READ_FREE (0x02)  → Leer datos públicos [Sin Auth]

ESCRITURA:
└─ WRITE (0x03)      → Escribir/actualizar datos [Requiere Auth]

ELIMINACIÓN:
└─ DELETE (0x04)     → Eliminar datos [Requiere Auth]

METADATA:
└─ READ_CELL (0x05)  → Leer info de célula [Requiere Auth]

REPRODUCCIÓN:
├─ DIFERIR (0x06)    → Crear hijo de 1 padre [Requiere Auth]
└─ CRUZAR (0x07)     → Crear hijo de 2 padres [Requiere Auth]
```

## Primeros Pasos (5 minutos)

### 1. Crear una Tienda (Store)

```go
package main

import "github.com/usuario/samsara-go"

func main() {
    // Crear tienda en ./data con máximo 1000 registros
    store, err := samsara.New("./data", 1000)
    defer store.Close()
}
```

### 2. Crear una Célula Base

```go
import "crypto/rand"

// Generar salt aleatorio
salt := [16]byte{}
rand.Read(salt[:])

// Definir credentials
secret := []byte("mi_contraseña")
genome := uint32(0xFF) // Todos los permisos

// Crear célula
cell := samsara.NewCellWithSecret(salt, secret, genome, 0, 0, 0)
cellIndex, _ := store.DB().Append(cell)

fmt.Printf("Célula creada: %d\n", cellIndex)
```

### 3. Escribir un Dato

```go
result := store.Write("usuario:nombre", []byte("Juan"), cellIndex, secret)

if result.Status == samsara.StatusOK {
    fmt.Println("✓ Dato escrito")
    cellIndex = result.NewCellIndex // Actualizar index
}
```

### 4. Leer el Dato

```go
result := store.Read("usuario:nombre", cellIndex, secret)

if result.Status == samsara.StatusOK {
    fmt.Printf("Dato: %s\n", string(result.Value))
    cellIndex = result.NewCellIndex
}
```

### 5. Crear Dato Público

```go
// Crear célula pública con flag LeerLibre
publicCell := samsara.NewCellWithSecret(salt, []byte("pub"), 0x04, 0, 0, 0)
pubIndex, _ := store.DB().Append(publicCell)

// Escribir dato
store.Write("anuncio:bienvenida", []byte("¡Hola!"), pubIndex, []byte("pub"))

// Leer públicamente (sin auth)
result := store.ReadFree("anuncio:bienvenida")
fmt.Printf("Anuncio: %s\n", result.Value)
```

## Patrones Comunes

### Patrón: Actualizar Dato

```go
// Leer
readResult := store.Read(key, cellIndex, secret)
if readResult.Status != samsara.StatusOK {
    return
}

// Modificar (deserializar, actualizar, serializar)
newValue := updateValue(readResult.Value)

// Escribir
writeResult := store.Write(key, newValue, readResult.NewCellIndex, secret)
if writeResult.Status == samsara.StatusOK {
    cellIndex = writeResult.NewCellIndex
}
```

### Patrón: Sistema de Roles

```go
// Crear Admin
adminCell := samsara.NewCellWithSecret(salt, []byte("admin_pwd"), 0xFF, 0, 0, 0)
adminIndex, _ := store.DB().Append(adminCell)

// Crear Usuario desde Admin (DIFERIR)
result := store.Diferir(
    adminIndex,
    []byte("admin_pwd"),
    []byte("user_pwd"),
    childSalt,
    0x09, // Solo LeerSelf | EscribirSelf
    0, 0, 0,
)

if result.Status == samsara.StatusOK {
    userIndex := result.DeferredIndex
    // Usuario puede leer y escribir sus propios datos
}
```

### Patrón: Crear Cuenta Superadmin

```go
// Fusionar dos admins para crear superadmin
result := store.Cruzar(
    admin1Index, admin1Secret,
    admin2Index, admin2Secret,
    []byte("superadmin_pwd"),
    childSalt,
    0, 0, 0,
)

// El superadmin tendrá: admin1 permisos | admin2 permisos
```

## Tabla de Permisos (Flags de Genoma)

```
Bit  Valor  Nombre        Descripción
─────────────────────────────────────────────
0    0x01   LeerSelf      Leer datos propios
1    0x02   LeerAny       Leer datos de otros
2    0x04   LeerLibre     Hacer datos públicos
3    0x08   EscribirSelf  Escribir datos propios
4    0x10   EscribirAny   Escribir datos de otros
5    0x20   BorrarSelf    Borrar datos propios
6    0x40   BorrarAny     Borrar datos de otros
7    0x80   Diferir       Crear células hijo
8    0x100  Fucionar      Fusionar con otra célula
```

### Combinaciones Comunes

```go
// Solo lectura (usuario guest)
const Guest = 0x01 // LeerSelf

// Lectura y escritura (usuario normal)
const User = 0x09 // LeerSelf | EscribirSelf

// Administrador
const Admin = 0xFE // Todo excepto Fucionar

// Superadmin (con reproducción)
const SuperAdmin = 0xFF // Todo

// Lectura pública
const PublicRead = 0x04 // LeerLibre
```

## Flujo Típico: CRUD

### Create (Crear)

```go
value := []byte("nuevo dato")
result := store.Write("key", value, cellIndex, secret)
if result.Status == samsara.StatusOK {
    cellIndex = result.NewCellIndex
    // ✓ Creado
}
```

### Read (Leer)

```go
result := store.Read("key", cellIndex, secret)
if result.Status == samsara.StatusOK {
    data := result.Value
    cellIndex = result.NewCellIndex
    // ✓ Leído
}
```

### Update (Actualizar)

```go
// Read → Modify → Write (ver patrón anterior)
```

### Delete (Borrar)

```go
result := store.Delete("key", cellIndex, secret)
if result.Status == samsara.StatusOK {
    cellIndex = result.NewCellIndex
    // ✓ Borrado
}
```

## Resolución de Problemas

| Problema | Causa | Solución |
|----------|-------|----------|
| `unauthorized` | Secret inválido | Verificar secret y cell index |
| `undefined` | Dato no existe | Verificar clave, crear primero |
| `error_db` | Error BD | Revisar permisos, reintentar |
| NewCell no cambia | Cell no está refrescando | Siempre usar NewCellIndex retornado |

## Estados de Respuesta

```go
const (
    StatusOK           = "ok"           // ✓ Exitoso
    StatusUnauthorized = "unauthorized" // ✗ Sin permisos
    StatusUndefined    = "undefined"    // ✗ No existe
    StatusErrorDB      = "error_db"     // ✗ Error BD
)
```

## Estructura Básica de Célula

```go
type Celula struct {
    Hash   [32]byte   // BLAKE3(salt + secret)
    Salt   [16]byte   // Aleatorio
    Genoma  uint32    // Permisos (flags)
    X, Y, Z uint32    // Coordenadas (opcional)
}
```

## Workflow: Administrar Usuarios

```
1. Admin crea usuario
   └─ store.Diferir(adminIndex, ...) → userIndex

2. Usuario autentica
   └─ store.ReadCell(userIndex, userSecret)

3. Usuario escribe datos
   └─ store.Write(key, value, userIndex, userSecret)

4. Admin supervisa
   └─ store.Read(key, adminIndex, adminSecret)

5. Revocar acceso
   └─ store.Delete(key, adminIndex, adminSecret)
```

## Archivos de Referencia Completos

- **INDEX.md** - Tabla de todos los opcodes
- **README.md** - Guía conceptual
- **READ.md** - Detalles opcode READ
- **READ_FREE.md** - Detalles opcode READ_FREE
- **WRITE.md** - Detalles opcode WRITE
- **DELETE.md** - Detalles opcode DELETE
- **READ_CELL.md** - Detalles opcode READ_CELL
- **DIFERIR.md** - Detalles opcode DIFERIR
- **CRUZAR.md** - Detalles opcode CRUZAR
- **PERMISOS.md** - Tabla completa de permisos
- **QUICK_START.md** - Este archivo

## Próximos Pasos

1. Lee [README.md](README.md) para entender conceptos
2. Consulta [INDEX.md](INDEX.md) para overview de opcodes
3. Implementa operaciones CRUD básicas (READ, WRITE, DELETE)
4. Experimenta con READ_CELL para entender permisos
5. Crea usuarios derivados con DIFERIR
6. Fusiona roles con CRUZAR

## Tips y Trucos

### Siempre Guardar Nuevo Index

```go
result := store.Write(key, value, cellIndex, secret)
cellIndex = result.NewCellIndex // ✓ Importante
```

### Validar Estado Antes de Actuar

```go
if result.Status == samsara.StatusOK {
    // Proceder
} else if result.Status == samsara.StatusUnauthorized {
    // Permisos insuficientes
} else if result.Status == samsara.StatusUndefined {
    // No existe
}
```

### Usa Genoma 0xFF para Testing

```go
// Célula de prueba con TODOS los permisos
testGenome := uint32(0xFF)
cell := samsara.NewCellWithSecret(salt, secret, testGenome, 0, 0, 0)
```

### Leer Antes de Modificar

```go
// Siempre verifica si datos existen
result := store.Read(key, cellIndex, secret)
if result.Status == samsara.StatusUndefined {
    // Crear nuevo
}
// Proceder con modificación
```

## Ejemplo Completo: Chat Simple

```go
func createChat(store *samsara.Store) error {
    // 1. Crear célula para el chat
    salt := [16]byte{}
    rand.Read(salt[:])
    
    chatCell := samsara.NewCellWithSecret(salt, []byte("chat_pwd"), 0xFF, 0, 0, 0)
    chatIndex, _ := store.DB().Append(chatCell)
    
    // 2. Escribir primer mensaje
    msg := []byte("Mensaje inicial")
    store.Write("chat:messages:1", msg, chatIndex, []byte("chat_pwd"))
    
    // 3. Publicar para todos (READ_FREE)
    store.Write("chat:public", msg, chatIndex, []byte("chat_pwd"))
    
    // 4. Cualquiera puede leer (sin auth)
    result := store.ReadFree("chat:public")
    fmt.Printf("Mensaje público: %s\n", result.Value)
    
    return nil
}
```

## Documentación Completa

Para detalles de implementación, casos de error avanzados, y ejemplos complejos, consulta la carpeta `opcodes/` con la documentación individual de cada opcode.

