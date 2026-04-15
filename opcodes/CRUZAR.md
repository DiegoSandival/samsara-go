# Opcode: CRUZAR (0x07)

## Descripción

Fusiona dos células padre para crear una célula hijo con permisos combinados. Es una operación de "reproducción sexual" donde el hijo hereda la unión de los permisos de ambos padres.

## Información del Opcode

| Propiedad | Valor |
|-----------|-------|
| Código | `0x07` |
| Requiere Autenticación | **Sí** (Ambos Padres) |
| Requeridos Cell Indices | **Sí** (Index de A y B) |
| Retorna Nuevas Cells | **Sí** (A + B + Hijo) |
| Tipo | Operación Genética (Fusión) |

## Parámetros

| Parámetro | Tipo | Tamaño | Descripción |
|-----------|------|--------|------------|
| `cellIndexA` | uint32 | 4 bytes | Índice de la primera célula padre |
| `secretA` | []byte | Variable | Secret del padre A para autenticación |
| `cellIndexB` | uint32 | 4 bytes | Índice de la segunda célula padre |
| `secretB` | []byte | Variable | Secret del padre B para autenticación |
| `childSecret` | []byte | Variable | Secret del hijo a crear |
| `childSalt` | [16]byte | 16 bytes | Salt aleatorio para el hijo |
| `x`, `y`, `z` | uint32 | 12 bytes | Coordenadas del hijo (3x4 bytes) |

## Estructura Binaria

```
[Opcode: 0x07]
[CellIndexA: 4 bytes]
[SecretA Length: 4 bytes]
[SecretA: N bytes]
[CellIndexB: 4 bytes]
[SecretB Length: 4 bytes]
[SecretB: M bytes]
[ChildSecret Length: 4 bytes]
[ChildSecret: P bytes]
[ChildSalt: 16 bytes]
[X: 4 bytes]
[Y: 4 bytes]
[Z: 4 bytes]
```

## Resultado

```go
type CruzarResult struct {
    Status        Status  // ok | unauthorized | undefined | error_db
    CellIndexA    uint32  // Índice de padre A actual
    CellIndexB    uint32  // Índice de padre B actual
    ChildIndex    uint32  // Índice de la célula hijo creada
    NewCellIndexA uint32  // Nuevo índice del padre A después de refresh
    NewCellIndexB uint32  // Nuevo índice del padre B después de refresh
    HasCellIndexA bool    // La respuesta contiene CellIndexA
    HasCellIndexB bool    // La respuesta contiene CellIndexB
    HasChild      bool    // La respuesta contiene ChildIndex
    HasNewCellA   bool    // La respuesta contiene NewCellIndexA
    HasNewCellB   bool    // La respuesta contiene NewCellIndexB
}
```

## Estados de Respuesta

| Status | Causa | ChildIndex | NewCellIndexA | NewCellIndexB |
|--------|-------|---|---|---|
| `ok` | Fusión exitosa | ✓ Retornado | ✓ Retornado | ✓ Retornado |
| `unauthorized` | Secrets incorrectos o padres sin flag Fucionar | - | ✓ Diagnostico | ✓ Diagnostico |
| `undefined` | No aplicable | - | - | - |
| `error_db` | Error en base de datos | - | - | - |

## Restricciones de Permisos

### Requisitos de Ambos Padres

1. **Autenticación de A**: Debe autenticarse correctamente con secretA
2. **Flag Fucionar en A**: Debe tener flag `Fucionar` (0x100)
3. **Autenticación de B**: Debe autenticarse correctamente con secretB
4. **Flag Fucionar en B**: Debe tener flag `Fucionar` (0x100)

### Herencia de Genoma del Hijo

El genoma del hijo es la **unión (OR)** de los genomas de ambos padres:

```go
childGenoma = genomeA | genomeB
```

**Ejemplo**:
```
Padre A: 0b00001111 (0x0F) - Lectura y escritura
Padre B: 0b11100000 (0xE0) - Borrado y reproducción
─────────────────────────
Hijo:    0b11101111 (0xEF) - Todos los permisos de ambos
```

## Proceso de Operación

