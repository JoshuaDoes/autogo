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
	funcs map[string]*Function
	pos int

	//Runtime and memory
	running bool
	suspended bool
	returnValue *Token
	vars map[string]*Token
	parentScope *AutoItVM
	stdout, stderr string
}

func NewAutoItScriptVM(scriptPath string, script []byte, parentScope *AutoItVM) (*AutoItVM, error) {
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
		funcs: make(map[string]*Function),
		vars: make(map[string]*Token),
		returnValue: NewToken(tNUMBER, 0),
	}, nil
}

func NewAutoItTokenVM(scriptPath string, tokens []*Token, parentScope *AutoItVM) (*AutoItVM, error) {
	scriptPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return nil, err
	}

	return &AutoItVM{
		scriptPath: scriptPath,
		tokens: tokens,
		parentScope: parentScope,
		funcs: make(map[string]*Function),
		vars: make(map[string]*Token),
		returnValue: NewToken(tNUMBER, 0),
	}, nil
}

func (vm *AutoItVM) Run() error {
	if vm.Running() {
		return nil
	}

	preprocess := vm.Preprocess()
	if preprocess != nil {
		return preprocess
	}

	vm.running = true
	vm.vars = make(map[string]*Token)
	for vm.Running() {
		if vm.Suspended() {
			time.Sleep(time.Millisecond * 1)
			continue
		}
		step := vm.Step()
		if step == io.EOF {
			vm.Stop()
			return nil
		}
		if step != nil {
			vm.Stop()
			return step
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
	vm.Log("step: %v", *token)

	switch token.Type {
	case tILLEGAL:
		return vm.Error("illegal token encountered: %v", *token)
	case tEXIT:
		tExitCode := vm.ReadToken()
		if tExitCode != nil {
			os.Exit(tExitCode.Int())
		}
		os.Exit(0)
	case tSCOPE, tVARIABLE, tCALL, tFUNC, tFUNCRETURN:
		vm.Move(-1)
		eval := NewEvaluator(vm, vm.tokens[vm.pos:])
		_, tRead, err := eval.Eval(false)
		if err != nil {
			return err
		}
		vm.Move(tRead)
	case tFLAG:
		switch strings.ToLower(token.String()) {
		case "include":
			includeFile := vm.ReadToken()
			if includeFile.Type != tSTRING {
				return vm.Error("expected string containing path to include")
			}
			vm.Log("INCLUDE %s", includeFile.String())

			includeLexer, err := NewLexerFromFile(includeFile.String())
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
				if tOp.String() != "=" {
					return vm.Error("unexpected flag operator: %s", tOp.String())
				}

				eval := NewEvaluator(vm, vm.tokens[vm.pos:])
				tValue, tRead, err := eval.Eval(true)
				if err != nil {
					return vm.Error("error getting flag value: %v", err)
				}

				vm.Move(tRead)
				vm.Log("FLAG %s = %s", token.String(), tValue.String())
			case tEOL:
				switch strings.ToLower(token.String()) {
				case "debug":
					vm.Logger = true
				}

				vm.Move(-1)
				vm.Log("FLAG %s", token.String())
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
}

func (vm *AutoItVM) Stdout() string {
	return vm.stdout
}
func (vm *AutoItVM) Stderr() string {
	return vm.stderr
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
		if variable := vm.parentScope.GetVariable(strings.ToLower(variableName)); variable != nil {
			return variable
		}
	}
	return nil
}
func (vm *AutoItVM) SetVariable(variableName string, token *Token) {
	vm.Log("SET $%s = %v", variableName, *token)
	vm.vars[strings.ToLower(variableName)] = token
}