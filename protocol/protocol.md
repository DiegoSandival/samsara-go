/*CREATE_DB (0x20)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[Secret Len: 4]
[DB Size: 4]
[Genome: 4]|
[DB Name: N]
[Secret: M]*/

/*CREATE_DB result
[ID: 16]
[Status: 4]*/

/*DELETE_DB (0x21)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Secret: M]
*/

/*DELETE_DB result
[ID: 16]
[Status: 4]*/

/*WRITE (0x22)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Key Len: 4]
[Value Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Key: M]
[Value: P]
[Secret: Q]*/

/*WRITE result
[ID: 16]
[Status: 4]*/

/*
READ (0x23)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Key Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Key: M]
[Secret: P]
*/

/*
READ Result
[ID: 16]
[Status: 4]
[CellIndex: 4]
[Value: 4]
*/

/*READ_FREE (0x24)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[Key Len: 4] |
[DB Name: N]
[Key: M]*/

/*READ_FREE result
[ID: 16]
[Status: 4]
[Value: 4]*/

/*DELETE (0x25)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Key Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Key: M]
[Secret: P]*/

/*DELETE result
[ID: 16]
[Status: 4]*/

/*READ_CELL (0x26)
[Opcode: 4]
[ID: 16]
[CellIndex: 4]
[DB Name Len: 4]
[Secret Len: 4] |
[DB Name: N]
[Secret: M]*/

/*READ_CELL result
[ID: 16]
[Status: 4]
[Value: 4]*/

/*DIFERIR (0x27)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[CellIndex: 4]
[ParentSecret Len: 4]
[ChildGenome: 4]
[X: 4]
[Y: 4]
[Z: 4]
[ChildSalt: 16]
[ChildSecret Len: 4] |
[DB Name: N]
[ParentSecret: M]
[ChildSecret: P]*/

/*DIFERIR result
[ID: 16]
[Status: 4]
[CellIndex: 4]*/

/*CRUZAR (0x28)
[Opcode: 4]
[ID: 16]
[DB Name Len: 4]
[CellIndexA: 4]
[CellIndexB: 4]
[SecretA Len: 4]
[SecretB Len: 4]
[X: 4]
[Y: 4]
[Z: 4]
[ChildSecret Len: 4] |
[DB Name: N]
[SecretA: M]
[SecretB: P]
[ChildSecret: Q]*/

/*CRUZAR result
[ID: 16]
[Status: 4]
[CellIndex: 4]*/

/*GENERIC ERROR result
[ID: 16]
[ErrorID: 4]
[Info: N]*/

See `protocol/error_codes.md` for the centralized catalog and mapping table.