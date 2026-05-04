package protocol

import "errors"

type ErrorCode uint32

const (
	ErrorCodeOpcodeFrameTooShort ErrorCode = 1100
	ErrorCodeUnknownOpcode       ErrorCode = 1101
)

const (
	ErrorCodeCreateDBReqTooShort   ErrorCode = 1200
	ErrorCodeCreateDBReqIncomplete ErrorCode = 1201
	ErrorCodeDeleteDBReqTooShort   ErrorCode = 1202
	ErrorCodeDeleteDBReqIncomplete ErrorCode = 1203
	ErrorCodeWriteReqTooShort      ErrorCode = 1204
	ErrorCodeWriteReqIncomplete    ErrorCode = 1205
	ErrorCodeReadReqTooShort       ErrorCode = 1206
	ErrorCodeReadReqIncomplete     ErrorCode = 1207
	ErrorCodeReadFreeReqTooShort   ErrorCode = 1208
	ErrorCodeReadFreeReqIncomplete ErrorCode = 1209
	ErrorCodeDeleteReqTooShort     ErrorCode = 1210
	ErrorCodeDeleteReqIncomplete   ErrorCode = 1211
	ErrorCodeReadCellReqTooShort   ErrorCode = 1212
	ErrorCodeReadCellReqIncomplete ErrorCode = 1213
	ErrorCodeDiferirReqTooShort    ErrorCode = 1214
	ErrorCodeDiferirChildSaltShort ErrorCode = 1215
	ErrorCodeDiferirReqIncomplete  ErrorCode = 1216
	ErrorCodeCruzarReqTooShort     ErrorCode = 1217
	ErrorCodeCruzarReqIncomplete   ErrorCode = 1218
)

const (
	ErrorCodeRequestIDGenerationFailed ErrorCode = 1300
	ErrorCodeChildSaltGenerationFailed ErrorCode = 1301
)

const (
	ErrorCodeDatabaseAlreadyExists  ErrorCode = 2100
	ErrorCodeDatabaseNotFound       ErrorCode = 2101
	ErrorCodeAuthenticationFailed   ErrorCode = 2102
	ErrorCodeMembraneNotFound       ErrorCode = 2103
	ErrorCodePermissionDenied       ErrorCode = 2104
	ErrorCodeInvalidGenomeFlags     ErrorCode = 2105
	ErrorCodeChildGenomeEscalation  ErrorCode = 2106
	ErrorCodeMergeCapabilityMissing ErrorCode = 2107
)

const (
	ErrorCodeStoreCreateFailed     ErrorCode = 3100
	ErrorCodeStoreOpenFailed       ErrorCode = 3101
	ErrorCodeStoreDestroyFailed    ErrorCode = 3102
	ErrorCodeMembraneReadFailed    ErrorCode = 3103
	ErrorCodeMembraneWriteFailed   ErrorCode = 3104
	ErrorCodeMembraneDeleteFailed  ErrorCode = 3105
	ErrorCodeInitialCellAppendFail ErrorCode = 3106
	ErrorCodeCellAppendFailed      ErrorCode = 3107
	ErrorCodeCellRefreshFailed     ErrorCode = 3108
	ErrorCodeRandomSourceFailed    ErrorCode = 3109
	ErrorCodeConfigLoadFailed      ErrorCode = 3110
	ErrorCodeBaseDirCreateFailed   ErrorCode = 3111
	ErrorCodeStoreDirReadFailed    ErrorCode = 3112
	ErrorCodeStartupStoreOpenFail  ErrorCode = 3113
)

const (
	ErrorCodeInternalUnexpected ErrorCode = 9000
)

var (
	ErrMessageTooShort        = errors.New("message too short")
	ErrMessageIncomplete      = errors.New("message incomplete")
	ErrChildSaltTooShort      = errors.New("message too short for child salt")
	ErrRandomSourceFailure    = errors.New("random source failure")
	ErrUnknownProtocolFailure = errors.New("unknown protocol failure")
)

type ErrorDefinition struct {
	Code        ErrorCode
	Name        string
	Layer       string
	Description string
	Operations  string
}

