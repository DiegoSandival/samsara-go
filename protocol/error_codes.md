# Centralized Error Codes

## Envelope

All protocol-level errors now use the same binary envelope:

`[ID:16][ErrorID:4][Info:N]`

- `ID`: request ID normalized to 16 bytes.
- `ErrorID`: global centralized error code in big endian.
- `Info`: free-form ASCII context bytes. There is no explicit length field; the rest of the message is the info payload.

If the request does not contain a full ID, the parser copies the available bytes and pads the remainder with `0x00`.

## Ranges

| Range | Layer | Purpose |
| --- | --- | --- |
| 1100-1199 | routing | framing and opcode dispatch |
| 1200-1299 | parser | request parsing failures by opcode |
| 1300-1399 | builder | outgoing request generation blind spots |
| 2100-2199 | business | validation, auth and permission failures |
| 3100-3199 | storage/startup | filesystem, DB and persistence failures |
| 9000-9099 | internal | unexpected or unmapped failures |

## Mapping Table

| Code | Name | Layer | Trigger | Current surface |
| --- | --- | --- | --- | --- |
| 1100 | OPCODE_FRAME_TOO_SHORT | routing | `ProcessRequest` receives fewer than 4 bytes | emitted on wire |
| 1101 | UNKNOWN_OPCODE | routing | opcode outside `0x20-0x28` | emitted on wire |
| 1200 | CREATE_DB_REQ_TOO_SHORT | parser | fixed header shorter than 36 bytes in `CreateDBReq` | emitted on wire |
| 1201 | CREATE_DB_REQ_INCOMPLETE | parser | variable section shorter than declared lengths in `CreateDBReq` | emitted on wire |
| 1202 | DELETE_DB_REQ_TOO_SHORT | parser | fixed header shorter than 32 bytes in `DeleteDBReq` | emitted on wire |
| 1203 | DELETE_DB_REQ_INCOMPLETE | parser | variable section shorter than declared lengths in `DeleteDBReq` | emitted on wire |
| 1204 | WRITE_REQ_TOO_SHORT | parser | fixed header shorter than 40 bytes in `WriteReq` | emitted on wire |
| 1205 | WRITE_REQ_INCOMPLETE | parser | variable section shorter than declared lengths in `WriteReq` | emitted on wire |
| 1206 | READ_REQ_TOO_SHORT | parser | fixed header shorter than 36 bytes in `ReadReq` | emitted on wire |
| 1207 | READ_REQ_INCOMPLETE | parser | variable section shorter than declared lengths in `ReadReq` | emitted on wire |
| 1208 | READ_FREE_REQ_TOO_SHORT | parser | fixed header shorter than 28 bytes in `ReadFreeReq` | emitted on wire |
| 1209 | READ_FREE_REQ_INCOMPLETE | parser | variable section shorter than declared lengths in `ReadFreeReq` | emitted on wire |
| 1210 | DELETE_REQ_TOO_SHORT | parser | fixed header shorter than 36 bytes in `DeleteReq` | emitted on wire |
| 1211 | DELETE_REQ_INCOMPLETE | parser | variable section shorter than declared lengths in `DeleteReq` | emitted on wire |
| 1212 | READ_CELL_REQ_TOO_SHORT | parser | fixed header shorter than 32 bytes in `ReadCellReq` | emitted on wire |
| 1213 | READ_CELL_REQ_INCOMPLETE | parser | variable section shorter than declared lengths in `ReadCellReq` | emitted on wire |
| 1214 | DIFERIR_REQ_TOO_SHORT | parser | fixed header shorter than 68 bytes in `DiferirReq` | emitted on wire |
| 1215 | DIFERIR_CHILD_SALT_TOO_SHORT | parser | `ChildSalt` shorter than 16 bytes in `DiferirReq` | emitted on wire |
| 1216 | DIFERIR_REQ_INCOMPLETE | parser | variable section shorter than declared lengths in `DiferirReq` | emitted on wire |
| 1217 | CRUZAR_REQ_TOO_SHORT | parser | fixed header shorter than 52 bytes in `CruzarReq` | emitted on wire |
| 1218 | CRUZAR_REQ_INCOMPLETE | parser | variable section shorter than declared lengths in `CruzarReq` | emitted on wire |
| 1300 | REQUEST_ID_GENERATION_FAILED | builder | random ID generation fails in outgoing request builders | catalogued blind spot |
| 1301 | CHILD_SALT_GENERATION_FAILED | builder | child salt generation fails in outgoing `DiferirReqBytes` | catalogued blind spot |
| 2100 | DATABASE_ALREADY_EXISTS | business | `CreateDB` detects an existing store | emitted on wire |
| 2101 | DATABASE_NOT_FOUND | business | store missing in `DelDB`, `Read`, `Write`, `ReadFree`, `Delete`, `ReadCell`, `Diferir`, `Cruzar` | emitted on wire |
| 2102 | AUTHENTICATION_FAILED | business | `ResolveCellAuth` or `resolveCell` fails | emitted on wire |
| 2103 | MEMBRANE_NOT_FOUND | business | requested key does not exist in Bolt | emitted on wire |
| 2104 | PERMISSION_DENIED | business | genome lacks required read/write/delete permission | emitted on wire |
| 2105 | INVALID_GENOME_FLAGS | business | `CreateDB` receives a root genome already marked as migrated | emitted on wire |
| 2106 | CHILD_GENOME_ESCALATION | business | `Diferir` requests bits not owned by parent cell | emitted on wire |
| 2107 | MERGE_CAPABILITY_MISSING | business | `Cruzar` parent cell lacks `Merge` capability | emitted on wire |
| 3100 | STORE_CREATE_FAILED | storage | `NewStore` fails while creating Ouroboros or Bolt | emitted on wire |
| 3101 | STORE_OPEN_FAILED | storage | opening an existing store fails | catalogued blind spot |
| 3102 | STORE_DESTROY_FAILED | storage | `Destroy` fails during delete database | emitted on wire |
| 3103 | MEMBRANE_READ_FAILED | storage | Bolt read or membrane decode fails | emitted on wire |
| 3104 | MEMBRANE_WRITE_FAILED | storage | Bolt write fails while inserting or updating a membrane | emitted on wire |
| 3105 | MEMBRANE_DELETE_FAILED | storage | Bolt delete fails while deleting a membrane | emitted on wire |
| 3106 | INITIAL_CELL_APPEND_FAILED | storage | root cell append fails in `CreateDB` | emitted on wire |
| 3107 | CELL_APPEND_FAILED | storage | child append fails in `Diferir` or `Cruzar` | emitted on wire |
| 3108 | CELL_REFRESH_FAILED | storage | refresh path fails in `Read` or `Write` | emitted on wire |
| 3109 | RANDOM_SOURCE_FAILED | storage | random salt generation fails in server-side flows | emitted on wire |
| 3110 | CONFIG_LOAD_FAILED | startup | `.env` loading fails in `NewCentralHandler` | catalogued blind spot |
| 3111 | BASE_DIR_CREATE_FAILED | startup | `os.MkdirAll` fails during startup | catalogued blind spot |
| 3112 | STORE_DIR_READ_FAILED | startup | `os.ReadDir` fails during startup | catalogued blind spot |
| 3113 | STARTUP_STORE_OPEN_FAILED | startup | startup scan finds a store on disk but `Open` fails | catalogued blind spot |
| 9000 | INTERNAL_UNEXPECTED | internal | fallback for unmapped failures or impossible states | reserved fallback |

