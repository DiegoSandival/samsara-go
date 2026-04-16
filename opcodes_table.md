# Tabla de Opcodes del Protocolo

Esta tabla resume todos los opcodes definidos en el protocolo Samsara.

## Contrato del Handler Central

El handler central consume y produce bytes crudos, independiente del transporte.

### Request Envelope (entrada al handler)

```
[Opcode: 1 byte] [ID: 16 bytes] [Payload: N bytes]
```

- `ID` se usa para correlacion request/response fuera del handler.
- El handler no interpreta semanticamente `ID`; solo parsea el envelope de entrada.

## Operaciones de Datos

### Lectura

| Opcode | Código | Descripción | Requiere Auth | Parámetros | Retorna New Index |
|--------|--------|-------------|---------------|-----------|-------------------|
| READ | `0x01` | Leer valor con autenticación | **Sí** | `dbName, key, cellIndex, secret` | **Sí** |
| READ_FREE | `0x02` | Leer valor público (sin auth) | **No** | `dbName, key` | **No** |

### Escritura y Eliminación

| Opcode | Código | Descripción | Requiere Auth | Parámetros | Retorna New Index |
|--------|--------|-------------|---------------|-----------|-------------------|
| WRITE | `0x03` | Escribir/actualizar valor | **Sí** | `dbName, key, value, cellIndex, secret` | **Sí** |
| DELETE | `0x04` | Eliminar valor | **Sí** | `dbName, key, cellIndex, secret` | **Sí** |

### Gestión de Bases de Datos

| Opcode | Código | Descripción | Requiere Auth | Parámetros | Retorna New Index |
|--------|--------|-------------|---------------|-----------|-------------------|
| CREATE_DB | `0x08` | Crear una base de datos | **NO** | `dbName, cellIndex, secret` | **No** |
| DELETE_DB | `0x09` | Eliminar una base de datos | **NO** | `dbName, cellIndex, secret` | **No** |

## Operaciones de Célula (Genética)

| Opcode | Código | Descripción | Requiere Auth | Parámetros | Retorna New Index |
|--------|--------|-------------|---------------|-----------|-------------------|
| READ_CELL | `0x05` | Leer información de célula | **Sí** | `dbName, cellIndex, secret` | **Sí** |
| DIFERIR | `0x06` | Reproducción: crear hijo de padre | **Sí** | `dbName, cellIndex, parentSecret, childSecret, childSalt, childGenome, x, y, z` | **Sí** |
| CRUZAR | `0x07` | Fusión: combinar dos padres | **Sí** | `dbName, cellIndexA, secretA, cellIndexB, secretB, childSecret, childSalt, x, y, z` | **Sí** |

## Detalles de Estructura Binaria

### READ (0x01)
```
[Opcode: 0x01] [DB Name Length: 4 bytes] [DB Name: N bytes] [Key Length: 4 bytes] [Key: N bytes] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: M bytes]
```

### READ_FREE (0x02)
```
[Opcode: 0x02] [DB Name Length: 4 bytes] [DB Name: N bytes] [Key Length: 4 bytes] [Key: N bytes]
```

### WRITE (0x03)
```
[Opcode: 0x03] [DB Name Length: 4 bytes] [DB Name: N bytes] [Key Length: 4 bytes] [Key: N bytes] [Value Length: 4 bytes] [Value: M bytes] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: P bytes]
```

### DELETE (0x04)
```
[Opcode: 0x04] [DB Name Length: 4 bytes] [DB Name: N bytes] [Key Length: 4 bytes] [Key: N bytes] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: M bytes]
```

### CREATE_DB (0x08)
```
[Opcode: 0x08] [DB Name Length: 4 bytes] [DB Name: N bytes] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: M bytes]
```

### DELETE_DB (0x09)
```
[Opcode: 0x09] [DB Name Length: 4 bytes] [DB Name: N bytes] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: M bytes]
```

### READ_CELL (0x05)
```
[Opcode: 0x05] [DB Name Length: 4 bytes] [DB Name: N bytes] [CellIndex: 4 bytes] [Secret Length: 4 bytes] [Secret: N bytes]
```

### DIFERIR (0x06)
```
[Opcode: 0x06]
[DB Name Length: 4 bytes] [DB Name: N bytes]
[CellIndex: 4 bytes]
[ParentSecret Length: 4 bytes] [ParentSecret: N bytes]
[ChildSecret Length: 4 bytes] [ChildSecret: M bytes]
[ChildSalt: 16 bytes]
[ChildGenome: 4 bytes]
[X: 4 bytes] [Y: 4 bytes] [Z: 4 bytes]
```

### CRUZAR (0x07)
```
[Opcode: 0x07]
[DB Name Length: 4 bytes] [DB Name: N bytes]
[CellIndexA: 4 bytes] [SecretA Length: 4 bytes] [SecretA: N bytes]
[CellIndexB: 4 bytes] [SecretB Length: 4 bytes] [SecretB: M bytes]
[ChildSecret Length: 4 bytes] [ChildSecret: P bytes]
[ChildSalt: 16 bytes]
[X: 4 bytes] [Y: 4 bytes] [Z: 4 bytes]
```

## Estados de Respuesta

Todos los opcodes retornan uno de estos estados, codificados en 1 byte:

| Estado | Descripción |
|--------|-------------|
| `ok` | Operación exitosa |
| `unauthorized` | Autenticación fallida o permisos insuficientes |
| `undefined` | El recurso no existe |
| `error_db` | Error en la base de datos |

