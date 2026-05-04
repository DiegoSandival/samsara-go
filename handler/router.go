// router.go o dispatcher.go
package samsara

import (
	"fmt"

	protocol "github.com/DiegoSandival/samsara-go/protocol"
)

// ProcessRequest es la función principal que tu servidor (TCP, HTTP, etc.) llamará.
func ProcessRequest(msg []byte, parser *protocol.ProtocolParser, handler *CentralHandler) []byte {
	// 1. Validar que el mensaje tenga al menos el tamaño del opcode
	if len(msg) < 4 {
		return parser.ErrorResultBytes(parser.RequestID(msg), uint32(protocol.ErrorCodeOpcodeFrameTooShort), []byte("router.opcode"))
	}

	// 3. Enrutar según el opcode
	switch parser.Opcode(msg) {
	case protocol.OpcodeCreateDB:
		return handler.CreateDB(parser, msg)
	case protocol.OpcodeDeleteDB:
		return handler.DelDB(parser, msg)
	case protocol.OpcodeWrite:
		return handler.Write(parser, msg)
	case protocol.OpcodeRead:
		return handler.Read(parser, msg)
	case protocol.OpcodeReadFree:
		return handler.ReadFree(parser, msg)
	case protocol.OpcodeDelete:
		return handler.Delete(parser, msg)
	case protocol.OpcodeReadCell:
		return handler.ReadCell(parser, msg)
	case protocol.OpcodeDiferir:
		return handler.Diferir(parser, msg)
	case protocol.OpcodeCruzar:
		return handler.Cruzar(parser, msg)

	default:
		info := fmt.Sprintf("router.opcode=0x%02x", parser.Opcode(msg))
		return parser.ErrorResultBytes(parser.RequestID(msg), uint32(protocol.ErrorCodeUnknownOpcode), []byte(info))
	}
}
