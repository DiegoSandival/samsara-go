# Opcode: READ_CELL (0x05)

## Descripción

Lee la información completa de una célula, incluyendo su hash, salt, genoma y coordenadas. Requiere autenticación con el secret de la célula.

## Información del Opcode

| Propiedad | Valor |
|-----------|-------|
| Código | `0x05` |
| Requiere Autenticación | **Sí** |
| Requiere Cell Index | **Sí** |
| Retorna Nuevo Cell Index | **Sí** |
| Tipo | Lectura de Metadata |

## Parámetros

| Parámetro | Tipo | Tamaño | Descripción |
|-----------|------|--------|------------|
| `cellIndex` | uint32 | 4 bytes | Índice de la célula a leer |
| `secret` | []byte | Variable | Contraseña de la célula para autenticación |

## Estructura Binaria

```
[Opcode: 0x05] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: N bytes]
```

## Estructura de Célula Retornada

```go
type Celula struct {
    Hash   [32]byte   // Hash BLAKE3 (256 bits)
    Salt   [16]byte   // Salt aleatorio (128 bits)
    Genoma uint32     // Permisos/flags (32 bits)
    X      uint32     // Coordenada X
    Y      uint32     // Coordenada Y
    Z      uint32     // Coordenada Z
}
```

Tamaño total: **100 bytes**

## Resultado

```go
type CellReadResult struct {
    Status       Status        // ok | unauthorized | undefined | error_db
    Cell         Celula        // Información completa de la célula
    CellIndex    uint32        // Índice de la célula actual
    NewCellIndex uint32        // Nuevo índice después de refresh
    HasCell      bool          // La respuesta contiene Cell
    HasCellIndex bool          // La respuesta contiene CellIndex
}
```

## Estados de Respuesta

| Status | Causa | Cell | CellIndex | NewCellIndex |
|--------|-------|------|-----------|--------------|
| `ok` | Lectura exitosa | ✓ Retornado | ✓ Retornado | ✓ Retornado |
| `unauthorized` | Secret incorrecto | ✗ Vacío | - | - |
| `undefined` | No aplicable | - | - | - |
| `error_db` | Error accediendo BD | ✗ Vacío | - | - |

## Estructura de Genoma (32 bits)

```
Bit 0: LeerSelf      (0x01) - Leer datos propios
Bit 1: LeerAny       (0x02) - Leer datos de otros
Bit 2: LeerLibre     (0x04) - Datos públicos legibles
Bit 3: EscribirSelf  (0x08) - Escribir datos propios
Bit 4: EscribirAny   (0x10) - Escribir datos de otros
Bit 5: BorrarSelf    (0x20) - Borrar datos propios
Bit 6: BorrarAny     (0x40) - Borrar datos de otros
Bit 7: Diferir       (0x80) - Crear célula hijo
Bit 8: Fucionar      (0x100) - Fusionar con otra célula
```

### Ejemplo de Genoma

```go
// Lectura total
const ReadAll = 0x07 // 0b00000111 (LeerSelf|LeerAny|LeerLibre)

// Escritura total
const WriteAll = 0x18 // 0b00011000 (EscribirSelf|EscribirAny)

// Borrado total
const DeleteAll = 0x60 // 0b01100000 (BorrarSelf|BorrarAny)

// Todos los permisos
const AllPermissions = 0xFF // 0b11111111
```

## Proceso de Operación

1. **Resolver Célula** - Buscar célula por index
2. **Validar Autenticación** - Verificar secret contra el hash almacenado
3. **Si válido**:
   - Retornar información completa de la célula
   - Generar nuevo cell index (refresh)
4. **Si inválido**:
   - Retornar error unauthorized

## Ejemplo de Uso - Go