| Código (byte) | Estado |
|---------------|--------|
| `0` | `ok` |
| `1` | `unauthorized` |
| `2` | `undefined` |
| `3` | `error_db` |

## Formato Binario de Response (implementado)

Todos los responses usan el mismo envelope:

```
[StatusCode: 1 byte] [Payload Length: 4 bytes LE] [Payload: N bytes]
```

Si `StatusCode` es `2` (`undefined`) o `3` (`error_db`) por error de parsing/opcode/db, el payload es:

```
[Error Length: 4 bytes LE] [Error Message: N bytes UTF-8]
```

### Payload por Opcode

### READ (0x01)

```
[Value Length: 4 bytes LE] [Value: N bytes]
[CellIndex: 4 bytes LE]
[NewCellIndex: 4 bytes LE]
[Flags: 1 byte]
```

`Flags`:
- bit 0 (`0x01`): `HasValue`
- bit 1 (`0x02`): `HasCellIndex`
- bit 2 (`0x04`): `HasNewCell`

### READ_FREE (0x02)

```
[Value Length: 4 bytes LE] [Value: N bytes]
[Flags: 1 byte]
```

`Flags`:
- bit 0 (`0x01`): `HasValue`

### WRITE (0x03)

```
[CellIndex: 4 bytes LE]
[NewCellIndex: 4 bytes LE]
[Flags: 1 byte]
```

`Flags`:
- bit 0 (`0x01`): `HasCellIndex`
- bit 1 (`0x02`): `HasNewCell`

### DELETE (0x04)

```
[CellIndex: 4 bytes LE]
[NewCellIndex: 4 bytes LE]
[Flags: 1 byte]
```

`Flags`:
- bit 0 (`0x01`): `HasCellIndex`
- bit 1 (`0x02`): `HasNewCell`

### READ_CELL (0x05)

```
[Hash: 32 bytes]
[Salt: 16 bytes]
[Genome: 4 bytes LE]
[X: 4 bytes LE]
[Y: 4 bytes LE]
[Z: 4 bytes LE]
[CellIndex: 4 bytes LE]
[Flags: 1 byte]
```

Tamano total fijo: 69 bytes.

`Flags`:
- bit 0 (`0x01`): `HasCell`
- bit 1 (`0x02`): `HasCellIndex`

### DIFERIR (0x06)

```
[CellIndex: 4 bytes LE]
[DeferredIndex: 4 bytes LE]
[NewCellIndex: 4 bytes LE]
[Flags: 1 byte]
```

`Flags`:
- bit 0 (`0x01`): `HasCellIndex`
- bit 1 (`0x02`): `HasDeferred`
- bit 2 (`0x04`): `HasNewCell`

### CRUZAR (0x07)

```
[CellIndexA: 4 bytes LE]
[CellIndexB: 4 bytes LE]
[ChildIndex: 4 bytes LE]
[NewCellIndexA: 4 bytes LE]
[NewCellIndexB: 4 bytes LE]
[Flags: 1 byte]
```

`Flags`:
- bit 0 (`0x01`): `HasCellIndexA`
- bit 1 (`0x02`): `HasCellIndexB`
- bit 2 (`0x04`): `HasChild`
- bit 3 (`0x08`): `HasNewCellA`
- bit 4 (`0x10`): `HasNewCellB`

### CREATE_DB (0x08)

Response `ok`:

```
[DB Name Length: 4 bytes LE] [DB Name: N bytes]
```

### DELETE_DB (0x09)

Response `ok`:

```
[DB Name Length: 4 bytes LE] [DB Name: N bytes]
```

## Nota de Flujo Manual

No existe DB por defecto. El flujo esperado es manual:

1. Enviar `CREATE_DB`.
2. Operar (`READ`, `WRITE`, `DELETE`, `READ_CELL`, `DIFERIR`, `CRUZAR`) sobre ese `dbName`.
3. Enviar `DELETE_DB` cuando corresponda.

## Sistema de Permisos (Genoma)

El genoma de una célula es un `uint32` con banderas de permisos:

| Bit | Código | Flag | Descripción |
|-----|--------|------|-------------|
| 0 | 0x01 | `LeerSelf` | Leer datos propios |
| 1 | 0x02 | `LeerAny` | Leer datos de otros |
| 2 | 0x04 | `LeerLibre` | Hacer datos legibles públicamente |
| 3 | 0x08 | `EscribirSelf` | Escribir datos propios |
| 4 | 0x10 | `EscribirAny` | Escribir datos de otros |
| 5 | 0x20 | `BorrarSelf` | Borrar datos propios |
| 6 | 0x40 | `BorrarAny` | Borrar datos de otros |
| 7 | 0x80 | `Diferir` | Crear células hijo |
| 8 | 0x100 | `Cruzar` | Fusionar con otra célula |

## Referencias

- [Documentación completa de READ](opcodes/READ.md)
- [Documentación completa de READ_FREE](opcodes/READ_FREE.md)
- [Documentación completa de WRITE](opcodes/WRITE.md)
- [Documentación completa de DELETE](opcodes/DELETE.md)
- [Documentación completa de READ_CELL](opcodes/READ_CELL.md)
- [Documentación completa de DIFERIR](opcodes/DIFERIR.md)
- [Documentación completa de CRUZAR](opcodes/CRUZAR.md)
- [Sistema de Permisos](opcodes/PERMISOS.md)