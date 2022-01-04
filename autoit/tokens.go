package autoit

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type Token struct {
	Type       TokenType //Type of the token
	Data       string    //String representing the token
	LineNumber int       //Starts from 0
	LinePos    int       //Starts from 0
}
func NewToken(tType TokenType, data interface{}) *Token {
	switch data.(type) {
	case int, int32, int64:
		return &Token{Type: tNUMBER, Data: fmt.Sprintf("%d", data)}
	case float32, float64:
		return &Token{Type: tNUMBER, Data: fmt.Sprintf("%f", data)}
	case string:
		if data.(string) != "" {
			return &Token{Type: tSTRING, Data: data.(string)}
		}
	case bool:
		boolTF := data.(bool)
		if boolTF {
			return &Token{Type: tBOOLEAN, Data: "True"}
		}
		return &Token{Type: tBOOLEAN, Data: "False"}
	case byte:
		return &Token{Type: tBINARY, Data: "0x" + fmt.Sprintf("%x", data)}
	case []byte:
		return &Token{Type: tBINARY, Data: hex.EncodeToString(data.([]byte))}
	}
	return &Token{Type: tType, Data: fmt.Sprintf("%v", data)}
}

func (t *Token) IsEmpty() bool {
	return t.Data == ""
}
func (t *Token) String() string {
	if t.IsEmpty() {
		return ""
	}
	switch t.Type {
	case tBINARY:
		return string(t.Bytes())
	}
	return t.Data
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
		return 0
	}
	return number
}
func (t *Token) Int64() int64 {
	if t.IsEmpty() {
		return 0
	}
	data := t.Data
	data = strings.ReplaceAll(data, "\r", "")
	data = strings.ReplaceAll(data, "\n", "")
	data = strings.ReplaceAll(data, "\t", "")
	data = strings.ReplaceAll(data, " ", "")
	number, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0
	}
	return number
}
func (t *Token) Float64() float64 {
	if t.IsEmpty() {
		return 0
	}
	data := t.Data
	data = strings.ReplaceAll(data, "\r", "")
	data = strings.ReplaceAll(data, "\n", "")
	data = strings.ReplaceAll(data, "\t", "")
	data = strings.ReplaceAll(data, " ", "")
	number, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return 0
	}
	return number
}
func (t *Token) Bytes() []byte {
	if t.IsEmpty() {
		return make([]byte, 0)
	}
	switch t.Type {
	case tBINARY:
		hexadecimal, err := hex.DecodeString(t.Data)
		if err != nil {
			return nil
		}
		return hexadecimal
	}
	return []byte(t.Data)
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
	tBINARY TokenType = "BINARY"
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