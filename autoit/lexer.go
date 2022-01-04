package autoit

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

type Lexer struct {
	curLineNum, curLinePos int
	position int
	data []byte
}

func NewLexer(script []byte) *Lexer {
	tmpScript := string(script)
	tmpScript = strings.ReplaceAll(tmpScript, "\r\n", "\n")
	tmpScript = strings.ReplaceAll(tmpScript, "\r", "\n")
	tmpScript = strings.ReplaceAll(tmpScript, "\t", " ")
	return &Lexer{data: []byte(tmpScript)}
}
func NewLexerFromFile(path string) (*Lexer, error) {
	script, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewLexer(script), nil
}
func (l *Lexer) GetTokens() ([]*Token, error) {
	if len(l.data) > 0 {
		tokens := make([]*Token, 0)
		for {
			token, err := l.ReadToken()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}

			//fmt.Println("lexer:", *token)
			tokens = append(tokens, token)
		}
		return tokens, nil
	}

	return nil, fmt.Errorf("lexer: no input data")
}
func (l *Lexer) ReadToken() (*Token, error) {
	token := &Token{Type: tILLEGAL, Data: "", LineNumber: l.curLineNum, LinePos: l.curLinePos}

	for {
		r, err := l.ReadRune()
		if err != nil {
			return nil, err
		}
		token.Data = string(r)

		switch r {
		case ' ':
			continue //Skip whitespace
		case '\n':
			token.Type = tEOL
			token.Data = ""
		case '_':
			/*
				$sName = "Joshua" & _
					"Does"
				$iAge = 0 + _
					21
			*/
			token.Type = tEXTEND
		case '"':
			token.Type = tSTRING
			token.Data = l.ReadUntil('"', true)
		case '\'':
			token.Type = tSTRING
			token.Data = l.ReadUntil('\'', true)
		case '#':
			token.Type = tFLAG
			token.Data = l.ReadFlag()

			switch token.Data {
			case "cs", "comments-start":
				token.Type = tCOMMENT
				token.Data = ""
				for {
					token.Data += l.ReadUntil('#', false)
					flag := l.ReadFlag()
					if flag == "ce" || flag == "comments-end" {
						break
					}
					if flag == "" {
						return nil, fmt.Errorf("lexer: comment block beginning at %d:%d does not end", token.LineNumber, token.LinePos)
					}
				}
			}
		case '@':
			token.Type = tMACRO
			token.Data = l.ReadIdent()
		case ';':
			token.Type = tCOMMENT
			token.Data = l.ReadUntil('\n', false)
		case '$':
			token.Type = tVARIABLE
			token.Data = l.ReadIdent()
		case '(':
			token.Type = tBLOCK
		case ')':
			token.Type = tBLOCKEND
		case '[':
			token.Type = tMAP
		case ']':
			token.Type = tMAPEND
		case ',':
			token.Type = tSEPARATOR
		case '=', '&', '+', '-', '*', '/', '<', '>':
			token.Type = tOP

			rEquals, err := l.ReadRune()
			if err == nil && rEquals == '=' {
				token.Data += "="
			} else {
				l.Move(-1)
			}
		default:
			if unicode.IsDigit(r) {
				rX, err := l.ReadRune()
				if err == nil && r == '0' && (rX == 'x' || rX == 'X') {
					tmpBinary, err := l.ReadBinary()
					if err == nil {
						token.Type = tBINARY
						token.Data = tmpBinary
					}
				} else {
					l.Move(-2)
					tmpNumber, err := l.ReadNumber()
					if err == nil {
						token.Type = tNUMBER
						token.Data = tmpNumber
					}
				}
			} else if isIdent(r) {
				l.Move(-1)
				token.Type = tCALL
				token.Data = l.ReadIdent()
				switch strings.ToLower(token.Data) {
					case "exit":
						token.Type = tEXIT
					case "null":
						token.Type = tNULL
					case "default":
						token.Type = tDEFAULT
					case "true":
						token.Type = tBOOLEAN
					case "false":
						token.Type = tBOOLEAN
					case "and":
						token.Type = tAND
					case "or":
						token.Type = tOR
					case "not":
						token.Type = tNOT
					case "func":
						token.Type = tFUNC
					case "return":
						token.Type = tFUNCRETURN
					case "endfunc":
						token.Type = tFUNCEND
					case "if":
						token.Type = tIF
					case "then":
						token.Type = tTHEN
					case "else":
						token.Type = tELSE
					case "elseif":
						token.Type = tELSEIF
					case "endif":
						token.Type = tIFEND
					case "for":
						token.Type = tFOR
					case "to":
						token.Type = tTO
					case "step":
						token.Type = tSTEP
					case "in":
						token.Type = tIN
					case "next":
						token.Type = tNEXT
					case "while":
						token.Type = tWHILE
					case "wend":
						token.Type = tWEND
					case "with":
						token.Type = tWITH
					case "endwith":
						token.Type = tWITHEND
					case "do":
						token.Type = tDO
					case "until":
						token.Type = tUNTIL
					case "switch":
						token.Type = tSWITCH
					case "endswitch":
						token.Type = tSWITCHEND
					case "select":
						token.Type = tSELECT
					case "endselect":
						token.Type = tSELECTEND
					case "case":
						token.Type = tCASE
					case "continuecase":
						token.Type = tCASEREPEAT
					case "continueloop":
						token.Type = tLOOPREPEAT
					case "exitloop":
						token.Type = tLOOPEXIT
					case "dim":
						token.Type = tSCOPE
						token.Data = "Local"
					case "redim":
						token.Type = tREVAR
					case "local":
						token.Type = tSCOPE
					case "global":
						token.Type = tSCOPE
					case "const":
						token.Type = tSCOPE
					case "static":
						token.Type = tSCOPE
					case "enum":
						token.Type = tENUM
					case "volatile":
						token.Type = tVOLATILE
				}
			}
		}

		return token, nil
	}

	return token, nil
}

