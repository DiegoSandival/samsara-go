# Opcode: READ (0x01)

## Descripción

Lee un valor almacenado en Samsara con autenticación. El acceso está protegido por permisos de la célula que realiza la operación.

## Información del Opcode

| Propiedad | Valor |
|-----------|-------|
| Código | `0x01` |
| Requiere Autenticación | **Sí** |
| Requiere Cell Index | **Sí** |
| Retorna Nuevo Cell Index | **Sí** |
| Tipo | Lectura (Segura) |

## Parámetros

| Parámetro | Tipo | Tamaño | Descripción |
|-----------|------|--------|------------|
| `key` | String | Variable | Clave del dato a leer |
| `cellIndex` | uint32 | 4 bytes | Índice de la célula que ejecuta la lectura |
| `secret` | []byte | Variable | Contraseña de la célula para autenticación |

## Estructura Binaria

```
[Opcode: 0x01] [Key Length: 4 bytes] [Key: N bytes] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: M bytes]
```

## Proceso de Validación

1. **Resolver Célula** - Obtener la célula usando cellIndex y secret
2. **Validar Autenticación** - Verificar que el secret es correcto usando BLAKE3
3. **Verificar Permiso de Lectura**
   - Si cell owner == data owner → Requiere `LeerSelf`
   - Si cell owner != data owner → Requiere `LeerAny`
4. **Refrescar Célula** - Si todo es correcto, se retorna nuevo cell index

## Resultado

```go
type ReadResult struct {
    Status       Status  // ok | unauthorized | undefined | error_db
    Value        []byte  // Valor leído (solo si Status == ok)
    CellIndex    uint32  // Índice de célula actual (diagnostico)
    NewCellIndex uint32  // Nuevo índice de célula (después de refresh)
    HasCellIndex bool    // La respuesta contiene CellIndex
    HasNewCell   bool    // La respuesta contiene NewCellIndex
    HasValue     bool    // La respuesta contiene Value
}
```

## Estados de Respuesta

| Status | Causa | Value | NewCellIndex |
|--------|-------|-------|--------------|
| `ok` | Lectura exitosa | ✓ Retornado | ✓ Retornado |
| `unauthorized` | Secret inválido o permisos insuficientes | ✗ Vacío | - |
| `undefined` | La clave no existe | ✗ Vacío | ✓ Retornado |
| `error_db` | Error accediendo base de datos | ✗ Vacío | ✗ Vacío |

## Restricciones de Permisos

### Propietario del Dato

La célula que realiza la lectura debe tener los permisos correspondientes:

- Si la célula es el **propietario del dato** → Requiere flag `LeerSelf` (0x01)
- Si la célula es **otro propietario** → Requiere flag `LeerAny` (0x02)

### Permisos en Genoma

```
Genoma de célula = 0b00000011 (LeerSelf | LeerAny habilitados)

Lectura de dato propios: Válido ✓
Lectura de dato ajeno: Válido ✓
Lectura sin permisos: Rechazado ✗
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
    
    // Crear célula con permisos de lectura
    cellGenome := uint32(0xFF) // Todos los permisos
    cell := samsara.NewCellWithSecret(salt, secret, cellGenome, 0, 0, 0)
    cellIndex, _ := store.DB().Append(cell)
    
    // Leer un valor
    result := store.Read("usuario:nombre", cellIndex, secret)
    
    if result.Status == samsara.StatusOK {
        fmt.Printf("Valor: %s\n", string(result.Value))
        fmt.Printf("Nuevo cell index: %d\n", result.NewCellIndex)
    } else if result.Status == samsara.StatusUnauthorized {
        fmt.Println("Error: Permisos insuficientes")
    } else if result.Status == samsara.StatusUndefined {
        fmt.Println("Error: La clave no existe")
    }
}
```

## Ejemplo de Uso - Protocolo Binario

```python
import struct

def read_opcode(store, key, cell_index, secret):
    # Datos
    opcode = 0x01
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
        value_len = struct.unpack('<I', response[1:5])[0]
        value = response[5:5+value_len]
        new_cell_index = struct.unpack('<I', response[5+value_len:9+value_len])[0]
        return {'status': 'ok', 'value': value, 'new_cell_index': new_cell_index}
    
    return {'status': 'error'}
```

## Comportamiento de Refresh

Cuando una lectura es exitosa, el sistema "refresca" la célula, retornando un nuevo cell index. Esto permite:

- **Rastreo de actividad** - Cada operación avanza la posición de la célula
- **Seguridad temporal** - Previene reutilización de índices antiguos
- **Validación de estado** - Permite verificar que la célula sigue siendo válida

## Casos de Error Comunes

### Error: `unauthorized`

- Secret incorrecto
- Célula no existe en el índice especificado
- Célula no tiene permiso `LeerSelf` o `LeerAny` según corresponda
- Dato está protegido y la célula es de otro propietario

### Error: `undefined`

- La clave no existe en la base de datos
- El dato fue eliminado previamente

### Error: `error_db`

- Problema con la base de datos Ouroboros
- Problema con la base de datos de membranas (BoltDB)
- Corrupción de datos

## Notas de Implementación

- **Timeout** - Considera implementar timeouts para lecturas
- **Caché** - Los datos leídos pueden cachearse localmente
- **Logging** - Registra intentos de lectura para auditoría
- **Cell Rotation** - Siempre usa el `NewCellIndex` retornado para próximas operaciones

## Comparación con READ_FREE

| Aspecto | READ | READ_FREE |
|---------|------|-----------|
| Requiere Autenticación | Sí | No |
| Requiere Permiso Específico | Sí | Propietario necesita `LeerLibre` |
| Retorna Nuevo Cell Index | Sí | No |
| Ideal Para | Datos privados | Datos públicos |
| Seguridad | Alta | Baja |