var CentralErrorCatalog = []ErrorDefinition{
	{Code: ErrorCodeOpcodeFrameTooShort, Name: "OPCODE_FRAME_TOO_SHORT", Layer: "routing", Description: "El mensaje no tiene los 4 bytes necesarios para leer el opcode.", Operations: "global"},
	{Code: ErrorCodeUnknownOpcode, Name: "UNKNOWN_OPCODE", Layer: "routing", Description: "El opcode no pertenece al rango soportado por el servidor.", Operations: "global"},
	{Code: ErrorCodeCreateDBReqTooShort, Name: "CREATE_DB_REQ_TOO_SHORT", Layer: "parser", Description: "Cabecera fija incompleta en CREATE_DB.", Operations: "create_db"},
	{Code: ErrorCodeCreateDBReqIncomplete, Name: "CREATE_DB_REQ_INCOMPLETE", Layer: "parser", Description: "Payload variable incompleto en CREATE_DB.", Operations: "create_db"},
	{Code: ErrorCodeDeleteDBReqTooShort, Name: "DELETE_DB_REQ_TOO_SHORT", Layer: "parser", Description: "Cabecera fija incompleta en DELETE_DB.", Operations: "delete_db"},
	{Code: ErrorCodeDeleteDBReqIncomplete, Name: "DELETE_DB_REQ_INCOMPLETE", Layer: "parser", Description: "Payload variable incompleto en DELETE_DB.", Operations: "delete_db"},
	{Code: ErrorCodeWriteReqTooShort, Name: "WRITE_REQ_TOO_SHORT", Layer: "parser", Description: "Cabecera fija incompleta en WRITE.", Operations: "write"},
	{Code: ErrorCodeWriteReqIncomplete, Name: "WRITE_REQ_INCOMPLETE", Layer: "parser", Description: "Payload variable incompleto en WRITE.", Operations: "write"},
	{Code: ErrorCodeReadReqTooShort, Name: "READ_REQ_TOO_SHORT", Layer: "parser", Description: "Cabecera fija incompleta en READ.", Operations: "read"},
	{Code: ErrorCodeReadReqIncomplete, Name: "READ_REQ_INCOMPLETE", Layer: "parser", Description: "Payload variable incompleto en READ.", Operations: "read"},
	{Code: ErrorCodeReadFreeReqTooShort, Name: "READ_FREE_REQ_TOO_SHORT", Layer: "parser", Description: "Cabecera fija incompleta en READ_FREE.", Operations: "read_free"},
	{Code: ErrorCodeReadFreeReqIncomplete, Name: "READ_FREE_REQ_INCOMPLETE", Layer: "parser", Description: "Payload variable incompleto en READ_FREE.", Operations: "read_free"},
	{Code: ErrorCodeDeleteReqTooShort, Name: "DELETE_REQ_TOO_SHORT", Layer: "parser", Description: "Cabecera fija incompleta en DELETE.", Operations: "delete"},
	{Code: ErrorCodeDeleteReqIncomplete, Name: "DELETE_REQ_INCOMPLETE", Layer: "parser", Description: "Payload variable incompleto en DELETE.", Operations: "delete"},
	{Code: ErrorCodeReadCellReqTooShort, Name: "READ_CELL_REQ_TOO_SHORT", Layer: "parser", Description: "Cabecera fija incompleta en READ_CELL.", Operations: "read_cell"},
	{Code: ErrorCodeReadCellReqIncomplete, Name: "READ_CELL_REQ_INCOMPLETE", Layer: "parser", Description: "Payload variable incompleto en READ_CELL.", Operations: "read_cell"},
	{Code: ErrorCodeDiferirReqTooShort, Name: "DIFERIR_REQ_TOO_SHORT", Layer: "parser", Description: "Cabecera fija incompleta en DIFERIR.", Operations: "diferir"},
	{Code: ErrorCodeDiferirChildSaltShort, Name: "DIFERIR_CHILD_SALT_TOO_SHORT", Layer: "parser", Description: "El bloque ChildSalt no tiene 16 bytes en DIFERIR.", Operations: "diferir"},
	{Code: ErrorCodeDiferirReqIncomplete, Name: "DIFERIR_REQ_INCOMPLETE", Layer: "parser", Description: "Payload variable incompleto en DIFERIR.", Operations: "diferir"},
	{Code: ErrorCodeCruzarReqTooShort, Name: "CRUZAR_REQ_TOO_SHORT", Layer: "parser", Description: "Cabecera fija incompleta en CRUZAR.", Operations: "cruzar"},
	{Code: ErrorCodeCruzarReqIncomplete, Name: "CRUZAR_REQ_INCOMPLETE", Layer: "parser", Description: "Payload variable incompleto en CRUZAR.", Operations: "cruzar"},
	{Code: ErrorCodeRequestIDGenerationFailed, Name: "REQUEST_ID_GENERATION_FAILED", Layer: "builder", Description: "Fallo al generar el ID aleatorio de un request saliente.", Operations: "create_db,delete_db,write,read,read_free,read_cell,diferir,cruzar"},
	{Code: ErrorCodeChildSaltGenerationFailed, Name: "CHILD_SALT_GENERATION_FAILED", Layer: "builder", Description: "Fallo al generar ChildSalt en DIFERIR saliente.", Operations: "diferir"},
	{Code: ErrorCodeDatabaseAlreadyExists, Name: "DATABASE_ALREADY_EXISTS", Layer: "business", Description: "La base de datos ya existe en el registro del handler.", Operations: "create_db"},
	{Code: ErrorCodeDatabaseNotFound, Name: "DATABASE_NOT_FOUND", Layer: "business", Description: "La base de datos solicitada no está registrada o no pudo abrirse.", Operations: "delete_db,read,write,read_free,delete,read_cell,diferir,cruzar"},
	{Code: ErrorCodeAuthenticationFailed, Name: "AUTHENTICATION_FAILED", Layer: "business", Description: "La celda o secreto no pudieron autenticarse.", Operations: "delete_db,read,write,delete,read_cell,diferir,cruzar"},
	{Code: ErrorCodeMembraneNotFound, Name: "MEMBRANE_NOT_FOUND", Layer: "business", Description: "La key solicitada no existe en el bucket de membranas.", Operations: "read,read_free,delete"},
	{Code: ErrorCodePermissionDenied, Name: "PERMISSION_DENIED", Layer: "business", Description: "El genoma activo no tiene permisos para la operación.", Operations: "read,write,delete"},
	{Code: ErrorCodeInvalidGenomeFlags, Name: "INVALID_GENOME_FLAGS", Layer: "business", Description: "La operación CREATE_DB recibe flags de genoma incompatibles con la raíz.", Operations: "create_db"},
	{Code: ErrorCodeChildGenomeEscalation, Name: "CHILD_GENOME_ESCALATION", Layer: "business", Description: "DIFERIR intenta crear un hijo con permisos no heredados del padre.", Operations: "diferir"},
	{Code: ErrorCodeMergeCapabilityMissing, Name: "MERGE_CAPABILITY_MISSING", Layer: "business", Description: "Alguna de las celdas no tiene el bit Merge para CRUZAR.", Operations: "cruzar"},
	{Code: ErrorCodeStoreCreateFailed, Name: "STORE_CREATE_FAILED", Layer: "storage", Description: "Falló la creación de la DB Ouroboros o Bolt para una nueva store.", Operations: "create_db"},
	{Code: ErrorCodeStoreOpenFailed, Name: "STORE_OPEN_FAILED", Layer: "storage", Description: "Falló la apertura de una store existente.", Operations: "startup"},
	{Code: ErrorCodeStoreDestroyFailed, Name: "STORE_DESTROY_FAILED", Layer: "storage", Description: "Falló el cierre o borrado físico de la store.", Operations: "delete_db"},
	{Code: ErrorCodeMembraneReadFailed, Name: "MEMBRANE_READ_FAILED", Layer: "storage", Description: "Falló la lectura o decodificación de una membrana.", Operations: "read,write,read_free,delete"},
	{Code: ErrorCodeMembraneWriteFailed, Name: "MEMBRANE_WRITE_FAILED", Layer: "storage", Description: "Falló la escritura de una membrana.", Operations: "write"},
	{Code: ErrorCodeMembraneDeleteFailed, Name: "MEMBRANE_DELETE_FAILED", Layer: "storage", Description: "Falló la eliminación de una membrana.", Operations: "delete"},
	{Code: ErrorCodeInitialCellAppendFail, Name: "INITIAL_CELL_APPEND_FAILED", Layer: "storage", Description: "Falló el append de la célula raíz al crear la base.", Operations: "create_db"},
	{Code: ErrorCodeCellAppendFailed, Name: "CELL_APPEND_FAILED", Layer: "storage", Description: "Falló el append de una nueva célula en DIFERIR o CRUZAR.", Operations: "diferir,cruzar"},
	{Code: ErrorCodeCellRefreshFailed, Name: "CELL_REFRESH_FAILED", Layer: "storage", Description: "Falló el refresh/migración de la célula activa.", Operations: "read,write"},
	{Code: ErrorCodeRandomSourceFailed, Name: "RANDOM_SOURCE_FAILED", Layer: "storage", Description: "Falló la fuente de entropía usada para ID o salt.", Operations: "create_db,diferir,builders"},
	{Code: ErrorCodeConfigLoadFailed, Name: "CONFIG_LOAD_FAILED", Layer: "startup", Description: "Falló la lectura del archivo .env.", Operations: "startup"},
	{Code: ErrorCodeBaseDirCreateFailed, Name: "BASE_DIR_CREATE_FAILED", Layer: "startup", Description: "Falló la creación del directorio base de datos.", Operations: "startup"},
	{Code: ErrorCodeStoreDirReadFailed, Name: "STORE_DIR_READ_FAILED", Layer: "startup", Description: "Falló el escaneo del directorio base de stores.", Operations: "startup"},
	{Code: ErrorCodeStartupStoreOpenFail, Name: "STARTUP_STORE_OPEN_FAILED", Layer: "startup", Description: "Una store detectada en disco no pudo abrirse durante el arranque.", Operations: "startup"},
	{Code: ErrorCodeInternalUnexpected, Name: "INTERNAL_UNEXPECTED", Layer: "internal", Description: "Fallo no clasificado o estado imposible detectado por el servidor.", Operations: "global"},
}

