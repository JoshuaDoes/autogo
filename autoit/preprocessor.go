package autoit

import (
	"os"
	"strings"
)

//Preprocess steps through the script and prepares it for execution
func (vm *AutoItVM) Preprocess() error {
	includeErr := vm.PreprocessIncludes()
	if includeErr != nil {
		return includeErr
	}
	callErr := vm.PreprocessFuncCalls()
	if callErr != nil {
		return callErr
	}
	funcErr := vm.PreprocessFuncs()
	if funcErr != nil {
		return funcErr
	}
	return nil
}

func (vm *AutoItVM) PreprocessIncludes() error {
	startLine := true
	for {
		token := vm.ReadToken()
		if token == nil {
			break
		}
		switch token.Type {
		case tFLAG:
			if !startLine {
				return vm.Error("preprocess: unexpected flag")
			}
			startLine = false

			switch strings.ToLower(token.String()) {
			case "include":
				includeFile := vm.ReadToken()
				if includeFile.Type != tSTRING {
					return vm.Error("preprocess: expected string containing path to include")
				}
				includeScript, err := os.ReadFile(includeFile.String())
				if err != nil {
					return err
				}
				includeLexer := NewLexer(includeScript)
				includeTokens, err := includeLexer.GetTokens()
				if err != nil {
					return err
				}
				includeTokens = append(includeTokens, NewToken(tEOL, ""))
				vm.tokens = append(vm.tokens[:vm.pos], append(includeTokens, vm.tokens[vm.pos:]...)...)
				vm.RemoveTokens(vm.pos-2, vm.pos)
				vm.Move(-1)
				vm.Log("preprocess: include %s preloaded successfully: %v", includeFile.String(), includeTokens)
			}
		case tEOL, tCOMMENT:
			startLine = true
		default:
			startLine = false
		}
	}

	vm.Stop()
	return nil
}

func (vm *AutoItVM) PreprocessFuncs() error {
	startLine := true
	for {
		token := vm.ReadToken()
		if token == nil {
			break
		}
		switch token.Type {
		case tFUNC:
			if !startLine {
				return vm.Error("preprocess: unexpected func")
			}
			startLine = false

			funcErr := vm.PreprocessFunc()
			if funcErr != nil {
				return funcErr
			}
		case tEOL, tCOMMENT:
			startLine = true
		default:
			startLine = false
		}
	}

	vm.Stop()
	return nil
}

//PreprocessFunc reads and processes a function
func (vm *AutoItVM) PreprocessFunc() error {
	startPos := vm.GetPos()
	vm.Log("func start pos: %d %v", startPos, vm.Token())

	tCall := vm.ReadToken()
	if tCall.Type != tCALL {
		return vm.Error("preprocessor: expected call after func, instead got: %v", tCall)
	}

	tBlock := vm.ReadToken()
	if tBlock.Type != tLEFTPAREN {
		return vm.Error("preprocessor: expected block after func call, instead got: %v", tBlock)
	}

	funcArgs := make([]*FunctionArg, 0)
	for {
		funcArg := &FunctionArg{}

		tVar := vm.ReadToken()
		if tVar == nil || tVar.Type == tRIGHTPAREN || tVar.Type == tEOL {
			break
		}
		if tVar.Type != tVARIABLE {
			return vm.Error("preprocessor: expected variable in func call block, instead got: %v", tVar)
		}
		funcArg.Name = strings.ToLower(tVar.String())

		tOp := vm.ReadToken()
		switch tOp.Type {
		case tSEPARATOR:
			funcArgs = append(funcArgs, funcArg)
			continue
		case tRIGHTPAREN:
			funcArgs = append(funcArgs, funcArg)
			break
		case tOP:
			switch tOp.String() {
			case "=":
				tValue := vm.ReadToken()
				if tValue == nil {
					return vm.Error("preprocessor: expected default value for variable in func call block, instead got: %v", tValue)
				}

				funcArg.DefaultValue = tValue
				funcArgs = append(funcArgs, funcArg)
			default:
				return vm.Error("preprocessor: unexpected operator for variable in func call block: %v", tOp)
			}
		default:
			return vm.Error("preprocessor: unexpected token in func call block: %v", tOp)
		}
	}

	funcBlock := make([]*Token, 0)
	for {
		tBlock := vm.ReadToken()
		if tBlock == nil {
			return vm.Error("preprocessor: unexpected end of func")
		}

		if tBlock.Type == tFUNCEND {
			break
		}

		funcBlock = append(funcBlock, tBlock)
	}

	if _, exists := vm.funcs[tCall.String()]; exists {
		return vm.Error("preprocess: func %s already defined", tCall.String())
	}
	vm.funcs[strings.ToLower(tCall.String())] = &Function{Args: funcArgs, Block: funcBlock}
	vm.Log("preprocess: func %s preloaded successfully: %v", tCall.String(), vm.funcs[tCall.String()])

	endPos := vm.GetPos()
	if endPos > len(vm.tokens) {
		vm.Log("end pos greater than end")
	}
	vm.Log("func end pos: %d %v", endPos, vm.GetToken(endPos))
	vm.RemoveTokens(startPos+1, endPos)
	return nil
}