```go
package main

import (
    "fmt"
    "github.com/usuario/samsara-go"
)

func main() {
    store, _ := samsara.New("./data", 1000)
    defer store.Close()

    // Datos conocidos
    cellIndex := uint32(42)
    secret := []byte("mi_secreto")
    
    // Leer información de la célula
    result := store.ReadCell(cellIndex, secret)
    
    if result.Status == samsara.StatusOK {
        cell := result.Cell
        fmt.Printf("Hash: %x\n", cell.Hash)
        fmt.Printf("Salt: %x\n", cell.Salt)
        fmt.Printf("Genoma: 0x%02X\n", cell.Genoma)
        fmt.Printf("Coordenadas: (%d, %d, %d)\n", cell.X, cell.Y, cell.Z)
        fmt.Printf("Nuevo cell index: %d\n", result.NewCellIndex)
        
        // Guardar el nuevo index
        cellIndex = result.NewCellIndex
    } else if result.Status == samsara.StatusUnauthorized {
        fmt.Println("Error: Secret incorrecto")
    }
}
```

## Ejemplo de Uso - Protocolo Binario

```python
import struct

def read_cell_opcode(store, cell_index, secret):
    # Datos
    opcode = 0x05
    secret_bytes = secret.encode('utf-8')
    
    # Construir mensaje
    message = bytearray()
    message.append(opcode)
    message.extend(struct.pack('<I', cell_index))
    message.extend(struct.pack('<I', len(secret_bytes)))
    message.extend(secret_bytes)
    
    # Enviar y recibir
    response = store.send(bytes(message))
    
    # Parsear respuesta
    status = response[0]
    if status == 0:  # ok
        hash = response[1:33]          # 32 bytes
        salt = response[33:49]         # 16 bytes
        genoma = struct.unpack('<I', response[49:53])[0]
        x = struct.unpack('<I', response[53:57])[0]
        y = struct.unpack('<I', response[57:61])[0]
        z = struct.unpack('<I', response[61:65])[0]
        cell_index = struct.unpack('<I', response[65:69])[0]
        new_cell_index = struct.unpack('<I', response[69:73])[0]
        
        return {
            'status': 'ok',
            'cell': {
                'hash': hash.hex(),
                'salt': salt.hex(),
                'genoma': genoma,
                'x': x, 'y': y, 'z': z,
                'current_index': cell_index,
                'new_index': new_cell_index
            }
        }
    
    return {'status': 'error'}
```

## Análisis de Permisos

```go
func analyzePermissions(cell samsara.Celula) {
    fmt.Println("=== Permisos de Célula ===")
    
    // Lectura
    fmt.Printf("Leer propios: %v\n", cell.Genoma&ouroboros.LeerSelf != 0)
    fmt.Printf("Leer ajenos: %v\n", cell.Genoma&ouroboros.LeerAny != 0)
    fmt.Printf("Datos públicos: %v\n", cell.Genoma&ouroboros.LeerLibre != 0)
    
    // Escritura
    fmt.Printf("Escribir propios: %v\n", cell.Genoma&ouroboros.EscribirSelf != 0)
    fmt.Printf("Escribir ajenos: %v\n", cell.Genoma&ouroboros.EscribirAny != 0)
    
    // Borrado
    fmt.Printf("Borrar propios: %v\n", cell.Genoma&ouroboros.BorrarSelf != 0)
    fmt.Printf("Borrar ajenos: %v\n", cell.Genoma&ouroboros.BorrarAny != 0)
    
    // Operaciones avanzadas
    fmt.Printf("Diferir (reproducir): %v\n", cell.Genoma&ouroboros.Diferir != 0)
    fmt.Printf("Fucionar (fusionar): %v\n", cell.Genoma&ouroboros.Fucionar != 0)
}
```

## Casos de Uso

### 1. Verificar Permisos de Célula

```go
func canDeleteData(store *samsara.Store, cellIndex uint32, secret []byte) bool {
    result := store.ReadCell(cellIndex, secret)
    if result.Status != samsara.StatusOK {
        return false
    }
    
    return result.Cell.Genoma&ouroboros.BorrarAny != 0 || 
           result.Cell.Genoma&ouroboros.BorrarSelf != 0
}
```

### 2. Auditar Capacidades de Célula

```go
func auditCell(store *samsara.Store, cellIndex uint32, secret []byte) {
    result := store.ReadCell(cellIndex, secret)
    if result.Status != samsara.StatusOK {
        fmt.Println("No se pudo autenticar")
        return
    }
    
    log.Printf("Célula %d autenticada", cellIndex)
    log.Printf("Hash: %x", result.Cell.Hash)
    log.Printf("Genoma: 0x%02X", result.Cell.Genoma)
    log.Printf("Ubicación: (%d, %d, %d)", result.Cell.X, result.Cell.Y, result.Cell.Z)
}
```

