# Opcode: DELETE (0x04)

## Descripción

Elimina un valor almacenado en Samsara. Requiere autenticación y validación de permisos para el dato a eliminar.

## Información del Opcode

| Propiedad | Valor |
|-----------|-------|
| Código | `0x04` |
| Requiere Autenticación | **Sí** |
| Requiere Cell Index | **Sí** |
| Retorna Nuevo Cell Index | **Sí** |
| Tipo | Eliminación (Modificación) |

## Parámetros

| Parámetro | Tipo | Tamaño | Descripción |
|-----------|------|--------|------------|
| `key` | String | Variable | Clave del dato a eliminar |
| `cellIndex` | uint32 | 4 bytes | Índice de la célula que ejecuta la eliminación |
| `secret` | []byte | Variable | Contraseña de la célula para autenticación |

## Estructura Binaria

```
[Opcode: 0x04] [Key Length: 4 bytes] [Key: N bytes] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: M bytes]
```

## Proceso de Operación

1. **Resolver Célula** - Obtener la célula usando cellIndex y secret
2. **Validar Autenticación** - Verificar que el secret es correcto
3. **Buscar Dato** - Localizar el dato con la clave especificada
4. **Validar Permisos**
   - Si cell owner == data owner → Requiere `BorrarSelf`
   - Si cell owner != data owner → Requiere `BorrarAny`
5. **Eliminar Dato** - Remover el dato de la base de datos
6. **Refrescar Célula** - Retornar nuevo cell index

## Resultado

```go
type DeleteResult struct {
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
| `ok` | Eliminación exitosa | - | ✓ Retornado |
| `unauthorized` | Secret inválido, dato no existe, o permisos insuficientes | ✓ Diagnostico | - |
| `undefined` | La clave no existe | ✓ Diagnostico | - |
| `error_db` | Error accediendo base de datos | - | - |

## Restricciones de Permisos

### Para Eliminar un Dato Existente

Según la relación entre propietarios:

| Escenario | Flag Requerido | Descripción |
|-----------|---|---|
| Mismo propietario | `BorrarSelf` (0x20) | Eliminar sus propios datos |
| Otro propietario | `BorrarAny` (0x40) | Eliminar datos de otros |

### Ejemplo de Validación de Genoma

```
Genoma de célula = 0b01100000 (BorrarSelf | BorrarAny)

Eliminar dato propio: Válido ✓
Eliminar dato ajeno: Válido ✓
Sin permisos de borrado: Rechazado ✗
```

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
    
    // Crear célula con permisos de borrado
    cellGenome := uint32(0xFF) // Todos los permisos
    cell := samsara.NewCellWithSecret(salt, secret, cellGenome, 0, 0, 0)
    cellIndex, _ := store.DB().Append(cell)
    
    // Primero escribir un dato
    store.Write("temp:datos", []byte("temporal"), cellIndex, secret)
    
    // Luego eliminarlo
    result := store.Delete("temp:datos", cellIndex, secret)
    
    if result.Status == samsara.StatusOK {
        fmt.Printf("Eliminación exitosa\n")
        fmt.Printf("Nuevo cell index: %d\n", result.NewCellIndex)
        cellIndex = result.NewCellIndex
    } else if result.Status == samsara.StatusUnauthorized {
        fmt.Println("Error: Permisos insuficientes para eliminar")
    } else if result.Status == samsara.StatusUndefined {
        fmt.Println("Error: El dato no existe")
    } else if result.Status == samsara.StatusErrorDB {
        fmt.Println("Error: Problema con la base de datos")
    }
}
```

## Ejemplo de Uso - Protocolo Binario

```python
import struct

def delete_opcode(store, key, cell_index, secret):
    # Datos
    opcode = 0x04
    key_bytes = key.encode('utf-8')
    secret_bytes = secret.encode('utf-8')
    
    # Construir mensaje
    message = bytearray()
    message.append(opcode)
    message.extend(struct.pack('<I', len(key_bytes)))
    message.extend(key_bytes)
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
    elif status == 2:  # undefined
        current_cell_index = struct.unpack('<I', response[1:5])[0]
        return {'status': 'undefined', 'current_cell_index': current_cell_index}
    
    return {'status': 'error'}
```

## Casos de Uso

### ✓ Eliminar Datos Personales (GDPR)

```go
func deleteUserData(store *samsara.Store, cellIndex uint32, secret []byte, userID string) error {
    // Eliminar todos los datos del usuario
    keys := []string{
        "user:" + userID + ":profile",
        "user:" + userID + ":preferences",
        "user:" + userID + ":history",
    }
    
    for _, key := range keys {
        result := store.Delete(key, cellIndex, secret)
        if result.Status != samsara.StatusOK {
            return fmt.Errorf("failed to delete %s: %s", key, result.Status)
        }
        cellIndex = result.NewCellIndex
    }
    
    return nil
}
```

### ✓ Limpiar Datos Temporales

```go
func cleanupTempData(store *samsara.Store, cellIndex uint32, secret []byte) {
    tempKeys := []string{
        "temp:session:12345",
        "temp:upload:67890",
        "cache:request:abc",
    }
    
    for _, key := range tempKeys {
        result := store.Delete(key, cellIndex, secret)
        if result.Status == samsara.StatusOK {
            cellIndex = result.NewCellIndex
            log.Printf("Cleaned up: %s", key)
        }
    }
}
```