1. **Resolver Padre A** - Autenticar con secretA
2. **Resolver Padre B** - Autenticar con secretB
3. **Validar Flags Fucionar** - Ambos padres deben tener flag
4. **Calcular Genoma Hijo** - genomeA | genomeB
5. **Crear Hijo** - Generar célula hijo con genoma combinado
6. **Almacenar Hijo** - Agregar a BD Ouroboros
7. **Refrescar Padres** - Generar nuevos índices para ambos
8. **Retornar Índices** - Return padres renovados + hijo

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

    // Padre A: Solo lectura
    saltA := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
    secretA := []byte("padre_a_secreto")
    genomeA := uint32(0x07 | 0x100) // Lectura + Fucionar
    cellA := samsara.NewCellWithSecret(saltA, secretA, genomeA, 0, 0, 0)
    indexA, _ := store.DB().Append(cellA)
    
    // Padre B: Solo escritura
    saltB := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
    secretB := []byte("padre_b_secreto")
    genomeB := uint32(0x18 | 0x100) // Escritura + Fucionar
    cellB := samsara.NewCellWithSecret(saltB, secretB, genomeB, 0, 0, 0)
    indexB, _ := store.DB().Append(cellB)
    
    // Crear hijo mediante CRUZAR
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    
    result := store.Cruzar(
        indexA, secretA,
        indexB, secretB,
        []byte("hijo_secreto"),
        childSalt,
        100, 200, 300, // Coordenadas
    )
    
    if result.Status == samsara.StatusOK {
        fmt.Printf("Hijo creado exitosamente\n")
        fmt.Printf("Índice del hijo: %d\n", result.ChildIndex)
        fmt.Printf("Nuevo índice padre A: %d\n", result.NewCellIndexA)
        fmt.Printf("Nuevo índice padre B: %d\n", result.NewCellIndexB)
        
        // El hijo hereda: lectura + escritura + Fucionar
        // Genoma hijo = 0x07 | 0x18 | 0x100 = 0x11F
    }
}
```

## Ejemplo de Uso - Protocolo Binario

```python
import struct
import os

def cruzar_opcode(store, index_a, secret_a, index_b, secret_b, 
                  child_secret, child_salt, x, y, z):
    # Datos
    opcode = 0x07
    secret_a_bytes = secret_a.encode('utf-8')
    secret_b_bytes = secret_b.encode('utf-8')
    child_secret_bytes = child_secret.encode('utf-8')
    
    # Construir mensaje
    message = bytearray()
    message.append(opcode)
    message.extend(struct.pack('<I', index_a))
    message.extend(struct.pack('<I', len(secret_a_bytes)))
    message.extend(secret_a_bytes)
    message.extend(struct.pack('<I', index_b))
    message.extend(struct.pack('<I', len(secret_b_bytes)))
    message.extend(secret_b_bytes)
    message.extend(struct.pack('<I', len(child_secret_bytes)))
    message.extend(child_secret_bytes)
    message.extend(child_salt)  # 16 bytes
    message.extend(struct.pack('<I', x))
    message.extend(struct.pack('<I', y))
    message.extend(struct.pack('<I', z))
    
    # Enviar y recibir
    response = store.send(bytes(message))
    
    # Parsear respuesta
    status = response[0]
    if status == 0:  # ok
        index_a_curr = struct.unpack('<I', response[1:5])[0]
        index_b_curr = struct.unpack('<I', response[5:9])[0]
        child_index = struct.unpack('<I', response[9:13])[0]
        new_index_a = struct.unpack('<I', response[13:17])[0]
        new_index_b = struct.unpack('<I', response[17:21])[0]
        
        return {
            'status': 'ok',
            'child_index': child_index,
            'new_index_a': new_index_a,
            'new_index_b': new_index_b,
        }
    
    return {'status': 'error'}
```

## Herencia de Genoma en Fusión

```
Padre A: 0b00000011 (Lectura)      = 0x03
Padre B: 0b00001100 (Escritura)    = 0x0C
─────────────────────────────────────────
Hijo:    0b00001111 (L + E)        = 0x0F

