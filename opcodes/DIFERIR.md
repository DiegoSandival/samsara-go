# Opcode: DIFERIR (0x06)

## Descripción

Crea una célula hijo derivada de una célula padre. Es una operación de "reproducción" que permite a una célula generar descendientes con permisos controlados. El genoma del hijo hereda solo los permisos que el padre permita.

## Información del Opcode

| Propiedad | Valor |
|-----------|-------|
| Código | `0x06` |
| Requiere Autenticación | **Sí** (Padre y Hijo) |
| Requiere Cell Index | **Sí** (Index del Padre) |
| Retorna Nuevas Cells | **Sí** (Padre + Hijo) |
| Tipo | Operación Genética |

## Parámetros

| Parámetro | Tipo | Tamaño | Descripción |
|-----------|------|--------|------------|
| `cellIndex` | uint32 | 4 bytes | Índice de la célula padre |
| `parentSecret` | []byte | Variable | Secret del padre para autenticación |
| `childSecret` | []byte | Variable | Secret del hijo a crear |
| `childSalt` | [16]byte | 16 bytes | Salt aleatorio para el hijo |
| `childGenome` | uint32 | 4 bytes | Permisos del hijo (subconjunto del padre) |
| `x`, `y`, `z` | uint32 | 12 bytes | Coordenadas del hijo (3x4 bytes) |

## Estructura Binaria

```
[Opcode: 0x06]
[CellIndex: 4 bytes]
[ParentSecret Length: 4 bytes]
[ParentSecret: N bytes]
[ChildSecret Length: 4 bytes]
[ChildSecret: M bytes]
[ChildSalt: 16 bytes]
[ChildGenome: 4 bytes]
[X: 4 bytes]
[Y: 4 bytes]
[Z: 4 bytes]
```

## Resultado

```go
type DiferirResult struct {
    Status        Status  // ok | unauthorized | undefined | error_db
    CellIndex     uint32  // Índice de célula padre actual
    DeferredIndex uint32  // Índice de la célula hijo creada
    NewCellIndex  uint32  // Nuevo índice del padre después de refresh
    HasCellIndex  bool    // La respuesta contiene CellIndex
    HasDeferred   bool    // La respuesta contiene DeferredIndex
    HasNewCell    bool    // La respuesta contiene NewCellIndex
}
```

## Estados de Respuesta

| Status | Causa | DeferredIndex | NewCellIndex |
|--------|-------|---|---|
| `ok` | Hijo creado exitosamente | ✓ Retornado | ✓ Retornado |
| `unauthorized` | Secret incorrecto o hijo excede permisos del padre | - | - |
| `undefined` | Padre no existe | - | - |
| `error_db` | Error accediendo base de datos | - | - |

## Restricciones de Permisos

### Requisito del Padre

1. **Autenticación**: El padre debe autenticarse correctamente con parentSecret
2. **Flag Diferir**: El padre debe tener el flag `Diferir` (0x80) habilitado
3. **Sublotamiento de Permisos**: El genoma del hijo no puede exceder el del padre

### Validación de Genoma del Hijo

```go
// La célula padre limita los permisos del hijo
if (childGenome & parentGenoma) != childGenome {
    // Error: hijo intenta tener permisos que padre no tiene
    return DiferirResult{Status: StatusUnauthorized}
}
```

## Proceso de Operación

1. **Resolver Padre** - Autenticar padre con parentSecret
2. **Validar Flag Diferir** - Verificar padre tiene permiso de reproducción
3. **Validar Herencia de Genoma** - Hijo ⊆ Padre (genéticamente)
4. **Crear Hijo** - Generar célula hijo con:
   - Hash = BLAKE3(childSalt + childSecret)
   - Genoma = childGenoma
   - Coordenadas = (x, y, z)
5. **Almacenar Hijo** - Agregar a la BD Ouroboros
6. **Refrescar Padre** - Generar nuevo index para padre
7. **Retornar Índices** - Return nuevo padre y hijo

## Ejemplo de Uso - Go

