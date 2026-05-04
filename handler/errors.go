package samsara

import protocol "github.com/DiegoSandival/samsara-go/protocol"

func errorFromPayload(parser *protocol.ProtocolParser, payload []byte, code protocol.ErrorCode, info string) []byte {
	return parser.ErrorResultBytes(parser.RequestID(payload), uint32(code), []byte(info))
}

func errorWithID(parser *protocol.ProtocolParser, id []byte, code protocol.ErrorCode, info string) []byte {
	return parser.ErrorResultBytes(id, uint32(code), []byte(info))
}

func requestParseError(parser *protocol.ProtocolParser, payload []byte, opcode byte, err error, info string) []byte {
	code := protocol.ParseRequestErrorCode(opcode, err)
	return parser.ErrorResultBytes(parser.RequestID(payload), uint32(code), []byte(info))
}
