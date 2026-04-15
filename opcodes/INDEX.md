# Índice de Opcodes - Samsara

Directorio completo de todos los opcodes disponibles en la librería Samsara. Cada opcode representa una operación que puede realizarse en el sistema de almacenamiento basado en células.

## Opcodes Disponibles

### Operaciones de Datos

| Opcode | Código | Descripción | Requiere Auth | Parámetros |
|--------|--------|-------------|---------------|-----------|
| [READ](READ.md) | `0x01` | Leer un valor con autenticación | Sí | `key, cellIndex, secret` |
| [READ_FREE](READ_FREE.md) | `0x02` | Leer un valor público (sin autenticación) | No | `key` |
| [WRITE](WRITE.md) | `0x03` | Escribir/actualizar un valor | Sí | `key, value, cellIndex, secret` |
| [DELETE](DELETE.md) | `0x04` | Eliminar un valor | Sí | `key, cellIndex, secret` |

### Operaciones de Célula

| Opcode | Código | Descripción | Requiere Auth | Parámetros |
|--------|--------|-------------|---------------|-----------|
| [READ_CELL](READ_CELL.md) | `0x05` | Leer información de una célula | Sí | `cellIndex, secret` |
| [DIFERIR](DIFERIR.md) | `0x06` | Crear una célula hijo desde una célula padre | Sí | `cellIndex, parentSecret, childSecret, childSalt, childGenome, x, y, z` |
| [CRUZAR](CRUZAR.md) | `0x07` | Fusionar dos células para crear una célula hijo | Sí | `cellIndexA, secretA, cellIndexB, secretB, childSecret, childSalt, x, y, z` |

## Estados de Respuesta

Todos los opcodes retornan un estado que indica el resultado de la operación:

- **`ok`** - Operación exitosa
- **`unauthorized`** - Autenticación fallida o permisos insuficientes
- **`undefined`** - El recurso no existe
- **`error_db`** - Error en la base de datos

## Estructura General de Solicitud

```
[Opcode: 1 byte]
[Parámetros específicos del opcode]
```

## Permisos de Célula (Genoma)

Las células en Samsara tienen un sistema de permisos basado en banderas de genoma:

- `LeerSelf` - Leer datos propios
- `LeerAny` - Leer datos de otros propietarios
- `LeerLibre` - Leer datos públicos
- `EscribirSelf` - Escribir datos propios
- `EscribirAny` - Escribir datos de otros propietarios
- `BorrarSelf` - Borrar datos propios
- `BorrarAny` - Borrar datos de otros propietarios
- `Diferir` - Crear células hijo
- `Fucionar` - Fusionar células

## Autenticación

Las operaciones autenticadas requieren:

1. **Cell Index** - Índice de la célula que realiza la operación
2. **Secret** - Contraseña/secreto de la célula (usada para derivar el hash)

La autenticación se valida mediante encriptación BLAKE3 sobre la célula.

## Guía de Uso

Consulta la documentación individual de cada opcode para:

- Estructura binaria exacta
- Parámetros detallados
- Tipos de respuesta
- Ejemplo de código
- Casos de error comunes

Comienza con los opcodes de datos (READ, READ_FREE, WRITE, DELETE) antes de explorar las operaciones más avanzadas de célula (DIFERIR, CRUZAR).
