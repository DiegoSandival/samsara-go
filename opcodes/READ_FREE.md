# Opcode: READ_FREE (0x02)

## Descripción

Lee un valor almacenado en Samsara de forma pública, sin requerir autenticación. El dato debe estar marcado como público por su propietario.

## Información del Opcode

| Propiedad | Valor |
|-----------|-------|
| Código | `0x02` |
| Requiere Autenticación | **No** |
| Requiere Cell Index | **No** |
| Retorna Nuevo Cell Index | **No** |
| Tipo | Lectura (Pública) |

## Parámetros

| Parámetro | Tipo | Tamaño | Descripción |
|-----------|------|--------|------------|
| `key` | String | Variable | Clave del dato a leer |

## Estructura Binaria

```
[Opcode: 0x02] [Key Length: 4 bytes] [Key: N bytes]
```

## Proceso de Validación

1. **Buscar Dato** - Obtener el dato usando la clave
2. **Verificar Propietario** - Obtener la célula propietaria del dato
3. **Validar Permiso Público** - Verificar que el propietario tiene flag `LeerLibre`
4. **Retornar Valor** - Si todo es correcto, retornar el valor

## Resultado

```go
type FreeReadResult struct {
    Status   Status  // ok | unauthorized | undefined | error_db
    Value    []byte  // Valor leído (solo si Status == ok)
    HasValue bool    // La respuesta contiene Value
}
```

## Estados de Respuesta

| Status | Causa | Value |
|--------|-------|-------|
| `ok` | Lectura exitosa | ✓ Retornado |
| `unauthorized` | Propietario no tiene permiso `LeerLibre` | ✗ Vacío |
| `undefined` | La clave no existe | ✗ Vacío |
| `error_db` | Error accediendo base de datos | ✗ Vacío |

## Restricciones de Permisos

### Permiso de Lectura Libre

El **propietario del dato** debe tener el flag `LeerLibre` (0x04) en su genoma:

```
Genoma de propietario = 0b00000100 (LeerLibre habilitado)

Lectura pública: Válida ✓
Lectura sin permiso: Rechazada ✗
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

    // Leer un valor público
    result := store.ReadFree("anuncio:mensaje_bienvenida")
    
    if result.Status == samsara.StatusOK {
        fmt.Printf("Anuncio: %s\n", string(result.Value))
    } else if result.Status == samsara.StatusUnauthorized {
        fmt.Println("Error: Los datos no son públicos")
    } else if result.Status == samsara.StatusUndefined {
        fmt.Println("Error: El dato no existe")
    }
}
```

## Ejemplo de Uso - Protocolo Binario

```python
import struct

def read_free_opcode(store, key):
    # Datos
    opcode = 0x02
    key_bytes = key.encode('utf-8')
    
    # Construir mensaje
    message = bytearray()
    message.append(opcode)
    message.extend(struct.pack('<I', len(key_bytes)))
    message.extend(key_bytes)
    
    # Enviar y recibir
    response = store.send(bytes(message))
    
    # Parsear respuesta
    status = response[0]
    if status == 0:  # ok
        value_len = struct.unpack('<I', response[1:5])[0]
        value = response[5:5+value_len]
        return {'status': 'ok', 'value': value}
    
    return {'status': 'error'}
```

## Configuración de Datos Públicos

### Paso 1: Crear Célula con Permiso Público

```go
// Crear célula con flag LeerLibre habilitado
salt := [16]byte{...}
secret := []byte("secreto")
genoma := uint32(0x04) // LeerLibre
cell := samsara.NewCellWithSecret(salt, secret, genoma, 0, 0, 0)
cellIndex, _ := store.DB().Append(cell)
```

### Paso 2: Escribir Dato Público

```go
// El dato se almacena con el propietario de la célula
result := store.Write("anuncio:bienvenida", []byte("¡Bienvenido!"), cellIndex, secret)
// Ahora cualquiera puede leerlo con READ_FREE
```

## Casos de Uso

### ✓ Datos Públicos Ideales

- Anuncios y noticias
- Información de catálogo
- Datos de lectura pública general
- Metadatos públicos

### ✗ Datos que NO Deben ser Públicos

- Credenciales
- Datos personales
- Información confidencial
- Datos de transacciones privadas

## Comparación con READ

| Aspecto | READ_FREE | READ |
|---------|-----------|------|
| Requiere Autenticación | No | Sí |
| Requiere Secret | No | Sí |
| Requiere Cell Index | No | Sí |
| Retorna Nuevo Cell Index | No | Sí |
| Seguridad | Baja | Alta |
| Ideal Para | Datos públicos | Datos privados |
| Permiso Requerido | `LeerLibre` | `LeerSelf` o `LeerAny` |

## Notas de Implementación

- **Caché Agresivo** - READ_FREE es seguro cachear por más tiempo
- **Sin Auditoría de Acceso** - Las lecturas públicas no se rastrean por célula
- **Performance** - Lee directamente sin validar autenticación
- **Disponibilidad** - Parte de tu API pública

## Casos de Error Comunes

### Error: `unauthorized`

- El propietario del dato no tiene activado el flag `LeerLibre`
- El dato existe pero no es público

### Error: `undefined`

- La clave no existe en la base de datos
- El dato fue eliminado

### Error: `error_db`

- Problema con la base de datos
- Corrupción de datos

## Flujo Típico de Lectura Pública

```
1. Cliente solicita READ_FREE("datos:publicos")
   └─ Sin autenticación
   └─ Sin permisos especiales

2. Servidor busca la clave
   └─ Localiza el dato

3. Servidor valida propietario
   └─ Obtiene célula propietaria
   └─ Verifica flag LeerLibre

4. Si válido → Retorna dato
   └─ Status: ok
   └─ Value: datos solicitados

5. Si inválido → Rechaza acceso
   └─ Status: unauthorized (propietario sin permiso)
   └─ Value: vacío
```

## Integración en API REST

```go
// Endpoint público, sin autenticación
func (h *Handler) GetPublicData(w http.ResponseWriter, r *http.Request) {
    key := r.URL.Query().Get("key")
    
    result := h.store.ReadFree(key)
    
    if result.Status != samsara.StatusOK {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.Write(result.Value)
}
```

## Seguridad

### ✓ Seguro Cachear

- Los datos públicos no cambian de privacidad constantemente
- Cachear por 1-5 minutos es seguro
- CDN friendly

### ⚠️ Validar Siempre

- Incluso datos públicos deben validarse
- Verificar integridad si es crítico
- Considerar firmas digitales para datos sensibles

### 🔒 Monitoreo

- Registra accesos a datos públicos críticos
- Detecta patrones de abuso
- Implementa rate limiting si es necesario