```go
package main

import (
    "fmt"
    "crypto/rand"
    "github.com/usuario/samsara-go"
)

func main() {
    store, _ := samsara.New("./data", 1000)
    defer store.Close()

    // Datos del padre
    parentSalt := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
    parentSecret := []byte("padre_secreto")
    parentGenome := uint32(0xFF) // Todos los permisos
    
    // Crear padre
    parent := samsara.NewCellWithSecret(parentSalt, parentSecret, parentGenome, 0, 0, 0)
    parentIndex, _ := store.DB().Append(parent)
    
    // Datos del hijo
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    childSecret := []byte("hijo_secreto")
    childGenome := uint32(0x7F) // Todos excepto Diferir y Fucionar
    
    // Crear hijo mediante DIFERIR
    result := store.Diferir(
        parentIndex,
        parentSecret,
        childSecret,
        childSalt,
        childGenome,
        100, 200, 300, // Coordenadas del hijo
    )
    
    if result.Status == samsara.StatusOK {
        fmt.Printf("Hijo creado exitosamente\n")
        fmt.Printf("Índice del hijo: %d\n", result.DeferredIndex)
        fmt.Printf("Nuevo índice del padre: %d\n", result.NewCellIndex)
    } else if result.Status == samsara.StatusUnauthorized {
        fmt.Println("Error: El padre no tiene permiso Diferir o hijo excede permisos")
    } else if result.Status == samsara.StatusErrorDB {
        fmt.Println("Error: Problema con la base de datos")
    }
}
```

## Ejemplo de Uso - Protocolo Binario

```python
import struct
import os

def diferir_opcode(store, parent_index, parent_secret, child_secret, child_salt, child_genome, x, y, z):
    # Datos
    opcode = 0x06
    parent_secret_bytes = parent_secret.encode('utf-8')
    child_secret_bytes = child_secret.encode('utf-8')
    
    # Construir mensaje
    message = bytearray()
    message.append(opcode)
    message.extend(struct.pack('<I', parent_index))
    message.extend(struct.pack('<I', len(parent_secret_bytes)))
    message.extend(parent_secret_bytes)
    message.extend(struct.pack('<I', len(child_secret_bytes)))
    message.extend(child_secret_bytes)
    message.extend(child_salt)  # 16 bytes
    message.extend(struct.pack('<I', child_genome))
    message.extend(struct.pack('<I', x))
    message.extend(struct.pack('<I', y))
    message.extend(struct.pack('<I', z))
    
    # Enviar y recibir
    response = store.send(bytes(message))
    
    # Parsear respuesta
    status = response[0]
    if status == 0:  # ok
        parent_index = struct.unpack('<I', response[1:5])[0]
        child_index = struct.unpack('<I', response[5:9])[0]
        new_parent_index = struct.unpack('<I', response[9:13])[0]
        
        return {
            'status': 'ok',
            'child_index': child_index,
            'new_parent_index': new_parent_index,
        }
    
    return {'status': 'error'}
```

## Jerarquía de Generaciones

```
Generación 0 (Raíz)
    └─ Padre (Genoma: 0xFF)
        ├─ Hijo 1 (Genoma: 0x7F - sin reproducción propia)
        ├─ Hijo 2 (Genoma: 0x3F - sin reproducción ni fusión)
        └─ Hijo 3 (Genoma: 0x0F - solo lectura/escritura)
            └─ Nieto (del Hijo 1, si tiene flag Diferir)
```

## Casos de Uso Comunes

### 1. Crear Cuenta Secundaria con Permisos Limitados

```go
func createRestrictedAccount(store *samsara.Store, 
    parentIndex uint32, parentSecret []byte) error {
    
    // Hijo solo puede leer y escribir datos propios
    childGenome := samsara.LeerSelf | samsara.EscribirSelf // 0x09
    
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    
    result := store.Diferir(
        parentIndex,
        parentSecret,
        []byte("cuenta_secundaria"),
        childSalt,
        childGenome,
        0, 0, 0,
    )
    
    if result.Status != samsara.StatusOK {
        return fmt.Errorf("failed to create account")
    }
    
    log.Printf("Cuenta secundaria creada: %d", result.DeferredIndex)
    return nil
}
```

### 2. Crear Cuenta con Permisos Administrativos

```go
func createAdminAccount(store *samsara.Store,
    parentIndex uint32, parentSecret []byte) error {
    
    // Admin puede leer, escribir y administrar datos
    childGenome := ^uint32(samsara.Diferir | samsara.Fucionar) // Todo menos reproducción
    
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    
    result := store.Diferir(
        parentIndex,
        parentSecret,
        []byte("admin_cuenta"),
        childSalt,
        childGenome,
        0, 0, 1, // Z=1 indica nivel admin
    )
    
    if result.Status != samsara.StatusOK {
        return fmt.Errorf("failed to create admin")
    }
    
    return nil
}
```

### 3. Crear Célula Especializada para Fusión

```go
func createFusibleCell(store *samsara.Store,
    parentIndex uint32, parentSecret []byte) error {
    
    // Célula solo para fusión con otra célula
    childGenome := samsara.Fucionar // 0x100
    
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    
    result := store.Diferir(
        parentIndex,
        parentSecret,
        []byte("fusionable"),
        childSalt,
        childGenome,
        0, 0, 0,
    )
    
    if result.Status != samsara.StatusOK {
        return fmt.Errorf("failed to create fusible cell")
    }
    
    return nil
}
```