## Blind Spots Found In Source Review

These call sites can still fail but do not currently have a protocol response path because they belong to startup or client-side request building:

| Location | Failure | Assigned code |
| --- | --- | --- |
| `handler/handler.go` `LoadConfig(".env")` | startup config read error is ignored | 3110 |
| `handler/handler.go` `os.MkdirAll(config.DBPath, 0755)` | base directory creation error is ignored | 3111 |
| `handler/handler.go` `os.ReadDir(config.DBPath)` | startup directory scan failure is ignored | 3112 |
| `handler/handler.go` `Open(fullPath)` inside startup loop | failed store open is ignored | 3113 |
| `protocol/0x01deleteDB.go` `DeleteDBReqBytes` | random ID failure falls back to a fixed string | 1300 |
| `protocol/0x02write.go` `WriteReqBytes` | random ID failure is ignored | 1300 |
| `protocol/0x03read.go` `ReadReqBytes` | random ID failure is ignored | 1300 |
| `protocol/0x04readFree.go` `ReadFreeReqBytes` | random ID failure is ignored | 1300 |
| `protocol/0x06readCell.go` `ReadCellReqBytes` | random ID failure is ignored | 1300 |
| `protocol/0x07diferir.go` `DiferirReqBytes` | random ID failure is ignored | 1300 |
| `protocol/0x07diferir.go` `DiferirReqBytes` | child salt generation failure is ignored | 1301 |
| `protocol/0x08cruzar.go` `CruzarReqBytes` | random ID failure is ignored | 1300 |

## Notes

- Success responses keep their existing opcode-specific formats.
- Error responses are opcode-agnostic and always use the generic envelope.
- `Info` is intended for short routing hints such as `read.parse`, `delete.permission` or `router.opcode=0x29`.