package autoit

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"path/filepath"

	"github.com/google/uuid"
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
	error int
	exitCode int
	exitMethod string
	extended int
	returnValue *Token
	numParams int
	vars map[string]*Token
	handles map[string]interface{}
	parentScope *AutoItVM
	stdout, stderr string
	ranIfStatement bool
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
		handles: make(map[string]interface{}, 0),
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
		handles: make(map[string]interface{}, 0),
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
			vm.exitCode = tExitCode.Int()
			vm.Move(-1)
			_, _, err := NewEvaluator(vm, vm.tokens[vm.pos:]).Eval(false)
			if err != nil {
				return err
			}
		}
		if vm.exitMethod == "" {
			os.Exit(vm.exitCode)
		}
		vm.HandleFunc(vm.exitMethod, nil)
	case tSCOPE, tVARIABLE, tCALL, tFUNC, tIF, tELSE, tELSEIF, tIFEND, tCASE, tSWITCH, tSWITCHEND:
		vm.Move(-1)
		eval := NewEvaluator(vm, vm.tokens[vm.pos:])
		_, tRead, err := eval.Eval(false)
		vm.Move(tRead)
		if err != nil {
			return err
		}
	case tFUNCRETURN:
		eval := NewEvaluator(vm, vm.tokens[vm.pos:])
		tValue, tRead, err := eval.Eval(true)
		vm.Move(tRead)
		if err != nil {
			return err
		}
		vm.SetReturnValue(tValue)
		vm.Stop()
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
		return fmt.Errorf("runtime: " + format, params...)
	}
	return fmt.Errorf("runtime: " + format)
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

func (vm *AutoItVM) GetError() int {
	return vm.error
}
func (vm *AutoItVM) SetError(error int) {
	vm.error = error
}
func (vm *AutoItVM) GetExtended() int {
	return vm.extended
}
func (vm *AutoItVM) SetExtended(extended int) {
	vm.extended = extended
}
func (vm *AutoItVM) GetReturnValue() *Token {
	return vm.returnValue
}
func (vm *AutoItVM) SetReturnValue(returnValue *Token) {
	vm.returnValue = returnValue
}

func (vm *AutoItVM) ExtendVM(tokens []*Token) (*AutoItVM, error) {
	vmPtr := *vm
	vmNew := &vmPtr
	vmNew.running = false
	vmNew.suspended = false
	vmNew.pos = 0
	vmNew.returnValue = NewToken(tNUMBER, 0)
	vmNew.error = 0
	vmNew.extended = 0
	vmNew.ranIfStatement = false
	vmNew.tokens = tokens
	vmNew.parentScope = vm
	return vmNew, vmNew.Preprocess()
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

//GetVariable returns the token for the specified variable or nil if it doesn't exist
func (vm *AutoItVM) GetVariable(variableName string) *Token {
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
//SetVariable sets the specified variable to the given token
func (vm *AutoItVM) SetVariable(variableName string, token *Token) {
	vm.Log("SET $%s = %v", variableName, *token)
	vm.vars[strings.ToLower(variableName)] = token
}

//GetHandle returns the token for the specified handle or nil if it doesn't exist
func (vm *AutoItVM) GetHandle(handleId string) interface{} {
	if handle, exists := vm.handles[handleId]; exists {
		return handle
	}
	if vm.parentScope != nil {
		if handle := vm.parentScope.GetHandle(handleId); handle != nil {
			return handle
		}
	}
	return nil
}
//AddHandle creates a handle from the given token and returns the handle id
func (vm *AutoItVM) AddHandle(value interface{}) string {
	handleId := uuid.NewString()
	vm.handles[handleId] = value
	return handleId
}
//DestroyHandle destroys the given handle id
func (vm *AutoItVM) DestroyHandle(handleId string) {
	vm.handles[handleId] = nil
}