package autoit

import (
	"strconv"
	"strings"
)

type Token struct {
	Type       TokenType //Type of the token
	Data       string    //String representing the token
	LineNumber int       //Starts from 0
	LinePos    int       //Starts from 0
}
func NewToken(tType TokenType, data string) *Token {
	return &Token{Type: tType, Data: data}
}

func (t *Token) IsEmpty() bool {
	return t.Data == ""
}
func (t *Token) Int() int {
	if t.IsEmpty() {
		return 0
	}
	data := t.Data
	data = strings.ReplaceAll(data, "\r", "")
	data = strings.ReplaceAll(data, "\n", "")
	data = strings.ReplaceAll(data, "\t", "")
	data = strings.ReplaceAll(data, " ", "")
	number, err := strconv.Atoi(data)
	if err != nil {
		return 1
	}
	return number
}

type TokenType string
const (
	tILLEGAL TokenType = "ILLEGAL"
	tEOL TokenType = "EOL"
	tEXTEND TokenType = "EXTEND"
	tCALL TokenType = "CALL"
	tSTRING TokenType = "STRING"
	tFLAG TokenType = "FLAG"
	tMACRO TokenType = "MACRO"
	tCOMMENT TokenType = "COMMENT"
	tVARIABLE TokenType = "VARIABLE"
	tBLOCK TokenType = "BLOCK"
	tBLOCKEND TokenType = "BLOCKEND"
	tMAP TokenType = "MAP"
	tMAPEND TokenType = "MAPEND"
	tSEPARATOR TokenType = "SEPARATOR"
	tOP TokenType = "OP"
	tEXIT TokenType = "EXIT"
	tNULL TokenType = "NULL"
	tDEFAULT TokenType = "DEFAULT"
	tBOOLEAN TokenType = "BOOLEAN"
	tNUMBER TokenType = "NUMBER"
	tFUNC TokenType = "FUNC"
	tFUNCRETURN TokenType = "FUNCRETURN"
	tFUNCEND TokenType = "FUNCEND"
	tIF TokenType = "IF"
	tTHEN TokenType = "THEN"
	tELSE TokenType = "ELSE"
	tELSEIF TokenType = "ELSEIF"
	tIFEND TokenType = "IFEND"
	tFOR TokenType = "FOR"
	tTO TokenType = "TO"
	tSTEP TokenType = "STEP"
	tIN TokenType = "IN"
	tNEXT TokenType = "NEXT"
	tWHILE TokenType = "WHILE"
	tWEND TokenType = "WEND"
	tWITH TokenType = "WITH"
	tWITHEND TokenType = "WITHEND"
	tDO TokenType = "DO"
	tUNTIL TokenType = "UNTIL"
	tSWITCH TokenType = "SWITCH"
	tSWITCHEND TokenType = "SWITCHEND"
	tSELECT TokenType = "SELECT"
	tSELECTEND TokenType = "SELECTEND"
	tCASE TokenType = "CASE"
	tCASEREPEAT TokenType = "CASEREPEAT"
	tLOOPREPEAT TokenType = "LOOPREPEAT"
	tLOOPEXIT TokenType = "LOOPEXIT"
	tSCOPE TokenType = "SCOPE"
	tREVAR TokenType = "REVAR"
	tENUM TokenType = "ENUM"
	tVOLATILE TokenType = "VOLATILE"
)