### 3. Validar Coordenadas

```go
func validateCellLocation(store *samsara.Store, cellIndex uint32, secret []byte) error {
    result := store.ReadCell(cellIndex, secret)
    if result.Status != samsara.StatusOK {
        return fmt.Errorf("auth failed")
    }
    
    const MAX_COORD = 1000000
    if result.Cell.X > MAX_COORD || result.Cell.Y > MAX_COORD || result.Cell.Z > MAX_COORD {
        return fmt.Errorf("coordenadas fuera de rango")
    }
    
    return nil
}
```

## Estructura de Respuesta Binaria

```
Posición | Tamaño | Campo
---------|--------|-------
0        | 1      | Status
1-32     | 32     | Hash (BLAKE3)
33-48    | 16     | Salt
49-52    | 4      | Genoma (uint32 LE)
53-56    | 4      | X (uint32 LE)
57-60    | 4      | Y (uint32 LE)
61-64    | 4      | Z (uint32 LE)
65-68    | 4      | CellIndex (uint32 LE)
69-72    | 4      | NewCellIndex (uint32 LE)
```

## Características del Hash BLAKE3

- **Algoritmo**: BLAKE3 (seguro criptográfico)
- **Tamaño**: 256 bits (32 bytes)
- **Entrada**: Hash(salt + secret)
- **Uso**: Autenticación sin almacenar secret en texto plano

## Características del Salt

- **Tamaño**: 128 bits (16 bytes)
- **Generación**: Aleatorio criptográfico
- **Propósito**: Prevenir rainbow table attacks
- **Almacenamiento**: Se guarda con la célula

## Coordenadas Espaciales

Las células pueden tener coordenadas X, Y, Z que representan:

- **Ubicación lógica** en el sistema
- **Relaciones geométricas** entre células (distancia)
- **Particionamiento de datos** basado en espacio
- **Optimizaciones de consulta** por proximidad

Usos comunes:
```go
// Organizar por región
cell.X = userID % 100
cell.Y = departmentID % 100
cell.Z = priority % 10

// O usar como contexto
cell.X = shardID
cell.Y = bucketID
cell.Z = reserved
```

## Casos de Error

### Error: `unauthorized`

- Secret incorrecto para la célula
- Cell index no válido
- Célula fue eliminada

Solución:
```go
// Reintentar solo si el error es temporal
for i := 0; i < 3; i++ {
    result := store.ReadCell(cellIndex, secret)
    if result.Status == samsara.StatusOK {
        return result
    }
    time.Sleep(100 * time.Millisecond)
}
```

### Error: `error_db`

- Problema de conectividad
- Base de datos corrupta

Solución:
```go
// Implementar reintentos con backoff
if result.Status == samsara.StatusErrorDB {
    return retryWithBackoff(func() error {
        result := store.ReadCell(cellIndex, secret)
        if result.Status != samsara.StatusOK {
            return fmt.Errorf("read cell failed")
        }
        return nil
    })
}
```

## Integración en API REST

```go
func (h *Handler) GetCellInfo(w http.ResponseWriter, r *http.Request) {
    var req struct {
        CellIndex uint32 `json:"cell_index"`
        Secret    string `json:"secret"`
    }
    
    json.NewDecoder(r.Body).Decode(&req)
    
    result := h.store.ReadCell(req.CellIndex, []byte(req.Secret))
    
    if result.Status != samsara.StatusOK {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "genoma": fmt.Sprintf("0x%02X", result.Cell.Genoma),
        "coordinates": map[string]uint32{
            "x": result.Cell.X,
            "y": result.Cell.Y,
            "z": result.Cell.Z,
        },
        "new_cell_index": result.NewCellIndex,
    })
}
```

## Comparación con Otros Opcodes

| Opcode | Lee Datos | Lee Metadata | Requiere Auth | Retorna Nuevo Index |
|--------|---|---|---|---|
| READ_CELL | ✗ No | ✓ Sí | ✓ Sí | ✓ Sí |
| READ | ✓ Sí | ✗ No | ✓ Sí | ✓ Sí |
| READ_FREE | ✓ Sí | ✗ No | ✗ No | ✗ No |