Padre A: 0b01100000 (Borrado)      = 0x60
Padre B: 0b10000000 (Reproducción) = 0x80
─────────────────────────────────────────
Hijo:    0b11100000 (B + R)        = 0xE0

Padre A: 0b00000001 (Lectura Self)      = 0x01
Padre B: 0b11111110 (Todo)              = 0xFE
─────────────────────────────────────────
Hijo:    0b11111111 (Todos)             = 0xFF
```

## Casos de Uso Comunes

### 1. Crear Cuenta Superadmin desde Dos Administradores

```go
func createSuperAdmin(store *samsara.Store,
    adminAIndex uint32, adminASecret []byte,
    adminBIndex uint32, adminBSecret []byte) (uint32, error) {
    
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    
    result := store.Cruzar(
        adminAIndex, adminASecret,
        adminBIndex, adminBSecret,
        []byte("superadmin_secret"),
        childSalt,
        0, 0, 2, // Z=2 indica nivel super admin
    )
    
    if result.Status != samsara.StatusOK {
        return 0, fmt.Errorf("failed to create super admin")
    }
    
    log.Printf("SuperAdmin creado: %d", result.ChildIndex)
    return result.ChildIndex, nil
}
```

### 2. Combinar Roles (Role A + Role B = Role C)

```go
func mergeRoles(store *samsara.Store,
    roleAIndex uint32, roleASecret []byte,
    roleBIndex uint32, roleBSecret []byte,
    newRoleName string) (uint32, error) {
    
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    
    result := store.Cruzar(
        roleAIndex, roleASecret,
        roleBIndex, roleBSecret,
        []byte(newRoleName),
        childSalt,
        0, 0, 0,
    )
    
    if result.Status != samsara.StatusOK {
        return 0, fmt.Errorf("role merge failed")
    }
    
    fmt.Printf("Role %s combina permisos: 0x%02X | 0x%02X\n",
        newRoleName, getGenoma(roleAIndex), getGenoma(roleBIndex))
    
    return result.ChildIndex, nil
}
```

### 3. Crear Cuenta con Capacidades de Ambos Padres

```go
func createHybridAccount(store *samsara.Store,
    devIndex uint32, devSecret []byte,
    businessIndex uint32, businessSecret []byte) (uint32, error) {
    
    // Developer tiene: READ, WRITE (datos propios)
    // Business tiene: READ, WRITE (datos de otros), DELETE
    // Hybrid tendrá: todos esos permisos combinados
    
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    
    result := store.Cruzar(
        devIndex, devSecret,
        businessIndex, businessSecret,
        []byte("hybrid_account"),
        childSalt,
        0, 0, 0,
    )
    
    if result.Status != samsara.StatusOK {
        return 0, fmt.Errorf("hybrid account creation failed")
    }
    
    return result.ChildIndex, nil
}
```

### 4. Fusionar Datos de Dos Fuentes

```go
// Crear células "portadoras" de permiso para acceso combinado
func allowDualAccess(store *samsara.Store,
    sourceAIndex uint32, sourceASecret []byte,
    sourceBIndex uint32, sourceBSecret []byte) (uint32, error) {
    
    result := store.Cruzar(
        sourceAIndex, sourceASecret,
        sourceBIndex, sourceBSecret,
        []byte("dual_source"),
        [16]byte{},
        0, 0, 0,
    )
    
    if result.Status != samsara.StatusOK {
        return 0, fmt.Errorf("dual access creation failed")
    }
    
    return result.ChildIndex, nil
}
```

## Ciclo de Vida de Fusión

```
Padre A (Genoma A)          Padre B (Genoma B)
        │                           │
        ├───────────────┬───────────┤
        │               │           │
        ▼               ▼           ▼
  Validación A    Validación B  Validación de Flags
     (Auth)          (Auth)      (Ambos Fucionar)
        │               │           │
        └───────────────┴───────────┘
                │
        ▼ (Ambas válidas)
    Crear Hijo
    Genoma = Genoma A | Genoma B
                │
        ▼
    Almacenar Hijo
                │
        ▼
    Refrescar Padre A → New Index A
    Refrescar Padre B → New Index B
                │
        ▼
    Retornar (Hijo, New A, New B)
