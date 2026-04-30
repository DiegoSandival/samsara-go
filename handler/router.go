// router.go o dispatcher.go
package samsara

import (
	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

// ProcessRequest es la función principal que tu servidor (TCP, HTTP, etc.) llamará.
func ProcessRequest(msg []byte, parser *protocol.ProtocolParser, handler *CentralHandler) []byte {
	// 1. Validar que el mensaje tenga al menos el tamaño del opcode
	if len(msg) < 1 {
		// Asumiendo que tu protocolo tiene una forma de devolver errores genéricos
		return []byte("error: mensaje muy corto")
	}

	// 3. Enrutar según el opcode
	switch parser.Opcode(msg) {
	case 0x20:
		return handler.CreateDB(parser, msg)
	case 0x21:
		return handler.DelDB(parser, msg)
	case 0x22:
		return handler.Write(parser, msg)
	case 0x23:
		return handler.Read(parser, msg)
	case 0x24:
		return handler.ReadFree(parser, msg)
	case 0x25:
		return handler.Delete(parser, msg)
	case 0x26:
		return handler.ReadCell(parser, msg)
	case 0x27:
		return handler.Diferir(parser, msg)
	case 0x28:
		return handler.Cruzar(parser, msg)

	default:
		// Opcode no soportado
		return []byte("error: opcode desconocido")
	}
}
