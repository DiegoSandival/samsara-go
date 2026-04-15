# Tabla Completa de Permisos (Genoma)

## Resumen de Flags

| Bit | Valor | Flag | Descripción |
|-----|-------|------|-------------|
| 0 | 0x01 | `LeerSelf` | Leer datos cuyo propietario eres tú |
| 1 | 0x02 | `LeerAny` | Leer datos cuyo propietario es otro |
| 2 | 0x04 | `LeerLibre` | Hacer tus datos legibles públicamente |
| 3 | 0x08 | `EscribirSelf` | Escribir datos de tu propiedad |
| 4 | 0x10 | `EscribirAny` | Escribir datos de otros |
| 5 | 0x20 | `BorrarSelf` | Borrar datos de tu propiedad |
| 6 | 0x40 | `BorrarAny` | Borrar datos de otros |
| 7 | 0x80 | `Diferir` | Crear células hijo (reproducción) |
| 8 | 0x100 | `Fucionar` | Fusionar con otra célula |

## Operaciones y Permisos Requeridos

### READ (0x01) - Leer Dato

Para leer un dato, la célula necesita:

- **Si el dato es propio** (propietario == cellIndex):
  - Flag: `LeerSelf` (0x01) ✓
  - Si no tiene: ✗ Rechazado

- **Si el dato es ajeno** (propietario != cellIndex):
  - Flag: `LeerAny` (0x02) ✓
  - Si no tiene: ✗ Rechazado

### READ_FREE (0x02) - Leer Público

Para que un dato sea legible públicamente:

- **Propietario del dato** necesita:
  - Flag: `LeerLibre` (0x04) ✓
  - Si no tiene: ✗ Rechazado (status: unauthorized)

### WRITE (0x03) - Crear o Actualizar

**Crear nuevo dato:**
- Solo se requiere autenticación válida
- El nuevo dato tendrá como propietario la célula actual

**Actualizar dato existente:**

- **Si el dato es propio** (propietario == cellIndex):
  - Flag: `EscribirSelf` (0x08) ✓
  - Si no tiene: ✗ Rechazado

- **Si el dato es ajeno** (propietario != cellIndex):
  - Flag: `EscribirAny` (0x10) ✓
  - Si no tiene: ✗ Rechazado

### DELETE (0x04) - Borrar Dato

Para borrar un dato:

- **Si el dato es propio** (propietario == cellIndex):
  - Flag: `BorrarSelf` (0x20) ✓
  - Si no tiene: ✗ Rechazado

- **Si el dato es ajeno** (propietario != cellIndex):
  - Flag: `BorrarAny` (0x40) ✓
  - Si no tiene: ✗ Rechazado

### DIFERIR (0x06) - Reproducción Asexual

Para crear una célula hijo:

- **Célula padre** necesita:
  - Flag: `Diferir` (0x80) ✓
  - Si no tiene: ✗ Rechazado

- **Genoma hijo** debe cumplir:
  - `(childGenoma & parentGenoma) == childGenoma`
  - No puede tener más permisos que el padre

### CRUZAR (0x07) - Reproducción Sexual

Para fusionar dos células:

- **Padre A** necesita:
  - Flag: `Fucionar` (0x100) ✓
  - Si no tiene: ✗ Rechazado

- **Padre B** necesita:
  - Flag: `Fucionar` (0x100) ✓
  - Si no tiene: ✗ Rechazado

- **Genoma hijo** será:
  - `childGenoma = genomeA | genomeB` (unión)

## Combinaciones de Permisos Comunes

### Nivel 1: Lectura Básica (0x01)

```go
const ReadOnly = 0x01 // LeerSelf

// Acciones permitidas:
// ✓ Leer datos propios
// ✗ Leer datos de otros
// ✗ Escribir
// ✗ Borrar
// ✗ Reproducirse
```

### Nivel 2: Lectura Completa (0x03)

```go
const ReadAll = 0x01 | 0x02 // LeerSelf | LeerAny = 0x03

// Acciones permitidas:
// ✓ Leer datos propios
// ✓ Leer datos de otros
// ✗ Escribir
// ✗ Borrar
// ✗ Reproducirse
```

### Nivel 3: Lectura + Escritura Propia (0x09)

