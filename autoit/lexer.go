package autoit

import (
	"fmt"
	"io"
	"os"
	"strconv"
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
			//continue
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

			rTmp, err := l.ReadRune()
			if err == nil && rTmp == '=' {
				token.Data += "="
			} else {
				l.Move(-1)
			}
		default:
			if isIdent(r) {
				token.Type = tCALL
				l.Move(-1)
				token.Data = l.ReadIdent()
				if _, numErr := strconv.Atoi(token.Data); numErr == nil {
					token.Type = tNUMBER
					//TODO: Do a check for a period followed by another number, it's a floating point number!
				} else {
					switch token.Data {
						case "Exit":
							token.Type = tEXIT
						case "Null":
							token.Type = tNULL
						case "Default":
							token.Type = tDEFAULT
						case "True":
							token.Type = tBOOLEAN
						case "False":
							token.Type = tBOOLEAN
						case "Func":
							token.Type = tFUNC
						case "Return":
							token.Type = tFUNCRETURN
						case "EndFunc":
							token.Type = tFUNCEND
						case "If":
							token.Type = tIF
						case "Then":
							token.Type = tTHEN
						case "Else":
							token.Type = tELSE
						case "ElseIf":
							token.Type = tELSEIF
						case "EndIf":
							token.Type = tIFEND
						case "For":
							token.Type = tFOR
						case "To":
							token.Type = tTO
						case "Step":
							token.Type = tSTEP
						case "In":
							token.Type = tIN
						case "Next":
							token.Type = tNEXT
						case "While":
							token.Type = tWHILE
						case "WEnd":
							token.Type = tWEND
						case "With":
							token.Type = tWITH
						case "EndWith":
							token.Type = tWITHEND
						case "Do":
							token.Type = tDO
						case "Until":
							token.Type = tUNTIL
						case "Switch":
							token.Type = tSWITCH
						case "EndSwitch":
							token.Type = tSWITCHEND
						case "Select":
							token.Type = tSELECT
						case "EndSelect":
							token.Type = tSELECTEND
						case "Case":
							token.Type = tCASE
						case "ContinueCase":
							token.Type = tCASEREPEAT
						case "ContinueLoop":
							token.Type = tLOOPREPEAT
						case "ExitLoop":
							token.Type = tLOOPEXIT
						case "Dim":
							token.Type = tSCOPE
							token.Data = "Local"
						case "ReDim":
							token.Type = tREVAR
						case "Local":
							token.Type = tSCOPE
						case "Global":
							token.Type = tSCOPE
						case "Const":
							token.Type = tSCOPE
						case "Static":
							token.Type = tSCOPE
						case "Enum":
							token.Type = tENUM
						case "Volatile":
							token.Type = tVOLATILE
					}
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
func isIdent(r rune) bool {
	if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
		return true
	}
	return false
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
	if isIdent(r) || r == '-' {
		return true
	}
	return false
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
			//l.Move(-1)
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

/*func tokenLexer(script string) []*Token {
	script = strings.ReplaceAll(script, "\r\n", "\n")
	script = strings.ReplaceAll(script, "\r", "\n")
	script = strings.ReplaceAll(script, "\t", " ")
	lines := strings.Split(script, "\n")
	tokens := make([][]*Token, 0)

	lexingComment := false
	lexingQuote := ""
	for lineNumber, line := range lines {
		if line != "" {
			newTokens := make([]*Token, 0)
			linePos := 0
			for _, token := range strings.Split(line, " ") {
				//Skip empty words
				if token == "" {
					linePos++ //Someone's using multiple spaces...
					continue
				}

				//Skip comment blocks
				if token[0] == ';' {
					break
				}
				if token == "#comments-start" || token == "#cs" {
					lexingComment = true
					break
				}
				if lexingComment && (token == "#comments-end" || token == "#ce") {
					lexingComment = false
					break
				}
				if lexingComment {
					break
				}

				/*
				if token[0] == "\"" {
					lexingQuote = true
					if len(token) > 1 {
						lexingQuoteData
					}
				}
				if lexingQuote {
					if token[len(token)-1] == "\"" {
						lexingQuote = false
					}
					lexingQuoteData += " " + token
					break
				}
				//

				//Add the token
				newTokens = append(newTokens, NewToken(token, lineNumber, linePos))
				linePos += len(token) + 1 //Add current token length and space suffix to simulate character tracking
			}
			if len(newTokens) > 0 {
				tokens = append(tokens, newTokens)
			}
		}
	}

	return tokens
}*/