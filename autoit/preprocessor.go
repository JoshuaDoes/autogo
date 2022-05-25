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
	/*ifErr := vm.PreprocessIfs()
	if ifErr != nil {
		return ifErr
	}
	switchErr := vm.PreprocessSwitches()
	if switchErr != nil {
		return switchErr
	}*/
	forErr := vm.PreprocessFors()
	if forErr != nil {
		return forErr
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
				//vm.Log("preprocess: include %s preloaded successfully: %v", includeFile.String(), includeTokens)
				vm.Log("preprocess: include %s preloaded successfully", includeFile.String())
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

func (vm *AutoItVM) PreprocessFors() error {
	startLine := true
	for {
		token := vm.ReadToken()
		if token == nil {
			break
		}
		switch token.Type {
		case tFOR:
			if !startLine {
				return vm.Error("preprocess: unexpected for")
			}
			startLine = false

			forErr := vm.PreprocessFor()
			if forErr != nil {
				return forErr
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

func (vm *AutoItVM) PreprocessFor() error {
	startPos := vm.GetPos() - 1
	vm.Log("preprocessor: for start pos: %d %v", startPos, vm.Token())

	tIndex := vm.ReadToken()
	if tIndex.Type != tVARIABLE {
		return vm.Error("preprocessor: expected index variable after for, instead got: %v", tIndex)
	}

	tEquals := vm.ReadToken()
	if tEquals.Type != tOP || tEquals.String() != "=" {
		return vm.Error("preprocessor: expected equals expression after for index variable, instead got: %v", tEquals)
	}

	tStartIndex := make([]*Token, 0)
	for {
		tIndex := vm.ReadToken()
		done := false
		switch tIndex.Type {
		case tTO:
			if len(tStartIndex) == 0 {
				return vm.Error("preprocessor: missing start index in for expression")
			}
			done = true
			break
		case tCOMMENT, tEOL, tSTEP, tFOR:
			return vm.Error("preprocessor: unexpected token during for start index expression: %v", tIndex)
		default:
			tStartIndex = append(tStartIndex, tIndex)
		}
		if done {
			break
		}
	}

	tEndIndex := make([]*Token, 0)
	exprStep := false
	for {
		tIndex := vm.ReadToken()
		done := false
		switch tIndex.Type {
		case tSTEP:
			if len(tEndIndex) == 0 {
				return vm.Error("preprocessor: step expression declared during for expression, but missing end index")
			}
			exprStep = true
			done = true
			break
		case tCOMMENT:
			continue
		case tEOL:
			done = true
			break
		case tTO, tFOR:
			return vm.Error("preprocessor: unexpected token during for end index expression: %v", tIndex)
		default:
			tEndIndex = append(tEndIndex, tIndex)
		}
		if done {
			break
		}
	}

	tStepIndex := make([]*Token, 0)
	if exprStep {
		for {
			tIndex := vm.ReadToken()
			done := false
			switch tIndex.Type {
			case tSTEP, tFOR, tTO:
				return vm.Error("preprocessor: unexpected token during for step index expression: %v", tIndex)
			case tCOMMENT:
				continue
			case tEOL:
				done = true
				break
			default:
				tStepIndex = append(tStepIndex, tIndex)
			}
			if done {
				break
			}
		}
	}

	tForBlock := make([]*Token, 0)
	for {
		tIndex := vm.ReadToken()
		done := false
		switch tIndex.Type {
		case tFOR:
			vm.Log("preprocessor: for: encountered nested for statement")
			forErr := vm.PreprocessFor()
			if forErr != nil {
				return forErr
			}
			tForBlock = append(tForBlock, vm.GetToken(vm.GetPos()-1))
		case tNEXT:
			done = true
			break
		default:
			tForBlock = append(tForBlock, tIndex)
		}
		if done {
			break
		}
	}

	tEoL := vm.ReadToken()
	if tEoL.Type != tEOL && tEoL.Type != tCOMMENT {
		return vm.Error("preprocessor: unexpected token following end of for ... next expression: %v", tEoL)
	}
	vm.Move(-1)

	forCall := &ForCall{Index: tIndex, Start: tStartIndex, End: tEndIndex, Step: tStepIndex, Block: tForBlock}
	tHandle := vm.AddHandle(forCall)
	vm.SetToken(startPos, NewToken(tFOR, tHandle.Handle()))
	vm.Log("preprocess: for statement preloaded successfully: %s = %v", tHandle.String(), *forCall)

	endPos := vm.GetPos()
	if endPos >= len(vm.tokens) {
		//vm.Log("call end pos greater than end")
	}
	vm.Log("preprocessor: for end pos: %d %v", endPos, vm.GetToken(endPos))
	vm.RemoveTokens(startPos+1, endPos)
	return nil
}

type ForCall struct {
	Index *Token
	Start []*Token
	End []*Token
	Step []*Token
	Block []*Token
}
func (fc *ForCall) Run(vm *AutoItVM) error {
	vm.Log("FOR: START")
	startIndex, _, err := NewEvaluator(vm, fc.Start).Eval(true)
	if err != nil {
		return err
	}
	si := startIndex.Int64()
	vm.Log("FOR: SI: %d", si)

	endIndex, _, err := NewEvaluator(vm, fc.End).Eval(true)
	if err != nil {
		return err
	}
	ei := endIndex.Int64()
	vm.Log("FOR: EI: %d", ei)

	for i := si; ; i++ {
		//Calculate the next index step ahead of time in case of infinite loop
		newI := i
		if len(fc.Step) > 0 {
			step := make([]*Token, 1)
			step [0] = NewToken(tNUMBER, i)
			step = append(step, fc.Step...)
			tI, _, err := NewEvaluator(vm, step).Eval(true)
			if err != nil {
				return err
			}
			newI = tI.Int64()
			if newI == i {
				break
			}
		}
		vm.Log("FOR: index:%d newIndex:%d", i, newI)

		vmFor, _ := vm.ExtendVM(fc.Block, false)
		vmFor.SetVariable(fc.Index.String(), NewToken(tNUMBER, i))
		vmForErr := vmFor.Run()
		if vmForErr != nil {
			return vmForErr
		}

		if newI < i && newI < ei {
			break
		}
		if newI > i && newI > ei {
			break
		}
		if newI < si || newI >= ei {
			break
		}
		i = newI
	}
	vm.Log("FOR: DONE")
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
	vm.Log("call start pos: %d %v", startPos, vm.Token())

	tCall := vm.ReadToken()
	if tCall.Type != tCALL {
		return vm.Error("preprocessor: expected call, instead got: %v", tCall)
	}

	isStd := false
	if _, exists := stdFunctions[strings.ToLower(tCall.String())]; exists {
		vm.Log("preprocess: setting call to std")
		vm.SetToken(startPos, NewToken(tCALL, tCall.String()))
		isStd = true
	} else {
		vm.Log("preprocess: setting call to udf")
		vm.SetToken(startPos, NewToken(tUDF, tCall.String()))
	}

	tStart := vm.ReadToken()
	if tStart.Type != tLEFTPAREN {
		//return vm.Error("preprocessor: expected block after func call, instead got: %v", tStart)

		//We don't need to process a call block, it's (hopefully) being used literally
		vm.Move(-1)
		return nil
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
		vm.Log("preprocessor: call: depth %d step %v %s", depth, blockToken.Type, blockToken)
		switch blockToken.Type {
		case tCALL:
			vm.Log("preprocessor: call: encountered nested call")
			vm.Move(-1)
			callErr := vm.PreprocessFuncCall()
			if callErr != nil {
				return callErr
			}
			callBlock = append(callBlock, vm.GetToken(vm.GetPos()-1))
		case tSEPARATOR:
			if depth > 0 {
				return vm.Error("preprocessor: unexpected separator in nested func call block: %v", blockToken)
			}
			callBlocks = append(callBlocks, callBlock)
			callBlock = make([]*Token, 0)
			vm.Log("preprocess: call: (,) separator encountered so moving to next block")
		case tLEFTPAREN:
			depth++
			vm.Log("preprocess: call: depth: %d", depth)
		case tRIGHTPAREN:
			depth--
			vm.Log("preprocess: call: depth: %d", depth)
		default:
			callBlock = append(callBlock, blockToken)
			//vm.Log("preprocess: call: added to callBlock: %v", *blockToken)
		}
	}
	if len(callBlock) > 0 {
		callBlocks = append(callBlocks, callBlock)
	}

	functionCall := &FunctionCall{Name: tCall.String(), Block: callBlocks}
	tHandle := vm.AddHandle(functionCall)
	if isStd {
		vm.SetToken(startPos, NewToken(tCALL, tHandle.Handle()))
	} else {
		vm.SetToken(startPos, NewToken(tUDF, tHandle.Handle()))
	}
	vm.Log("preprocess: call preloaded successfully: %s -> %s = %v", tCall.String(), tHandle.String(), callBlocks)

	endPos := vm.GetPos()
	if endPos >= len(vm.tokens) {
		//vm.Log("call end pos greater than end")
	}
	vm.Log("call end pos: %d %v", endPos, vm.GetToken(endPos))
	vm.RemoveTokens(startPos+1, endPos)
	return nil
}