```go
const UserMode = 0x01 | 0x08 // LeerSelf | EscribirSelf = 0x09

// Acciones permitidas:
// ✓ Leer datos propios
// ✓ Escribir datos propios
// ✗ Leer/escribir de otros
// ✗ Borrar
// ✗ Reproducirse
```

### Nivel 4: Usuario Completo (0x3F)

```go
const FullUser = 0x01 | 0x02 | 0x04 | 0x08 | 0x10 | 0x20 = 0x3F

// Acciones permitidas:
// ✓ Leer datos propios y ajenos
// ✓ Escribir datos propios y ajenos
// ✓ Borrar datos propios
// ✗ Borrar datos de otros
// ✗ Reproducirse
```

### Nivel 5: Moderador (0x7F)

```go
const Moderator = 0x7F // Todo excepto Diferir

// Acciones permitidas:
// ✓ CRUD completo (propios y ajenos)
// ✓ Datos públicos
// ✗ Reproducirse (DIFERIR)
// ✓ Fusionarse (CRUZAR)
```

### Nivel 6: Administrador (0xFE)

```go
const Admin = 0xFE // Todo excepto Fucionar

// Acciones permitidas:
// ✓ CRUD completo
// ✓ Crear usuarios (DIFERIR)
// ✓ Datos públicos
// ✗ Fusionarse (CRUZAR)
```

### Nivel 7: SuperAdmin (0xFF)

```go
const SuperAdmin = 0xFF // Todos los permisos

// Acciones permitidas:
// ✓ CRUD completo
// ✓ Crear usuarios (DIFERIR)
// ✓ Fusionar células (CRUZAR)
// ✓ Hacer datos públicos
```

### Nivel 0: Solo Público (0x04)

```go
const PublicOnly = 0x04 // LeerLibre

// Acciones permitidas:
// ✓ Sus datos pueden leerse públicamente
// ✗ No puede leer nada
// ✗ No puede escribir
// ✗ No puede borrar
// ✗ No puede reproducirse
```

## Matriz de Decisión de Permisos

### Para Operación: READ

```
┌─ ¿Dato es propio?
│  ├─ SÍ → ¿Tiene LeerSelf (0x01)? → PERMITIR
│  └─ NO → ¿Tiene LeerAny (0x02)? → PERMITIR
                                     RECHAZAR
```

### Para Operación: WRITE (actualizar)

```
┌─ ¿Dato es propio?
│  ├─ SÍ → ¿Tiene EscribirSelf (0x08)? → PERMITIR
│  └─ NO → ¿Tiene EscribirAny (0x10)? → PERMITIR
                                          RECHAZAR
```

### Para Operación: DELETE

```
┌─ ¿Dato es propio?
│  ├─ SÍ → ¿Tiene BorrarSelf (0x20)? → PERMITIR
│  └─ NO → ¿Tiene BorrarAny (0x40)? → PERMITIR
                                        RECHAZAR
```

## Cálculo de Permisos en Código

### Verificar Un Permiso

```go
// Verificar si tiene LeerSelf
if cell.Genoma&0x01 != 0 {
    // Tiene el permiso
}

// O usando constantes
if cell.Genoma&ouroboros.LeerSelf != 0 {
    // Tiene LeerSelf
}
```

### Verificar Múltiples Permisos

```go
// Verificar si tiene CUALQUIERA de (Leer + Escribir)
if cell.Genoma&(ouroboros.LeerSelf|ouroboros.EscribirSelf) != 0 {
    // Tiene al menos uno
}

// Verificar si tiene TODOS de (Leer + Escribir)
if (cell.Genoma & (ouroboros.LeerSelf|ouroboros.EscribirSelf)) == (ouroboros.LeerSelf|ouroboros.EscribirSelf) {
    // Tiene ambos
}
```

### Agregar Permiso

```go
// Habilitar LeerSelf
cell.Genoma |= ouroboros.LeerSelf // Suma el flag

// O agregar múltiples
cell.Genoma |= ouroboros.LeerSelf | ouroboros.EscribirSelf
```

### Remover Permiso

```go
// Deshabilitar BorrarAny
cell.Genoma &= ^ouroboros.BorrarAny // Resta el flag
```

### Crear Genoma Específico

```go
// Lectura y Escritura solamente
genome := ouroboros.LeerSelf | ouroboros.EscribirSelf // 0x09

// Todo excepto reproducción
genome := 0xFF & ^(ouroboros.Diferir | ouroboros.Fucionar) // 0x7E

// Copiar permisos de otro
genomeA := 0xFF
genomeB := 0x03
combinado := genomeA | genomeB // Unión (CRUZAR)
limitado  := genomeA & genomeB // Intersección
```

