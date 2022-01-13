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
	preprocessed bool

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
		//Logger: true,
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

	if !vm.preprocessed {
		preprocess := vm.Preprocess()
		if preprocess != nil {
			return preprocess
		}
		vm.preprocessed = true
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
		vm.HandleCall(&FunctionCall{Name: vm.exitMethod})
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
		return nil
	case tFLAG:
		switch strings.ToLower(token.String()) {
		case "include":
			return vm.Error("unexpected include attempt, did preprocessing fail?")
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
	time := fmt.Sprintf("%v", time.Now())
	if params != nil {
		fmt.Printf("[" + time + "] vm: " + format, params...)
	} else {
		fmt.Printf("[" + time + "] vm: " + format)
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

func (vm *AutoItVM) ExtendVM(tokens []*Token, preprocess bool) (*AutoItVM, error) {
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
	if preprocess {
		return vmNew, vmNew.Preprocess()
	}
	return vmNew, nil
}

//Tokens returns the full slice of tokens from the VM
func (vm *AutoItVM) Tokens() []*Token {
	return vm.tokens
}
//Token returns the token at the current position
func (vm *AutoItVM) Token() *Token {
	if vm.pos >= len(vm.tokens) {
		return nil
	}
	return vm.tokens[vm.pos]
}
//ReadToken returns the token at the current position and moves the position forward by one
func (vm *AutoItVM) ReadToken() *Token {
	if vm.pos >= len(vm.tokens) {
		return nil
	}
	defer vm.Move(1)
	return vm.tokens[vm.pos]
}
//GetToken returns the token at the given position
func (vm *AutoItVM) GetToken(pos int) *Token {
	if pos >= len(vm.tokens) {
		return nil
	}
	return vm.tokens[pos]
}
//SetToken sets the token at given position to the given token
func (vm *AutoItVM) SetToken(pos int, token *Token) {
	if pos >= len(vm.tokens) {
		return
	}
	vm.tokens[pos] = token
}
//RemoveTokens removes the specified position ranges of tokens
func (vm *AutoItVM) RemoveTokens(startPos, endPos int) {
	if startPos >= endPos {
		return
	}

	vm.Log("removing tokens %d to %d", startPos, endPos)
	//vm.Log("old tokens: %v\n", vm.tokens)
	//vm.Log("old pos: %d\n", vm.pos)
	vm.tokens = append(vm.tokens[:startPos], vm.tokens[endPos:]...)
	vm.Move(-1 * (endPos-startPos))
	//vm.Log("new tokens: %v\n", vm.tokens)
	//vm.Log("new pos: %d\n", vm.pos)
}
//RemoveToken removes the specified position from the tokens
func (vm *AutoItVM) RemoveToken(pos int) {
	if pos >= len(vm.tokens) {
		return
	}

	vm.Log("removing token %d: %v", pos, vm.tokens[pos])
	vm.tokens = append(vm.tokens[:pos], vm.tokens[pos:]...)
}
//GetPos returns the current position
func (vm *AutoItVM) GetPos() int {
	return vm.pos
}
//Move moves the position by the specified amount
func (vm *AutoItVM) Move(direction int) {
	vm.pos += direction
	/*if vm.Token() != nil {
		vm.Log("moved to: %d %v", vm.pos, vm.Token())
	}*/
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
//AddHandle creates a handle from the given token and returns the handle as a token
func (vm *AutoItVM) AddHandle(value interface{}) *Token {
	handleId := uuid.NewString()
	vm.handles[handleId] = value
	return NewToken(tHANDLE, handleId)
}
//DestroyHandle destroys the given handle id
func (vm *AutoItVM) DestroyHandle(handleId string) {
	delete(vm.handles, handleId)
}

func (vm *AutoItVM) MapGet(handleId, key string) *Token {
	return vm.handles[handleId].(map[string]*Token)[key]
}
func (vm *AutoItVM) MapSet(handleId, key string, value *Token) {
	vm.handles[handleId].(map[string]*Token)[key] = value
}