## Herencia de Permisos

```
Padre: 0xFF (11111111) - Todos los permisos
│
├─ Hijo A: 0x7F (01111111) - Sin Fucionar
│  └─ Nieto A1: 0x3F (00111111) - Sin Fucionar, Sin Diferir
│
├─ Hijo B: 0x0F (00001111) - Solo CRUD básico
│  └─ No puede crear nietos (sin Diferir)
│
└─ Hijo C: 0x03 (00000011) - Solo lectura
   └─ No puede reproducirse
```

## Ciclo de Vida Generacional

```
1. Crear Padre (WRITE o manualmente)
2. Padre tiene flag Diferir habilitado
3. Llamar DIFERIR con:
   - Referencia al padre
   - Secret del padre
   - Secret del hijo
   - Genoma del hijo (subconjunto del padre)
4. Hijo se almacena en la BD
5. Ambos (padre e hijo) obtienen nuevo index
6. Hijo puede realizar operaciones según su genoma
7. Si hijo tiene Diferir, puede crear su propio descendiente
```

## Casos de Error Comunes

### Error: `unauthorized` - Sin flag Diferir

```
Causa: Padre no tiene permiso Diferir
Síntoma: result.Status == StatusUnauthorized

Solución:
- Verificar que parentGenome & Diferir == Diferir
- O crear nueva célula con Diferir habilitado
```

### Error: `unauthorized` - Hijo excede permisos

```
Causa: childGenome & parentGenome != childGenome
Síntoma: result.Status == StatusUnauthorized

Solución:
- Reducir permisos del hijo
- childGenome &= parentGenome (operación AND)
```

### Error: `error_db` durante creación

```
Causa: Problema almacenando hijo
Síntoma: result.Status == StatusErrorDB

Solución:
- Reintentar con backoff exponencial
- Verificar espacio en disco
- Validar integridad de base de datos
```

## Validación Previa

```go
func canDiferir(store *samsara.Store, parentIndex uint32, parentSecret []byte) bool {
    // Leer información del padre
    result := store.ReadCell(parentIndex, parentSecret)
    if result.Status != samsara.StatusOK {
        return false
    }
    
    // Verificar flag Diferir
    return result.Cell.Genoma&ouroboros.Diferir != 0
}

func validateChildGenome(parentGenoma, childGenoma uint32) error {
    if (childGenoma & parentGenoma) != childGenoma {
        return fmt.Errorf("child genome exceeds parent genome")
    }
    return nil
}
```

## Integración en API de Cuentas

```go
type CreateAccountRequest struct {
    ParentIndex    uint32   `json:"parent_index"`
    ParentSecret   string   `json:"parent_secret"`
    ChildSecret    string   `json:"child_secret"`
    PermissionLevel string   `json:"permission_level"` // "admin", "user", "guest"
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
    var req CreateAccountRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // Determinar genoma según nivel
    var childGenome uint32
    switch req.PermissionLevel {
    case "admin":
        childGenoma = 0xFE // Todo excepto Diferir
    case "user":
        childGenome = 0x3F // CRUD + administración
    case "guest":
        childGenome = 0x03 // Solo lectura
    }
    
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    
    result := h.store.Diferir(
        req.ParentIndex,
        []byte(req.ParentSecret),
        []byte(req.ChildSecret),
        childSalt,
        childGenoma,
        0, 0, 0,
    )
    
    if result.Status != samsara.StatusOK {
        http.Error(w, "Creation failed", http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "child_index": result.DeferredIndex,
    })
}
```

## Monitoreo y Auditoría

```go
func logDiferir(parentIndex, childIndex uint32, childGenome uint32) {
    log.Printf("DIFERIR: parent=%d child=%d genome=0x%02X timestamp=%s",
        parentIndex, childIndex, childGenoma, time.Now().Format(time.RFC3339))
    
    // Para cuentas privilegiadas, alertar
    if childGenome&ouroboros.BorrarAny != 0 {
        alertAdmins("Cuenta privilegiada creada: child=%d", childIndex)
    }
}
```

## Comparación con CRUZAR

| Aspecto | DIFERIR | CRUZAR |
|---------|---------|--------|
| Padres | 1 | 2 |
| Requiere Flag | Diferir | Fucionar |
| Herencia de Genoma | Subconjunto del padre | Unión de ambos padres |
| Coordinadas | Especificadas | Especificadas |
| Caso de Uso | Reproducción asexual | Reproducción sexual |