```

## Casos de Error Comunes

### Error: `unauthorized` - Padre sin flag Fucionar

```
Causa: Padre A o Padre B no tiene flag Fucionar (0x100)
Síntoma: result.Status == StatusUnauthorized

Diagnóstico:
- Ambos padres retornan en HasCellIndexA/B
- Permite identificar cuál padre tiene problema

Solución:
- Crear nuevo padre con Fucionar habilitado
- O usar DIFERIR si solo necesitas un padre
```

### Error: `unauthorized` - Secret incorrecto

```
Causa: secretA o secretB no coinciden con hash de célula
Síntoma: result.Status == StatusUnauthorized

Solución:
- Verificar que los secrets son correctos
- Usar mismo secret que fue usado para crear células
```

### Error: `error_db` durante creación

```
Causa: Problema almacenando hijo o refrescando padres
Síntoma: result.Status == StatusErrorDB

Solución:
- Reintentar operación
- Verificar integridad de base de datos
- Liberar espacio si BD está llena
```

## Validación Previa

```go
func canCruzar(store *samsara.Store, 
    indexA uint32, secretA []byte,
    indexB uint32, secretB []byte) bool {
    
    // Validar célula A
    resultA := store.ReadCell(indexA, secretA)
    if resultA.Status != samsara.StatusOK {
        return false
    }
    if resultA.Cell.Genoma&ouroboros.Fucionar == 0 {
        return false
    }
    
    // Validar célula B
    resultB := store.ReadCell(indexB, secretB)
    if resultB.Status != samsara.StatusOK {
        return false
    }
    if resultB.Cell.Genoma&ouroboros.Fucionar == 0 {
        return false
    }
    
    return true
}

func getChildGenome(store *samsara.Store,
    indexA uint32, secretA []byte,
    indexB uint32, secretB []byte) uint32 {
    
    resultA := store.ReadCell(indexA, secretA)
    resultB := store.ReadCell(indexB, secretB)
    
    return resultA.Cell.Genoma | resultB.Cell.Genoma
}
```

## Integración en Sistema de Roles

```go
type RoleFusionRequest struct {
    RoleA string `json:"role_a"`
    RoleB string `json:"role_b"`
    Name  string `json:"name"`
}

func (h *Handler) FuseRoles(w http.ResponseWriter, r *http.Request) {
    var req RoleFusionRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // Buscar índices de ambos roles
    indexA := h.roles[req.RoleA]
    indexB := h.roles[req.RoleB]
    secretA := h.secrets[req.RoleA]
    secretB := h.secrets[req.RoleB]
    
    childSalt := [16]byte{}
    rand.Read(childSalt[:])
    
    result := h.store.Cruzar(
        indexA, secretA,
        indexB, secretB,
        []byte(req.Name),
        childSalt,
        0, 0, 0,
    )
    
    if result.Status != samsara.StatusOK {
        http.Error(w, "Fusion failed", http.StatusBadRequest)
        return
    }
    
    // Guardar nuevo rol
    h.roles[req.Name] = result.ChildIndex
    h.secrets[req.Name] = []byte(req.Name)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

## Monitoreo y Auditoría

```go
func logCruzar(indexA, indexB, childIndex uint32, genomeA, genomeB uint32) {
    childGenome := genomeA | genomeB
    log.Printf("CRUZAR: A=%d B=%d child=%d genomeA=0x%02X genomeB=0x%02X childGenome=0x%02X timestamp=%s",
        indexA, indexB, childIndex, genomeA, genomeB, childGenome, time.Now().Format(time.RFC3339))
}
```

## Comparación: DIFERIR vs CRUZAR

| Aspecto | DIFERIR | CRUZAR |
|---------|---------|--------|
| Padres | 1 | 2 |
| Autenticación | 1 secret | 2 secrets |
| Flag Requerido | Diferir | Fucionar (ambos) |
| Genoma Hijo | Subconjunto padre | Unión padres (OR) |
| Caso de Uso | Asexual | Sexual |
| Complejidad | Baja | Alta |
| Permisos | Restrictos | Expansivos |

