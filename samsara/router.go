// router.go o dispatcher.go
package main

import protocol "github.com/DiegoSandival/samsara-go/protocol"

// ProcessRequest es la función principal que tu servidor (TCP, HTTP, etc.) llamará.
func ProcessRequest(msg []byte, parser *protocol.ProtocolParser, handler *CentralHandler) []byte {
	// 1. Validar que el mensaje tenga al menos el tamaño del opcode
	if len(msg) < 1 {
		// Asumiendo que tu protocolo tiene una forma de devolver errores genéricos
		return []byte("error: mensaje muy corto")
	}

	// 2. Extraer el opcode (asumiendo que es el primer byte)
	// Ajusta los índices si tu opcode está en otra posición (ej. msg[0:4])
	opcode := msg[0]
	payload := msg[1:]

	// 3. Enrutar según el opcode
	switch opcode {
	case 0x00:
		return handler.CreateDB(parser, payload)
	// case 0x01:
	//     return handler.InsertData(parser, payload)
	default:
		// Opcode no soportado
		return []byte("error: opcode desconocido")
	}
}
