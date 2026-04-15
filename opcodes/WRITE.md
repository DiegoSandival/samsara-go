# Opcode: WRITE (0x03)

## Descripción

Escribe o actualiza un valor en Samsara. Crea una nueva entrada si no existe, o actualiza la existente si el permiso lo permite. Requiere autenticación obligatoria.

## Información del Opcode

| Propiedad | Valor |
|-----------|-------|
| Código | `0x03` |
| Requiere Autenticación | **Sí** |
| Requiere Cell Index | **Sí** |
| Retorna Nuevo Cell Index | **Sí** |
| Tipo | Escritura (Modificación) |

## Parámetros

| Parámetro | Tipo | Tamaño | Descripción |
|-----------|------|--------|------------|
| `key` | String | Variable | Clave donde almacenar el dato |
| `value` | []byte | Variable | Valor a escribir |
| `cellIndex` | uint32 | 4 bytes | Índice de la célula que ejecuta la escritura |
| `secret` | []byte | Variable | Contraseña de la célula para autenticación |

## Estructura Binaria

```
[Opcode: 0x03] [Key Length: 4 bytes] [Key: N bytes] [Value Length: 4 bytes] [Value: M bytes] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: P bytes]
```

## Proceso de Operación

### Caso 1: Crear Nuevo Dato

1. **Resolver Célula** - Obtener la célula usando cellIndex y secret
2. **Validar Autenticación** - Verificar que el secret es correcto
3. **Crear Membrana** - Almacenar nuevo dato con owner = cellIndex
4. **Refrescar Célula** - Retornar nuevo cell index

### Caso 2: Actualizar Dato Existente

1. **Resolver Célula** - Obtener la célula autenticada
2. **Buscar Dato** - Localizar el dato existente
3. **Validar Permisos**
   - Si cell owner == data owner → Requiere `EscribirSelf`
   - Si cell owner != data owner → Requiere `EscribirAny`
4. **Actualizar Valor** - Modificar el contenido del dato
5. **Refrescar Célula** - Retornar nuevo cell index

## Resultado

```go
type WriteResult struct {
    Status       Status  // ok | unauthorized | undefined | error_db
    CellIndex    uint32  // Índice de célula actual
    NewCellIndex uint32  // Nuevo índice de célula (después de refresh)
    HasCellIndex bool    // La respuesta contiene CellIndex
    HasNewCell   bool    // La respuesta contiene NewCellIndex
}
```

## Estados de Respuesta

| Status | Causa | CellIndex | NewCellIndex |
|--------|-------|-----------|--------------|
| `ok` | Escritura exitosa | - | ✓ Retornado |
| `unauthorized` | Secret inválido o permisos insuficientes | ✓ Diagnostico | - |
| `undefined` | No aplicable (siempre crea si no existe) | - | - |
| `error_db` | Error accediendo base de datos | - | - |

## Restricciones de Permisos

### Crear Nuevo Dato

Requerimiento mínimo:
- Célula debe estar autenticada correctamente
- No hay restricción adicional (el propietario será la célula)

### Actualizar Dato Existente

Según la relación entre propietarios:

| Escenario | Flag Requerido | Validación |
|-----------|---|---|
| Mismo propietario | `EscribirSelf` (0x08) | Debe existir en genoma |
| Otro propietario | `EscribirAny` (0x10) | Debe existir en genoma |

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

    // Datos de la célula
    salt := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
    secret := []byte("mi_secreto_super_seguro")
    
    // Crear célula con permisos de escritura
    cellGenome := uint32(0xFF) // Todos los permisos
    cell := samsara.NewCellWithSecret(salt, secret, cellGenome, 0, 0, 0)
    cellIndex, _ := store.DB().Append(cell)
    
    // Escribir un valor
    value := []byte("Datos importantes")
    result := store.Write("usuario:datos", value, cellIndex, secret)
    
    if result.Status == samsara.StatusOK {
        fmt.Printf("Escritura exitosa\n")
        fmt.Printf("Nuevo cell index: %d\n", result.NewCellIndex)
        cellIndex = result.NewCellIndex // Guardar para próximas operaciones
    } else if result.Status == samsara.StatusUnauthorized {
        fmt.Println("Error: Permisos insuficientes")
    } else if result.Status == samsara.StatusErrorDB {
        fmt.Println("Error: Problema con la base de datos")
    }
}
```

## Ejemplo de Uso - Protocolo Binario

```python
import struct

