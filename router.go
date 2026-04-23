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
	case 0x00:
		return handler.CreateDB(parser, msg)
	case 0x01:
		return handler.DelDB(parser, msg)
	case 0x02:
		return handler.Write(parser, msg)
	case 0x03:
		return handler.Read(parser, msg)
	case 0x04:
		return handler.ReadFree(parser, msg)
	case 0x05:
		return handler.Delete(parser, msg)

	default:
		// Opcode no soportado
		return []byte("error: opcode desconocido")
	}
}