## Tablas de Referencia Rápida

### CRUD: Quien puede qué

| Escenario | Requiere | Flag |
|-----------|----------|------|
| Leer dato propio | Auth | LeerSelf (0x01) |
| Leer dato ajeno | Auth | LeerAny (0x02) |
| Escribir dato propio (crear) | Auth | - |
| Escribir dato propio (actualizar) | Auth | EscribirSelf (0x08) |
| Escribir dato ajeno | Auth | EscribirAny (0x10) |
| Borrar dato propio | Auth | BorrarSelf (0x20) |
| Borrar dato ajeno | Auth | BorrarAny (0x40) |

### Operaciones Genéticas: Quien puede qué

| Operación | Requiere | Flag padre |
|-----------|----------|-----------|
| DIFERIR (reproducir) | Auth padre | Diferir (0x80) |
| CRUZAR (fusionar) | Auth A + B | Fucionar (0x100) en ambos |

### Datos Públicos

| Acción | Requerido | Flag |
|--------|-----------|------|
| Hacer dato público | Auth autor | LeerLibre (0x04) del autor |
| Leer dato público | Ninguno | - |

## Ejemplos Reales

### Estructura de Permisos: Red Social

```
SuperAdmin:    0xFF  - Control total
Admin:         0xFE  - Todo excepto fusión
Moderador:     0x7F  - CRUD + revisar contenido
Usuario:       0x09  - CRUD propios
Guest (anon):  0x01  - Lectura pública
```

### Estructura: Sistema Empresarial

```
CEO:           0xFF  - Todo
Director:      0xFE  - Todo excepto fusión
Gerente:       0x3F  - CRUD + supervisar
Empleado:      0x09  - Lectura escritura propios
Cliente:       0x05  - Lectura propia + pública
Visitante:     0x04  - Solo lectura pública
```

### Estructura: Sistema de Archivos

```
Propietario:   0xFF   - Lectura, escritura, borrado
Colaborador:   0x7F   - Escritura compartida
Lector:        0x03   - Lectura únicamente
Público:       0x04   - Lectura pública si permitido
```

## Herencia en DIFERIR

```
Padre:        0xFF (Todos)
   ├─ Hijo directo:     0x7F (Padre - Fucionar)
   │  └─ Nieto:         0x3F (Abuelo - Diferir - Fucionar)
   └─ Hijo limitado:    0x09 (Solo CRUD propios)
```

## Herencia en CRUZAR

```
Padre A: 0x01 (LeerSelf)
Padre B: 0x02 (LeerAny)
─────────────────────────
Hijo:    0x03 (Leer todo)

Padre A: 0x08 (EscribirSelf)
Padre B: 0x20 (BorrarSelf)
─────────────────────────
Hijo:    0x28 (Escribir y borrar propios)
```

## Validación de Herencia en Código

```go
func validateChildGenome(parentGenoma, childGenoma uint32) bool {
    // En DIFERIR: hijo ⊆ padre
    return (childGenoma & parentGenoma) == childGenoma
}

func combineGenome(genomeA, genomeB uint32) uint32 {
    // En CRUZAR: hijo = A ∪ B
    return genomeA | genomeB
}
```

## Flags de Guard para Operaciones Peligrosas

```go
// No permitir BorrarAny a menos que sea admin
if requestedGenome&ouroboros.BorrarAny != 0 && !isAdmin {
    return fmt.Errorf("unauthorized: delete any not allowed")
}

// No permitir reproducción a usuarios normales
if requestedGenome&(ouroboros.Diferir|ouroboros.Fucionar) != 0 && !isAdmin {
    return fmt.Errorf("unauthorized: reproduction not allowed")
}
```

## Conversión Entre Niveles

```go
// Escalar permisos (downgrade prevención)
func escalatePermissions(current, requested uint32) (uint32, error) {
    // Solo permitir agregar, nunca quitar en DIFERIR
    if (requested & current) != requested {
        return 0, fmt.Errorf("cannot request permissions beyond parent")
    }
    return requested, nil
}

// Revoke específicos permisos
func revokePermission(current uint32, toRevoke uint32) uint32 {
    return current &^ toRevoke // XOR para desactivar bits
}
```

