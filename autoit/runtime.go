package autoit

import (
	"fmt"
	"io"
	"os"
	"time"
)

type AutoItVM struct {
	//Config
	Logger bool
	
	//Runtime
	tokens []*Token
	pos int
	running bool
	suspended bool
	vars map[string]*Token

	//Counters and trackers
	region int
	block int
}

func NewAutoItVM(script []byte) (*AutoItVM, error) {
	lexer := NewLexer(script)
	tokens, err := lexer.GetTokens()
	if err != nil {
		return nil, err
	}

	return &AutoItVM{
		tokens: tokens,
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

func (vm *AutoItVM) Tokens() []*Token {
	return vm.tokens
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
func (vm *AutoItVM) ReadUntilEOL(line []*Token) (tokens []*Token) {
	tokens = make([]*Token, 0)
	counter := 0
	for {
		counter++
		lineToken := &Token{}

		/*
		if line != nil {
			if len(line) == 0 {
				return
			}
			if counter >= len(line) {
				break
			}
			lineToken = line[counter-1]
		} else {
		*/
			lineToken = vm.ReadToken()
			if lineToken == nil {
				vm.Move(-1)
				return
			}
		//}

		if lineToken.Type == tEXTEND {
			tLineEnd := vm.ReadToken()
			if tLineEnd == nil || tLineEnd.Type != tEOL {
				vm.Move(-1)
				return
			}
			continue
		}
		if lineToken.Type == tEOL {
			break
		}
		tokens = append(tokens, lineToken)
	}
	return
}

//ReadBlock reads until the next root level BLOCKEND token
func (vm *AutoItVM) ReadBlock(block []*Token) []*Token {
	blockDepth := -1
	blockTokens := make([]*Token, 0)

	counter := 0
	for {
		counter++
		blockToken := &Token{}

		if block != nil {
			if len(block) == 0 {
				return nil
			}
			if counter >= len(block) {
				break
			}
			blockToken = block[counter-1]
		} else {
			blockToken = vm.ReadToken()
			if blockToken == nil {
				break
			}
		}

		blockTokens = append(blockTokens, blockToken)
		fmt.Println("added to block", blockDepth, blockToken)

		switch blockToken.Type {
		case tBLOCK:
			blockDepth++
			if blockDepth == 0 {
				blockDepth = 1
			}
		case tBLOCKEND:
			blockDepth--
		default:
		}

		if blockDepth == 0 {
			break
		}
	}

	return blockTokens
}
func (vm *AutoItVM) Eval(tokens ...*Token) []*Token {
	for i := 0; i < len(tokens); i++ {
		fmt.Println("eval:", tokens[i])

		switch tokens[i].Type {
		case tSTRING, tNUMBER:
			switch tokens[i+1].Type {
			case tOP:
				switch tokens[i+1].Data {
				case "&":
					fmt.Println("Found ampersand!")
				default:
					return nil
				}
			default:
				return nil
			}
			return []*Token{tokens[i]}
		case tBLOCK:
			blockTokens := vm.ReadBlock(tokens[i+1:])
			if blockTokens == nil {
				return nil
			}

			curExpr := make([]*Token, 0)
			allEval := make([]*Token, 0)
			for j := 0; j < len(blockTokens); j++ {
				switch blockTokens[j].Type {
				case tSEPARATOR:
					if len(curExpr) == 0 {
						return nil
					}
					allEval = append(allEval, vm.Eval(curExpr...)...)
					if j < len(blockTokens) {
						if blockTokens[j+1].Type == tBLOCKEND {
							return nil
						}
					}
				default:
					curExpr = append(curExpr, blockTokens[j])
				}
			}
			if len(curExpr) > 0 {
				allEval = append(allEval, vm.Eval(curExpr...)...)
			}
			
			return allEval
		case tCALL:
			callTokens := vm.ReadBlock(tokens[i+1:])
			if callTokens == nil {
				return nil
			}

			callTokens = vm.Eval(callTokens...)

			vm.HandleFunc(tokens[i].Data, callTokens)
			return nil
		case tSCOPE:
			tVariable := tokens[i+1]
			if tVariable.Type != tVARIABLE {
				return nil
			}

			tOp := tokens[i+2]
			if tOp.Type != tOP && tOp.Data != "=" {
				return nil
			}

			tValue := vm.ReadUntilEOL(tokens[i+3:])
			if len(tValue) == 0 {
				return nil
			}
			fmt.Println("value:", tValue[0])
			tValue = vm.Eval(tValue...)
			if len(tValue) == 0 {
				return nil
			}
			fmt.Println("value:", tValue[0])
			vm.vars[tVariable.Data] = tValue[0] //TODO: Determine scopes other than "Global" via tokens[i].Data
		case tVARIABLE:
			tOp := tokens[i+1]
			switch tOp.Type {
			case tBLOCKEND:
				return []*Token{vm.vars[tokens[i].Data]}
			}
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
	//vm.Log("Line start: %v", *token)

	switch token.Type {
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
		switch token.Data {
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
				vm.Log("FLAG %s = %s", token.Data, tValue.Data)
				vm.Move(tRead)
			case tEOL:
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