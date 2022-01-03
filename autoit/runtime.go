package autoit

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"path/filepath"
)

type AutoItVM struct {
	//Runtime configuration
	Logger bool
	
	//Script trackers
	scriptPath string
	tokens []*Token
	pos int

	//Runtime and memory
	running bool
	suspended bool
	vars map[string]*Token
	parentScope *AutoItVM
}

func NewAutoItVM(scriptPath string, script []byte, parentScope *AutoItVM) (*AutoItVM, error) {
	scriptPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return nil, err
	}

	lexer := NewLexer(script)
	tokens, err := lexer.GetTokens()
	if err != nil {
		return nil, err
	}

	return &AutoItVM{
		scriptPath: scriptPath,
		tokens: tokens,
		parentScope: parentScope,
	}, nil
}

func (vm *AutoItVM) Run() error {
	if vm.Running() {
		return nil
	}
	vm.running = true
	vm.vars = make(map[string]*Token)
	for vm.Running() {
		if vm.Suspended() {
			time.Sleep(time.Millisecond * 1)
			continue
		}
		err := vm.Step()
		if err == io.EOF {
			vm.Stop()
			return nil
		}
		if err != nil {
			vm.Stop()
			return err
		}
	}
	return nil
}
func (vm *AutoItVM) Step() error {
	token := vm.ReadToken()
	if token == nil {
		return io.EOF
	}
	if token.Type == tEOL || token.Type == tCOMMENT {
		return nil
	}

	switch token.Type {
	case tILLEGAL:
		return vm.Error("illegal token encountered: %v", *token)
	case tEXIT:
		tExitCode := vm.ReadToken()
		if tExitCode != nil {
			os.Exit(tExitCode.Int())
		}
		os.Exit(0)
	case tSCOPE:
		vm.Move(-1)
		vm.Log("SCOPE %s", token.Data)
		eval := NewEvaluator(vm, vm.tokens[vm.pos:])
		_, tRead, err := eval.Eval(false)
		if err != nil {
			return vm.Error("error defining scoped variable: %v", err)
		}
		vm.Move(tRead)
	case tVARIABLE:
		vm.Move(-1)
		vm.Log("VARIABLE %s", token.Data)
		eval := NewEvaluator(vm, vm.tokens[vm.pos:])
		_, tRead, err := eval.Eval(false)
		if err != nil {
			return vm.Error("error setting variable: %v", err)
		}
		vm.Move(tRead)
	case tCALL: //Must be a function call, we aren't storing a call pointer here
		vm.Move(-1)
		vm.Log("CALL %s", token.Data)
		eval := NewEvaluator(vm, vm.tokens[vm.pos:])
		_, tRead, err := eval.Eval(false)
		if err != nil {
			return vm.Error("error executing call: %v", err)
		}
		vm.Move(tRead)
	case tFLAG:
		switch strings.ToLower(token.Data) {
		case "include":
			includeFile := vm.ReadToken()
			if includeFile.Type != tSTRING {
				return vm.Error("expected string containing path to include")
			}
			vm.Log("INCLUDE %s", includeFile.Data)

			includeLexer, err := NewLexerFromFile(includeFile.Data)
			if err != nil {
				return vm.Error("%v", err)
			}
			includeTokens, err := includeLexer.GetTokens()
			if err != nil {
				return vm.Error("%v", err)
			}

			tokens := make([]*Token, 0)
			tokens = append(tokens, vm.tokens[:vm.pos]...)
			tokens = append(tokens, &Token{Type: tEOL})
			tokens = append(tokens, includeTokens...)
			tokens = append(tokens, &Token{Type: tEOL})
			tokens = append(tokens, vm.tokens[vm.pos:]...)
			vm.tokens = tokens
		default:
			tOp := vm.ReadToken()
			switch tOp.Type {
			case tOP:
				if tOp.Data != "=" {
					return vm.Error("unexpected flag operator: %s", tOp.Data)
				}

				eval := NewEvaluator(vm, vm.tokens[vm.pos:])
				tValue, tRead, err := eval.Eval(true)
				if err != nil {
					return vm.Error("error getting flag value: %v", err)
				}

				vm.Move(tRead)
				vm.Log("FLAG %s = %s", token.Data, tValue.Data)
			case tEOL:
				switch strings.ToLower(token.Data) {
				case "debug":
					vm.Logger = true
				}

				vm.Move(-1)
				vm.Log("FLAG %s", token.Data)
			default:
				return vm.Error("unexpected token following flag: %v", tOp)
			}
		}
	default:
		return vm.Error("unexpected token: %v", *token)
	}

	for {
		nextToken := vm.ReadToken()
		if nextToken == nil {
			break
		}
		if nextToken.Type == tCOMMENT {
			continue
		}
		if nextToken.Type != tEOL {
			return vm.Error("expected end of line, instead found token: %v", *nextToken)
		}
		vm.Move(-1)
		break
	}

	return nil
}

func (vm *AutoItVM) Log(format string, params ...interface{}) {
	if !vm.Logger {
		return
	}
	if format[len(format)-1] != '\n' {
		format += "\n"
	}
	if params != nil {
		fmt.Printf("vm: " + format, params...)
	} else {
		fmt.Printf("vm: " + format)
	}
}
func (vm *AutoItVM) Error(format string, params ...interface{}) error {
	if params != nil {
		return fmt.Errorf("autoit: " + format, params...)
	}
	return fmt.Errorf("autoit: " + format)
}

func (vm *AutoItVM) Running() bool {
	return vm.running
}
func (vm *AutoItVM) Suspended() bool {
	return vm.suspended
}
func (vm *AutoItVM) Suspend() {
	vm.suspended = true
}
func (vm *AutoItVM) Resume() {
	vm.suspended = false
}
func (vm *AutoItVM) Stop() {
	vm.running = false
	vm.suspended = false
	vm.pos = 0
	vm.vars = make(map[string]*Token)
}

func (vm *AutoItVM) Tokens() []*Token {
	return vm.tokens
}
func (vm *AutoItVM) Token() *Token {
	return vm.tokens[vm.pos]
}
func (vm *AutoItVM) ReadToken() *Token {
	if vm.pos >= len(vm.tokens) {
		return nil
	}
	defer vm.Move(1)
	return vm.tokens[vm.pos]
}
func (vm *AutoItVM) Move(pos int) {
	vm.pos += pos
}

//GetVariable returns the specified variable or nil if it doesn't exist
func (vm *AutoItVM) GetVariable(variableName string) (*Token) {
	if variable, exists := vm.vars[strings.ToLower(variableName)]; exists {
		vm.Log("GET $%s", variableName)
		return variable
	}
	if vm.parentScope != nil {
		if variable := vm.parentScope.GetVariable(variableName); variable != nil {
			return variable
		}
	}
	return nil
}
func (vm *AutoItVM) SetVariable(variableName string, token *Token) {
	vm.Log("SET $%s = %v", variableName, *token)
	vm.vars[strings.ToLower(variableName)] = token
}