func (l *Lexer) ReadRune() (rune, error) {
	if l.position >= len(l.data) {
		return '\n', io.EOF
	}

	r := rune(l.data[l.position])
	l.Move(1)
	if r == '\n' {
		l.curLineNum++
		l.curLinePos = 0
	}

	return r, nil
}
func (l *Lexer) ReadNumber() (string, error) {
	read := ""
	readDeci := false
	for {
		r, err := l.ReadRune()
		if err == io.EOF {
			break
		}
		if r == '.' {
			if readDeci {
				return "", fmt.Errorf("two decimal places in number")
			}
			read += "."
			readDeci = true
			continue
		}
		if unicode.IsDigit(r) {
			read += string(r)
			continue
		}

		l.Move(-1)
		break
	}
	if read == "" {
		return "", fmt.Errorf("no number was read")
	}
	return read, nil
}
func isBinary(r rune) bool {
	if unicode.IsDigit(r) {
		return true
	}
	switch r {
	case 'a', 'A', 'b', 'B', 'c', 'C', 'd', 'D', 'e', 'E', 'f', 'F':
		return true
	}
	return false
}
func (l *Lexer) ReadBinary() (string, error) {
	read := ""
	for {
		r, err := l.ReadRune()
		if err == io.EOF {
			break
		}
		if isBinary(r) {
			read += string(r)
			continue
		}

		l.Move(-1)
		break
	}
	if read == "" {
		return "", fmt.Errorf("no binary was read")
	}
	return read, nil
}
func isIdent(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
func (l *Lexer) ReadIdent() string {
	read := ""
	for {
		r, err := l.ReadRune()
		if err == io.EOF {
			break
		}
		if isIdent(r) {
			read += string(r)
			continue
		}

		l.Move(-1)
		break
	}
	return read
}
func isFlag(r rune) bool {
	return isIdent(r) || r == '-'
}
func (l *Lexer) ReadFlag() string {
	read := ""
	for {
		r, err := l.ReadRune()
		if err == io.EOF {
			break
		}
		if isFlag(r) {
			read += string(r)
			continue
		}

		l.Move(-1)
		break
	}
	return read
}
func (l *Lexer) ReadUntil(r rune, escapable bool) string {
	read := ""
	for {
		rTmp, err := l.ReadRune()
		if err != nil {
			break
		}

		if escapable && rTmp == '\\' {
			rTmpEscaped, _ := l.ReadRune()
			if rTmpEscaped == '\\' {
				read += "\\"
			} else {
				l.Move(-1)
			}
		}

		if rTmp == r {
			break
		}
		read += string(rTmp)
	}
	return read
}
func (l *Lexer) Move(pos int) {
	l.position += pos
	l.curLinePos += pos
	if l.curLinePos < 0 {
		l.curLinePos = 0
	}
}