type ErrorResult struct {
	ID      []byte
	ErrorID uint32
	Info    []byte
}

func (p *ProtocolParser) ErrorResultBytes(id []byte, errorID uint32, info []byte) []byte {
	normalizedID := normalizeID(id)
	result := make([]byte, 16+4+len(info))
	copy(result[:16], normalizedID)
	putUint32(result[16:20], errorID)
	copy(result[20:], info)
	return result
}

func (p *ProtocolParser) ErrorResult(msg []byte) (ErrorResult, error) {
	var result ErrorResult
	if len(msg) < 20 {
		return result, ErrMessageTooShort
	}

	result.ID = normalizeID(msg[:16])
	result.ErrorID = readUint32(msg[16:20])
	result.Info = append([]byte(nil), msg[20:]...)
	return result, nil
}

func (p *ProtocolParser) RequestID(msg []byte) []byte {
	id := make([]byte, 16)
	if len(msg) <= 4 {
		return id
	}

	end := 20
	if len(msg) < end {
		end = len(msg)
	}
	copy(id, msg[4:end])
	return id
}

func ParseRequestErrorCode(opcode byte, err error) ErrorCode {
	switch opcode {
	case OpcodeCreateDB:
		if errors.Is(err, ErrMessageTooShort) {
			return ErrorCodeCreateDBReqTooShort
		}
		if errors.Is(err, ErrMessageIncomplete) {
			return ErrorCodeCreateDBReqIncomplete
		}
	case OpcodeDeleteDB:
		if errors.Is(err, ErrMessageTooShort) {
			return ErrorCodeDeleteDBReqTooShort
		}
		if errors.Is(err, ErrMessageIncomplete) {
			return ErrorCodeDeleteDBReqIncomplete
		}
	case OpcodeWrite:
		if errors.Is(err, ErrMessageTooShort) {
			return ErrorCodeWriteReqTooShort
		}
		if errors.Is(err, ErrMessageIncomplete) {
			return ErrorCodeWriteReqIncomplete
		}
	case OpcodeRead:
		if errors.Is(err, ErrMessageTooShort) {
			return ErrorCodeReadReqTooShort
		}
		if errors.Is(err, ErrMessageIncomplete) {
			return ErrorCodeReadReqIncomplete
		}
	case OpcodeReadFree:
		if errors.Is(err, ErrMessageTooShort) {
			return ErrorCodeReadFreeReqTooShort
		}
		if errors.Is(err, ErrMessageIncomplete) {
			return ErrorCodeReadFreeReqIncomplete
		}
	case OpcodeDelete:
		if errors.Is(err, ErrMessageTooShort) {
			return ErrorCodeDeleteReqTooShort
		}
		if errors.Is(err, ErrMessageIncomplete) {
			return ErrorCodeDeleteReqIncomplete
		}
	case OpcodeReadCell:
		if errors.Is(err, ErrMessageTooShort) {
			return ErrorCodeReadCellReqTooShort
		}
		if errors.Is(err, ErrMessageIncomplete) {
			return ErrorCodeReadCellReqIncomplete
		}
	case OpcodeDiferir:
		if errors.Is(err, ErrMessageTooShort) {
			return ErrorCodeDiferirReqTooShort
		}
		if errors.Is(err, ErrChildSaltTooShort) {
			return ErrorCodeDiferirChildSaltShort
		}
		if errors.Is(err, ErrMessageIncomplete) {
			return ErrorCodeDiferirReqIncomplete
		}
	case OpcodeCruzar:
		if errors.Is(err, ErrMessageTooShort) {
			return ErrorCodeCruzarReqTooShort
		}
		if errors.Is(err, ErrMessageIncomplete) {
			return ErrorCodeCruzarReqIncomplete
		}
	}

	return ErrorCodeInternalUnexpected
}

func normalizeID(id []byte) []byte {
	normalized := make([]byte, 16)
	copy(normalized, id)
	return normalized
}

func putUint32(dst []byte, value uint32) {
	dst[0] = byte(value >> 24)
	dst[1] = byte(value >> 16)
	dst[2] = byte(value >> 8)
	dst[3] = byte(value)
}

func readUint32(src []byte) uint32 {
	return uint32(src[0])<<24 | uint32(src[1])<<16 | uint32(src[2])<<8 | uint32(src[3])
}