def write_opcode(store, key, value, cell_index, secret):
    # Datos
    opcode = 0x03
    key_bytes = key.encode('utf-8')
    value_bytes = value.encode('utf-8') if isinstance(value, str) else value
    secret_bytes = secret.encode('utf-8')
    
    # Construir mensaje
    message = bytearray()
    message.append(opcode)
    message.extend(struct.pack('<I', len(key_bytes)))
    message.extend(key_bytes)
    message.extend(struct.pack('<I', len(value_bytes)))
    message.extend(value_bytes)
    message.extend(struct.pack('<I', cell_index))
    message.extend(struct.pack('<I', len(secret_bytes)))
    message.extend(secret_bytes)
    
    # Enviar y recibir
    response = store.send(bytes(message))
    
    # Parsear respuesta
    status = response[0]
    if status == 0:  # ok
        new_cell_index = struct.unpack('<I', response[1:5])[0]
        return {'status': 'ok', 'new_cell_index': new_cell_index}
    elif status == 1:  # unauthorized
        current_cell_index = struct.unpack('<I', response[1:5])[0]
        return {'status': 'unauthorized', 'current_cell_index': current_cell_index}
    
    return {'status': 'error'}
```

## Operaciones Comunes

### Crear Nuevo Dato

```go
// Crear y guardar un nuevo dato
result := store.Write("nuevo:datos", []byte("contenido"), cellIndex, secret)
if result.Status == samsara.StatusOK {
    // Guardar el nuevo cell index para operaciones futuras
    cellIndex = result.NewCellIndex
}
```

### Actualizar Dato Existente

```go
// Leer datos actuales
readResult := store.Read("usuario:perfil", cellIndex, secret)
if readResult.Status == samsara.StatusOK {
    // Modificar
    newData := modifyData(readResult.Value)
    
    // Guardar
    writeResult := store.Write("usuario:perfil", newData, readResult.NewCellIndex, secret)
    if writeResult.Status == samsara.StatusOK {
        cellIndex = writeResult.NewCellIndex
    }
}
```

### Pattern: Read-Modify-Write

```go
func updateUserProfile(store *samsara.Store, cellIndex uint32, secret []byte, updates map[string]string) error {
    // Leer perfil actual
    readResult := store.Read("user:profile", cellIndex, secret)
    if readResult.Status != samsara.StatusOK {
        return fmt.Errorf("read failed: %s", readResult.Status)
    }
    
    // Modificar (deserializar, actualizar, serializar)
    profile := deserialize(readResult.Value)
    for key, val := range updates {
        profile[key] = val
    }
    newData := serialize(profile)
    
    // Escribir modificación
    writeResult := store.Write("user:profile", newData, readResult.NewCellIndex, secret)
    if writeResult.Status != samsara.StatusOK {
        return fmt.Errorf("write failed: %s", writeResult.Status)
    }
    
    return nil
}
```

## Tamaños Recomendados

| Tipo de Dato | Tamaño Máximo Recomendado | Ejemplos |
|--------------|---------------------------|----------|
| Texto corto | < 1 KB | Nombres, tags, URLs |
| Datos estructurados | 1-100 KB | JSON, perfiles, configuración |
| Documentos | 100 KB - 1 MB | Textos, artículos |
| Contenido grande | > 1 MB | Considerar fragmentación |

## Casos de Error Comunes

### Error: `unauthorized`

- Secret incorrecto
- Célula no existe en el índice especificado
- Actualizando dato ajeno sin permiso `EscribirAny`
- Genoma de célula no tiene flags requeridos

### Error: `error_db`

- Base de datos llena
- Problema de permisos del archivo
- Corrupción de datos
- Error de BoltDB en membranas

## Buenas Prácticas

### ✓ Hacer

- Validar datos antes de escribir
- Guardar el `NewCellIndex` retornado
- Implementar reintentos con backoff exponencial
- Usar transacciones para múltiples escrituras
- Loguear operaciones críticas

### ✗ Evitar

- Escribir datos sin validar
- Ignorar el `NewCellIndex` retornado
- Escribir datos enormes sin fragmenting
- Confiar en que el cell index nunca cambia
- Escribir datos sensibles sin encriptar

## Integración en API REST

```go
// Endpoint para escribir datos
func (h *Handler) WriteData(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Key       string `json:"key"`
        Value     string `json:"value"`
        CellIndex uint32 `json:"cell_index"`
        Secret    string `json:"secret"`
    }
    
    json.NewDecoder(r.Body).Decode(&req)
    
    result := h.store.Write(req.Key, []byte(req.Value), req.CellIndex, []byte(req.Secret))
    
    if result.Status != samsara.StatusOK {
        http.Error(w, string(result.Status), http.StatusUnauthorized)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

## Monitoreo y Auditoría

```go
// Registrar operaciones de escritura
func logWrite(key string, size int, status samsara.Status) {
    log.Printf("WRITE: key=%s size=%d status=%s timestamp=%s",
        key, size, status, time.Now().Format(time.RFC3339))
}
```

## Comparación con Otros Opcodes

| Opcode | Modifica Datos | Requiere Permisos | Idempotente |
|--------|---|---|---|
| WRITE | ✓ Sí | ✓ Sí | ✗ No (siempre actualiza) |
| READ | ✗ No | ✓ Sí | ✓ Sí |
| DELETE | ✓ Sí | ✓ Sí | ✗ No |