### ✓ Revocar Acceso a Datos

```go
// Un admin puede eliminar datos de usuarios (si tiene BorrarAny)
result := store.Delete("user:malicioso:datos", adminCellIndex, adminSecret)
if result.Status == samsara.StatusOK {
    log.Println("Acceso revocado exitosamente")
}
```

## Ciclo de Vida de Datos

```
Creación (WRITE)
        ↓
Lectures múltiples (READ)
        ↓
Actualizaciones (WRITE)
        ↓
Eliminación (DELETE)
        ↓
        ∅ (No existe más)
```

## Casos de Error Comunes

### Error: `unauthorized`

- Secret incorrecto
- Célula no existe
- Eliminar dato ajeno sin permiso `BorrarAny`
- No tener permiso `BorrarSelf` para datos propios
- Genoma de célula no incluye flags de borrado

Solución:
```go
// Verificar permisos antes de intentar delete
cellResult := store.ReadCell(cellIndex, secret)
if cellResult.Status != samsara.StatusOK {
    // Autenticación falló
    return
}

// Verificar que tenemos permisos de borrado
if cellResult.Cell.Genoma&ouroboros.BorrarAny == 0 {
    // No tenemos permisos
    return
}

// Ahora proceder con el delete
deleteResult := store.Delete(key, cellIndex, secret)
```

### Error: `undefined`

- La clave no existe o ya fue eliminada
- El dato fue eliminado por otro proceso
- Se escribió mal la clave

Solución:
```go
// Verificar existencia antes de eliminar
readResult := store.Read(key, cellIndex, secret)
if readResult.Status == samsara.StatusUndefined {
    fmt.Println("El dato ya no existe")
    return
}

// Proceder con delete
deleteResult := store.Delete(key, cellIndex, secret)
```

### Error: `error_db`

- Problema de conexión con base de datos
- Información de célula corrupta
- Error crítico del sistema

Solución:
```go
// Implementar reintentos
for i := 0; i < 3; i++ {
    result := store.Delete(key, cellIndex, secret)
    if result.Status != samsara.StatusErrorDB {
        return result
    }
    time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
}
```

## Buenas Prácticas

### ✓ Hacer

- Verificar permiso antes de eliminar
- Registrar eliminaciones importantes
- Implementar soft-delete para datos auditables
- Usar transacciones para múltiples eliminaciones
- Tener política de retención de datos

### ✗ Evitar

- Eliminar sin validar autenticación
- Ignorar estado de error
- Eliminar en loop sin validar cada una
- Eliminar datos sin backup
- Confiar en que el servidor siempre tiene permisos

## Patrón: Backup Antes de Eliminar

```go
func safeDelete(store *samsara.Store, key string, cellIndex uint32, secret []byte) error {
    // Leer dato antes de eliminar
    readResult := store.Read(key, cellIndex, secret)
    if readResult.Status != samsara.StatusOK {
        return fmt.Errorf("read failed: %w", readResult.Status)
    }
    
    // Guardar backup
    backupKey := "backup:" + key + ":" + time.Now().Format(time.RFC3339)
    backupResult := store.Write(backupKey, readResult.Value, readResult.NewCellIndex, secret)
    if backupResult.Status != samsara.StatusOK {
        return fmt.Errorf("backup failed: %w", backupResult.Status)
    }
    
    // Proceder con delete
    deleteResult := store.Delete(key, backupResult.NewCellIndex, secret)
    if deleteResult.Status != samsara.StatusOK {
        return fmt.Errorf("delete failed: %w", deleteResult.Status)
    }
    
    return nil
}
```

## Integración en API REST

```go
// Endpoint para eliminar datos
func (h *Handler) DeleteData(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Key       string `json:"key"`
        CellIndex uint32 `json:"cell_index"`
        Secret    string `json:"secret"`
    }
    
    json.NewDecoder(r.Body).Decode(&req)
    
    result := h.store.Delete(req.Key, req.CellIndex, []byte(req.Secret))
    
    if result.Status != samsara.StatusOK {
        http.Error(w, string(result.Status), http.StatusUnauthorized)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "deleted",
        "new_cell_index": result.NewCellIndex,
    })
}
```

## Monitoreo y Auditoría

```go
// Registrar operaciones de eliminación
func logDelete(key string, status samsara.Status, cellIndex uint32) {
    log.Printf("DELETE: key=%s status=%s cell=%d timestamp=%s",
        key, status, cellIndex, time.Now().Format(time.RFC3339))
        
    // Para datos críticos, alertar
    if status == samsara.StatusOK && isImportantKey(key) {
        alertAdmins("Critical data deleted: " + key)
    }
}
```

## Recuperación ante Errores

Tabla de decisión para manejar errores en DELETE:

| Status | Reintentable | Acción Sugerida |
|--------|---|---|
| `ok` | No | Completar |
| `unauthorized` | No | Verificar permisos y autenticación |
| `undefined` | No | Confirmar que ya estuvo eliminado |
| `error_db` | Sí | Reintentar con backoff exponencial |

