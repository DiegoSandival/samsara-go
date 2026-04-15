# Guía de Opcodes de Samsara

Esta carpeta contiene la documentación completa de todos los opcodes disponibles en la librería Samsara, un sistema de almacenamiento basado en células con autenticación y control de permisos.

## ¿Qué es un Opcode?

Un opcode es un identificador de operación que especifica qué acción debe realizar el sistema Samsara. Cada opcode:

- Tiene un identificador numénico (1 byte)
- Define una estructura binaria específica para sus parámetros
- Retorna un resultado con un estado específico
- Puede o no requerir autenticación

## Estructura de Célula

Las células son el componente fundamental de Samsara:

```go
type Celula struct {
    Hash   [32]byte   // Hash BLAKE3 de salt + secret
    Salt   [16]byte   // Salt aleatorio
    Genoma uint32     // Permisos de la célula
    X, Y, Z uint32    // Coordenadas en el espacio
}
```

## Flujo Típico de Operación

1. **Crear una célula** - Generar salt y secret, crear célula con permisos
2. **Autenticar** - Usar cell index + secret para validar operaciones
3. **Realizar operación** - Ejecutar READ, WRITE, DELETE, etc.
4. **Refresh** - En operaciones autenticadas, se retorna un nuevo cell index

## Categorías de Opcodes

### 🔐 Operaciones de Datos (Require Autenticación o Acceso)

- **READ** - Leer datos privados o propios
- **READ_FREE** - Leer datos públicos (sin autenticación)
- **WRITE** - Escribir/actualizar datos
- **DELETE** - Eliminar datos

### 🧬 Operaciones de Célula (Genética)

- **READ_CELL** - Leer información de una célula
- **DIFERIR** - Reproducción: crear hijo de una célula padre
- **CRUZAR** - Fusión: combinar dos células padre en una célula hijo

## Seguridad

### Autenticación

```
Validación:
1. Obtener célula por index
2. Derivar hash con: BLAKE3(salt + secret)
3. Comparar con hash almacenado
4. Si exito → Operación autorizada
```

### Permisos (Genoma)

Cada operación en datos valida permisos:

- Si **ownerIndex == cellIndex** → Requiere flag `Self`
- Si **ownerIndex != cellIndex** → Requiere flag `Any`
- Para operaciones **públicas** → Propietario debe tener flag `Libre`

## Ejemplo de Uso

```go
package main

import "github.com/usuario/samsara-go"

func main() {
    // Crear store
    store, _ := samsara.New("./data", 1000)
    defer store.Close()

    // Crear célula
    salt := [16]byte{/* datos */}
    secret := []byte("mi_secreto") 
    cell := samsara.NewCellWithSecret(salt, secret, 0xFF, 0, 0, 0)
    
    // Escribir dato
    result := store.Write("mi_clave", []byte("valor"), cellIndex, secret)
    
    // Leer dato
    readResult := store.Read("mi_clave", cellIndex, secret)
    
    // Leer público
    freeResult := store.ReadFree("dato_publico")
}
```

## Archivos en Esta Carpeta

- **INDEX.md** - Tabla resumen de todos los opcodes
- **README.md** - Este archivo
- **[OPCODE_NAME].md** - Documentación detallada de cada opcode

## Tipos de Respuesta

Cada opcode documenta sus posibles respuestas, incluyendo:

- Status (ok, unauthorized, undefined, error_db)
- Datos retornados (si aplica)
- Índices de célula (actuales o nuevos)
- Flags booleanos indicando presencia de datos

## Referencia Rápida

| Opcode | Función | Privado | Returns Cell |
|--------|---------|---------|--------------|
| READ | Leer dato | Sí | Nuevo índice |
| READ_FREE | Leer público | No | - |
| WRITE | Escribir | Sí | Nuevo índice |
| DELETE | Borrar | Sí | Nuevo índice |
| READ_CELL | Leer célula | Sí | Índice + célula |
| DIFERIR | Reproducir | Sí | Nuevo padre + hijo |
| CRUZAR | Fusionar | Sí | Nuevo A + Nuevo B + hijo |

## Para Comenzar

1. Lee [INDEX.md](INDEX.md) para una visión general
2. Consulta [READ.md](READ.md) y [WRITE.md](WRITE.md) para operaciones básicas
3. Explora [DIFERIR.md](DIFERIR.md) y [CRUZAR.md](CRUZAR.md) para operaciones avanzadas
4. Implementa según tus necesidades específicas