func (vm *AutoItVM) PreprocessFuncCalls() error {
	for {
		token := vm.ReadToken()
		if token == nil {
			break
		}
		switch token.Type {
		case tCALL:
			vm.Move(-1)
			callErr := vm.PreprocessFuncCall()
			if callErr != nil {
				return callErr
			}
		case tFUNC:
			vm.Log("preprocess: skipped definition of func %s", vm.Token().String())
			vm.Move(1)
		}
	}

	vm.Stop()
	return nil
}

//PreprocessFuncCall reads and processes a function call
func (vm *AutoItVM) PreprocessFuncCall() error {
	startPos := vm.GetPos()
	//vm.Log("call start pos: %d %v", startPos, vm.Token())

	tCall := vm.ReadToken()
	if tCall.Type != tCALL {
		return vm.Error("preprocessor: expected call, instead got: %v", tCall)
	}

	tStart := vm.ReadToken()
	if tStart.Type != tLEFTPAREN {
		return vm.Error("preprocessor: expected block after func call, instead got: %v", tStart)
	}

	callBlocks := make([][]*Token, 0)
	callBlock := make([]*Token, 0)
	depth := 0
	for {
		if depth < 0 {
			break
		}
		blockToken := vm.ReadToken()
		if blockToken == nil {
			break
		}
		//vm.Log("preprocessor: call: depth %d step %v %s", depth, blockToken.Type, blockToken)
		switch blockToken.Type {
		case tCALL:
			vm.Log("preprocessor: call: encountered nested call")
			vm.Move(-1)
			callErr := vm.PreprocessFuncCall()
			if callErr != nil {
				return callErr
			}
		case tSEPARATOR:
			if depth > 0 {
				return vm.Error("preprocessor: unexpected separator in nested func call block: %v", blockToken)
			}
			callBlocks = append(callBlocks, callBlock)
			callBlock = make([]*Token, 0)
		case tLEFTPAREN:
			depth++
		case tRIGHTPAREN:
			depth--
		default:
			callBlock = append(callBlock, blockToken)
		}
	}
	if len(callBlock) > 0 {
		callBlocks = append(callBlocks, callBlock)
	}

	functionCall := &FunctionCall{Name: tCall.String(), Block: callBlocks}
	tHandle := vm.AddHandle(functionCall)
	vm.SetToken(startPos, NewToken(tCALL, tHandle.Handle()))
	vm.Log("preprocess: call preloaded successfully: %v %v", tCall.String(), callBlocks)

	endPos := vm.GetPos()
	if endPos >= len(vm.tokens) {
		//vm.Log("call end pos greater than end")
	}
	//vm.Log("call end pos: %d %v", endPos, vm.GetToken(endPos))
	vm.RemoveTokens(startPos+1, endPos)
	return nil
}