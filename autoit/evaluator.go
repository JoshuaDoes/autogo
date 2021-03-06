package autoit

import (
	"fmt"
)

type Evaluator struct {
	vm *AutoItVM
	tokens []*Token
	pos int
}
func NewEvaluator(vm *AutoItVM, tokens []*Token) *Evaluator {
	return &Evaluator{
		vm: vm,
		tokens: tokens,
	}
}
func (e *Evaluator) mergeValue(tSource *Token) (*Token, error) {
	if tSource == nil {
		return nil, nil
	}
	e.vm.Log("evaluator:mergeValue: tSource: %v", *tSource)
	switch tSource.Type {
	case tMACRO:
		tDest, err := e.vm.GetMacro(tSource.String())
		e.vm.Log("macro: @%s -> %v", tSource.String(), *tDest)
		if err != nil {
			return nil, err
		}
		return e.mergeValue(tDest)
	case tNOT:
		e.vm.Log("not: %v", *tSource)
		tBool, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
		e.move(tRead)
		if err != nil {
			return nil, err
		}
		tDest, err := e.mergeValue(tBool)
		if err != nil {
			return nil, err
		}
		return NewToken(tBOOLEAN, !tDest.Bool()), nil
	case tCALL, tUDF:
		e.vm.Log("call: %v", *tSource)
		functionCall := e.vm.GetHandle(tSource.String())
		if functionCall == nil {
			return tSource, nil
			//return nil, e.error("undefined call attempt, did preprocessing fail?: %v", *tSource)
		}
		tDest, err := e.vm.HandleCall(functionCall.(*FunctionCall))
		if err != nil {
			e.vm.Log("call failed: %v", err)
			return nil, err
		}
		e.vm.Log("call succeeded, merging value: %v", *tDest)
		return e.mergeValue(tDest)
	case tVARIABLE:
		tDest := e.vm.GetVariable(tSource.String())
		if tDest == nil {
			return nil, e.error("undeclared global variable $%s", tSource.String())
		}
		e.vm.Log("got value for $%s: %v", tSource.String(), *tDest)
		return e.mergeValue(tDest)
	}

	tOp := e.readToken()
	if tOp != nil {
		e.vm.Log("getting source value: %v", *tSource)
		tSourceValue, _, err := NewEvaluator(e.vm, []*Token{tSource}).Eval(true)
		if err != nil {
			e.vm.Log("error getting source value: %v %v", *tSource, err)
			return nil, err
		}

		switch tOp.Type {
		case tOP:
			switch tOp.String() {
			case "&":
				e.vm.Log("trying to append from pos %d: %v %v", e.pos, *e.tokens[e.pos], e.tokens[e.pos:])
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("could not append value: %v %v", *tSource, err)
				}
				if tValue == nil {
					return nil, e.error("could not append nil value")
				}

				tDest := NewToken(tSTRING, tSourceValue.String())
				tDest.Data += tValue.String()
				e.vm.Log("append: %v", *tDest)
				return e.mergeValue(tDest)
			case "+":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("no value to sum: %v", err)
				}

				tDest := NewToken(tDOUBLE, tSourceValue.Float64() + tValue.Float64())
				e.vm.Log("sum: %v", *tDest)
				return e.mergeValue(tDest)
			case "-":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("no value to subtract: %v", err)
				}

				tDest := NewToken(tDOUBLE, tSourceValue.Float64() - tValue.Float64())
				e.vm.Log("subtract: %v", *tDest)
				return e.mergeValue(tDest)
			case "*":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("no value to multiply: %v", err)
				}

				tDest := NewToken(tDOUBLE, tSourceValue.Float64() * tValue.Float64())
				e.vm.Log("multiply: %v", *tDest)
				return e.mergeValue(tDest)
			case "/":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("no value to divide: %v", err)
				}

				tDest := NewToken(tDOUBLE, tSourceValue.Float64() / tValue.Float64())
				e.vm.Log("divide: %v", *tDest)
				return e.mergeValue(tDest)
			case "<":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("no value to compare less than: %v", err)
				}

				tDest := NewToken(tBOOLEAN, tSourceValue.Float64() < tValue.Float64())
				e.vm.Log("less than: %v", *tDest)
				return e.mergeValue(tDest)
			case ">":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("no value to compare greater than: %v", err)
				}

				tDest := NewToken(tBOOLEAN, tSourceValue.Float64() > tValue.Float64())
				e.vm.Log("greater than: %v", *tDest)
				return e.mergeValue(tDest)
			case "=":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("no value to compare equals: %v", err)
				}

				tDest := NewToken(tBOOLEAN, tSourceValue.String() == tValue.String())
				e.vm.Log("equals: %v", *tDest)
				return e.mergeValue(tDest)
			default:
				e.move(-1)
				return nil, e.error("illegal operator following value to merge: %s", tOp.String())
			}
		case tAND:
			tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("no value to compare bool: %v", err)
			}

			tDest := NewToken(tBOOLEAN, tSourceValue.Bool() && tValue.Bool())
			e.vm.Log("compare bool: %v", *tDest)
			return e.mergeValue(tDest)
		case tOR:
			tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("no value to or bool: %v", err)
			}

			tDest := NewToken(tBOOLEAN, tSourceValue.Bool() || tValue.Bool())
			e.vm.Log("or bool: %v", *tDest)
			return e.mergeValue(tDest)
		case tLEFTBRACK:
			e.move(-1)
			mapTokens, err := e.readBlock(tLEFTBRACK, tRIGHTBRACK)
			if err != nil {
				return nil, err
			}
			if len(mapTokens) != 1 {
				return nil, e.error("expected map key for accessing %v", *tSource)
			}

			return e.vm.MapGet(tSource.Handle(), mapTokens[0].String()), nil
		case tEOL, tRIGHTPAREN, tRIGHTBRACK, tSEPARATOR, tTHEN:
			e.move(-1)
			return tSource, nil
		default:
			e.move(-1)
			return nil, e.error("illegal token following value to merge: %v", *tOp)
		}
	}

	return tSource, nil
}
func (e *Evaluator) Eval(expectValue bool) (*Token, int, error) {
	if len(e.tokens) == 0 {
		return nil, 0, nil
	}

	tEval := e.readToken()
	e.vm.Log("evaluating %v", *tEval)
	switch tEval.Type {
	case tLEFTPAREN:
		/*
		block := e.readBlock(tLEFTPAREN, tRIGHTPAREN)
		if block == nil {
			e.move(-1)
			return nil, e.pos, e.error("illegal block")
		}
		*/
		return nil, e.pos, e.error("block not implemented")
		//return nil, e.pos, e.error("expect value: %v", expectValue)
	case tEXTEND:
		if !expectValue {
			return nil, e.pos, e.error("unexpected value")
		}

		//Make sure we have a comment and/or end of line next, no other reason to use _ in a script
		for i := e.pos; i < len(e.tokens); i++ {
			tEoL := e.readToken()
			e.vm.Log("tEoL test %d: (%s) %v", i, tEoL.Type, tEoL)
			valid := false
			switch tEoL.Type {
				case tEOL:
					valid = true
					break
				case tCOMMENT:
					continue
				default:
					return nil, e.pos, e.error("expected end of line following extend, instead got: %v", tEoL)
			}
			if valid {
				break
			}
		}

		//tValue, err := e.mergeValue(tEval)
		//if err != nil {
		//	return nil, e.pos, e.error("no value to extend: %v", err)
		//}
		tValue := e.readToken()
		return tValue, e.pos, nil
		//tDest, err := e.mergeValue(tValue)
		//return tDest, e.pos, err
	case tDEFAULT:
		if !expectValue {
			return nil, e.pos, e.error("unexpected value")
		}
		return tEval, e.pos, nil
	case tSTRING, tNUMBER, tDOUBLE, tBOOLEAN, tNOT, tNULL, tBINARY, tMACRO, tHANDLE:
		if !expectValue {
			return nil, e.pos, e.error("unexpected value")
		}

		tValue, err := e.mergeValue(tEval)
		if err != nil {
			return nil, e.pos, err
		}
		return tValue, e.pos, nil
	case tIF, tELSEIF:
		if expectValue {
			return nil, e.pos, e.error("illegal if condition when expecting value")
		}

		tBool := e.readToken()
		if tBool == nil {
			return nil, e.pos, e.error("expected bool for not")
		}

		tValue, err := e.mergeValue(tBool)
		if err != nil {
			return nil, e.pos, err
		}

		tThen := e.readToken()
		if tThen == nil || tThen.Type != tTHEN {
			return nil, e.pos, e.error("expected then after if condition, instead got: %v", tThen)
		}

		blockIf := make([]*Token, 0)
		inLine := true
		depth := 0
		for {
			tBlock := e.readToken()
			if tBlock == nil {
				return nil, e.pos, e.error("expected end of if statement")
			}
			e.vm.Log("if: %v", *tBlock)

			endBlock := false
			switch tBlock.Type {
			case tIF:
				if inLine {
					return nil, e.pos, e.error("unexpected inline if")
				}
				depth++
				blockIf = append(blockIf, tBlock)
			case tELSE, tELSEIF:
				if inLine {
					return nil, e.pos, e.error("unexpected inline else/elseif")
				}
				if depth == 0 {
					endBlock = true
					break
				}
				blockIf = append(blockIf, tBlock)
			case tIFEND:
				if inLine {
					return nil, e.pos, e.error("unexpected inline endif")
				}
				if depth == 0 {
					endBlock = true
					break
				}
				depth--
				blockIf = append(blockIf, tBlock)
			default:
				if inLine && len(blockIf) == 0 {
					inLine = false
				}
				blockIf = append(blockIf, tBlock)
			}
			if endBlock {
				e.move(-2) //ifblock
				break
			}
		}

		e.vm.Log("bool value: %v", tValue)
		if tValue.Bool() {
			if e.vm.ranIfStatement {
				return nil, e.pos, nil
			}
			e.vm.ranIfStatement = true

			vmIf, _ := e.vm.ExtendVM(blockIf, false)
			/*if preprocess != nil {
				return nil, e.pos, preprocess
			}*/
			vmIfErr := vmIf.Run()
			if vmIfErr != nil {
				return nil, e.pos, vmIfErr
			}

			//Read out the rest of the if statement
			depth := 0
			for {
				tBlock := e.readToken()
				if tBlock == nil {
					return nil, e.pos, e.error("expected end of %s statement", tEval.Type)
				}

				endBlock := false
				switch tBlock.Type {
				case tIF:
					depth++
				case tIFEND:
					if depth == 0 {
						endBlock = true
						break
					}
					depth--
				}
				if endBlock {
					e.move(-2)
					break
				}
			}

			return nil, e.pos, nil
		}
		return nil, e.pos, nil
	case tELSE:
		if expectValue {
			return nil, e.pos, e.error("illegal else condition when expecting value")
		}

		blockIf := make([]*Token, 0)
		depth := 0
		for {
			tBlock := e.readToken()
			if tBlock == nil {
				return nil, e.pos, e.error("expected end of else statement")
			}

			endBlock := false
			switch tBlock.Type {
			case tIF:
				depth++
				blockIf = append(blockIf, tBlock)
			case tELSE, tELSEIF:
				if depth == 0 {
					return nil, e.pos, e.error("unexpected else after else")
				}
				blockIf = append(blockIf, tBlock)
			case tIFEND:
				if depth == 0 {
					endBlock = true
					break
				}
				depth--
				blockIf = append(blockIf, tBlock)
			default:
				blockIf = append(blockIf, tBlock)
			}
			if endBlock {
				e.move(-2) //ifblock
				break
			}
		}

		if e.vm.ranIfStatement {
			return nil, e.pos, nil //e.error("already ran else statement")
		}
		e.vm.ranIfStatement = true

		vmIf, _ := e.vm.ExtendVM(blockIf, false)
		/*if preprocess != nil {
			return nil, e.pos, preprocess
		}*/
		vmIfErr := vmIf.Run()
		if vmIfErr != nil {
			return nil, e.pos, vmIfErr
		}

		return nil, e.pos, nil
	case tTHEN:
		e.move(-1)
		return nil, e.pos, nil
	case tIFEND:
		e.vm.ranIfStatement = false
		return nil, e.pos, nil
	case tSWITCH:
		if expectValue {
			return nil, e.pos, e.error("illegal switch condition when expecting value")
		}

		tSourceValue := e.readToken()
		if tSourceValue == nil {
			return nil, e.pos, e.error("expected value to switch on")
		}
		e.vm.Log("source value: %v", *tSourceValue)

		tSwitchValue, err := e.mergeValue(tSourceValue)
		if err != nil {
			return nil, e.pos, err
		}
		e.vm.Log("switch value: %v", *tSwitchValue)

		blockSwitch := make([]*Token, 0)
		depth := 0
		addingCase := false
		evalCase := false
		for {
			tBlock := e.readToken()
			if tBlock == nil {
				return nil, e.pos, e.error("expected end of switch statement")
			}
			e.vm.Log("switch (%v): %v", addingCase, *tBlock)

			endBlock := false
			switch tBlock.Type {
			case tSWITCH:
				depth++
				if addingCase {
					blockSwitch = append(blockSwitch, tBlock)
				}
			case tCASE, tSEPARATOR:
				if evalCase {
					addingCase = false
					continue
				}

				if depth > 0 {
					if addingCase {
						blockSwitch = append(blockSwitch, tBlock)
					}
					continue
				}
				addingCase = false

				tCaseSourceValue := e.readToken()
				if tCaseSourceValue == nil {
					return nil, e.pos, e.error("expected case value and end of switch statement")
				}
				if tCaseSourceValue.Type == tELSE {
					addingCase = true
					evalCase = true
				} else {
					tCaseValue, err := e.mergeValue(tCaseSourceValue)
					if err != nil {
						return nil, e.pos, err
					}
					if tSwitchValue.String() == tCaseValue.String() {
						addingCase = true
						evalCase = true
					}

					for {
						tSeparator := e.readToken()
						if tSeparator != nil && tSeparator.Type == tSEPARATOR {
							//Make sure we only evaluate values once
							if evalCase {
								e.move(1)
								continue
							}

							//Check next case value after separator
							tCaseSourceValue = e.readToken()
							if tCaseSourceValue == nil {
								return nil, e.pos, e.error("expected value after case separator")
							}
							if tCaseSourceValue.Type == tELSE {
								return nil, e.pos, e.error("unexpected else after case separator")
							}
							tCaseValue, err = e.mergeValue(tCaseSourceValue)
							if err != nil {
								return nil, e.pos, err
							}
							if tSwitchValue.String() == tCaseValue.String() {
								addingCase = true
								evalCase = true
							}
						} else {
							e.move(-1)
							break
						}
					}
				}
			case tSWITCHEND:
				if depth == 0 {
					endBlock = true
					break
				}
				depth--
				if addingCase {
					blockSwitch = append(blockSwitch, tBlock)
				}
			default:
				if addingCase {
					blockSwitch = append(blockSwitch, tBlock)
				}
			}
			if endBlock {
				break
			}
		}

		e.vm.Log("Final switch block: %v", blockSwitch)

		vmSwitch, _ := e.vm.ExtendVM(blockSwitch, false)
		/*if preprocess != nil {
			return nil, e.pos, preprocess
		}*/
		vmSwitchErr := vmSwitch.Run()
		if vmSwitchErr != nil {
			return nil, e.pos, vmSwitchErr
		}

		return nil, e.pos, nil
	case tSCOPE:
		if expectValue {
			return nil, e.pos, e.error("illegal variable declaration when expecting value")
		}

		vm := e.vm
		switch tEval.String() {
		case "local":
			vm = e.vm
		case "global":
			if e.vm.parentScope != nil {
				vm = e.vm.parentScope
			}
		default:
			return nil, e.pos, e.error("illegal scope: %s", tEval.String())
		}

		_, tRead, err := NewEvaluator(vm, e.tokens[e.pos:]).Eval(false)
		e.move(tRead)
		if err != nil {
			return nil, e.pos, err
		}
		return nil, e.pos, nil
	case tFUNC:
		if expectValue {
			return nil, e.pos, e.error("illegal func declaration when expecting value")
		}

		for {
			token := e.readToken()
			if token == nil || token.Type == tFUNCEND {
				break
			}
		}
		return nil, e.pos, nil
	case tVARIABLE:
		if expectValue {
			tValue, err := e.mergeValue(tEval)
			return tValue, e.pos, err
		}

		e.move(-1)
		for {
			tVariable := e.readToken()
			if tVariable == nil {
				return nil, e.pos, nil
			}
			if tVariable.Type == tEOL {
				e.move(-1)
				return nil, e.pos, nil
			}
			if tVariable.Type == tSEPARATOR {
				continue
			}
			e.vm.Log("tVariable: %v", *tVariable)

			tOp := e.readToken()
			if tOp == nil {
				e.move(-1)
				e.vm.SetVariable(tVariable.String(), NewToken(tSTRING, ""))
				e.vm.Log("$%s = \"\"", tVariable.String())
				return nil, e.pos, nil
			}
			e.vm.Log("tOp: %v", *tOp)

			switch tOp.Type {
			case tOP:
				switch tOp.String() {
				case "=":
					if expectValue {
						return nil, e.pos, e.error("illegal variable declaration when expecting value")
					}

					tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
					e.move(tRead)
					if err != nil {
						return nil, e.pos+tRead, e.error("no value for variable declaration of $%s: %v", tVariable.String(), err)
					}

					e.vm.SetVariable(tVariable.String(), tValue)
					e.vm.Log("$%s = %v:%v", tVariable.String(), *e.tokens[e.pos], *tValue)
					continue
				default:
					return nil, e.pos, e.error("illegal operator following variable $%s: %s", tVariable.String(), tOp.String())
				}
			case tSEPARATOR:
				e.vm.SetVariable(tVariable.String(), NewToken(tSTRING, ""))
				continue
			case tLEFTBRACK:
				e.move(-1)
				mapTokens, err := e.readBlock(tLEFTBRACK, tRIGHTBRACK)
				if err != nil {
					return nil, e.pos, err
				}

				if len(mapTokens) == 0 {
					tEndLine := e.readToken()
					if tEndLine == nil || tEndLine.Type == tEOL {
						handle := e.vm.AddHandle(make(map[string]*Token))
						e.vm.SetVariable(tVariable.String(), handle)
						
						if tEndLine != nil && tEndLine.Type == tEOL {
							e.move(-1)
						}
						return nil, e.pos, nil
					}
				}

				if len(mapTokens) == 1 {
					tEndLine := e.readToken()
					if tEndLine == nil || tEndLine.Type == tEOL {
						handle := e.vm.AddHandle(make([]*Token, mapTokens[0].Int64()))
						e.vm.SetVariable(tVariable.String(), handle)

						if tEndLine != nil && tEndLine.Type == tEOL {
							e.move(-1)
						}
						return nil, e.pos, nil
					}

					switch tEndLine.Type {
					case tOP:
						switch tEndLine.String() {
						case "=":
							if expectValue {
								return nil, e.pos, e.error("illegal variable declaration when expecting value")
							}

							tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
							e.move(tRead)
							if err != nil {
								return nil, e.pos+tRead, e.error("no value for variable declaration of $%s: %v", tVariable.String(), err)
							}

							//e.vm.SetVariable(tVariable.String(), tValue)
							//e.vm.Log("$%s = %v:%v", tVariable.String(), *e.tokens[e.pos], *tValue)
							e.vm.MapSet(e.vm.GetVariable(tVariable.String()).Handle(), mapTokens[0].String(), tValue)
							e.vm.Log("$%s[%s] = %v", tVariable.String(), mapTokens[0].String(), *tValue)
							continue
						default:
							return nil, e.pos, e.error("illegal operator following variable $%s: %s", tVariable.String(), tEndLine.String())
						}
					default:
						return nil, e.pos, e.error("illegal token following variable $%s: %v", tVariable.String(), *tEndLine)
					}

					switch tEndLine.Type {
					case tOP:
						if tEndLine.String() != "=" {
							return nil, e.pos, e.error("unexpected operator following $%s: %s", tVariable.String(), tEndLine.String())
						}
					}

					return nil, e.pos, e.error("expected end of line following array declaration of $%s (predefined arrays not yet supported), instead got: %v", tVariable.String(), *tEndLine)
				}

				return nil, e.pos, e.error("unexpected comma after array count for $%s", tVariable.String())
			case tEOL, tRIGHTPAREN:
				e.move(-1)
				e.vm.SetVariable(tVariable.String(), NewToken(tSTRING, ""))
				e.vm.Log("$%s = \"\"", tVariable.String())
				return nil, e.pos, nil
			default:
				e.move(-1)
				return nil, e.pos, e.error("illegal token following variable $%s: %v", tVariable.String(), *tOp)
			}
		}
	case tFOR:
		if expectValue {
			return nil, e.pos, e.error("illegal for declaration when expecting value")
		}

		forHandle := e.vm.GetHandle(tEval.String())
		if forHandle == nil {
			return nil, e.pos, e.error("undefined for attempt, did preprocessing fail?: %v", *tEval)
		}
		forCall := forHandle.(*ForCall)

		forErr := forCall.Run(e.vm)
		if forErr != nil {
			e.vm.Log("for call error: %v", forErr)
		}
		return nil, e.pos, nil
	case tCALL, tUDF:
		e.vm.Log("tCALL: %s", tEval.String())
		functionCall := e.vm.GetHandle(tEval.String())
		if functionCall == nil {
			if expectValue {
				return tEval, e.pos, nil
			}
			return nil, e.pos, e.error("undefined call attempt, did preprocessing fail?: %v", *tEval)
		}
		funcCall := functionCall.(*FunctionCall)

		funcCallBlock := ""
		for i := 0; i < len(funcCall.Block); i++ {
			for j := 0; j < len(funcCall.Block[i]); j++ {
				funcCallBlock += fmt.Sprintf("funcCallBlock: %d:%d %s %v\n", i, j, funcCall.Block[i][j].Type, funcCall.Block[i][j])
			}
		}
		e.vm.Log("root call: %s -- \n%s--", funcCall.Name, funcCallBlock)
		
		tValue, err := e.vm.HandleCall(funcCall)
		if err != nil {
			e.vm.Log("root call error: %v", err)
			return nil, e.pos, err
		}
		e.vm.Log("root call merging: %v", tValue)
		tValue, err = e.mergeValue(tValue)
		e.vm.Log("root call merged: %v %v", err, tValue)
		return tValue, e.pos, err
	case tEOL:
		if expectValue {
			return nil, e.pos, e.error("illegal end of line when expecting value")
		}
		return nil, e.pos, nil
	case tCOMMENT:
		return nil, e.pos, nil
	default:
		return nil, e.pos, e.error("token not implemented: %v", *tEval)
	}

	return nil, e.pos, e.error("reached end of eval attempts for token: %v", *tEval)
}
func (e *Evaluator) move(direction int) {
	e.pos += direction
}
func (e *Evaluator) readToken() *Token {
	if e.pos >= len(e.tokens) {
		return nil
	}
	e.move(1)
	if e.tokens[e.pos-1].Type == tCOMMENT {
		return e.readToken()
	}
	return e.tokens[e.pos-1]
}
func (e *Evaluator) readBlock(start, end TokenType) ([]*Token, error) {
	block := make([]*Token, 0)
	depth := 0

	for {
		token := e.readToken()
		if token == nil {
			e.vm.Log("BLOCK REACHED EOF")
			return block, nil
		}

		switch token.Type {
		case start:
			if depth > 0 {
				block = append(block, token)
			}
			depth++
			e.vm.Log("BLOCK+ %d", depth)
		case end:
			depth--
			e.vm.Log("BLOCK- %d", depth)
			if depth == 0 {
				return block, nil
			}
			block = append(block, token)
		case tEOL:
			if depth > 0 {
				for i := 0; i < depth; i++ {
					block = append(block, NewToken(end, ""))
				}
			}

			e.vm.Log("BLOCK REACHED EOL")
			e.move(-1)
			return block, nil
		case tSEPARATOR:
			block = append(block, token)
			e.vm.Log("BLOCK FOUND SEPARATOR")
		default:
			e.vm.Log("BLOCK DEPTH: %d", depth)
			if depth < 0 {
				//e.move(-1)
				return block, nil
			}

			tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos-1:]).Eval(true)
			e.move(tRead-1)
			if err != nil {
				e.vm.Log("BLOCK %d ERR: %v, %v", depth, *token, err)
				return nil, err
			} else {
				e.vm.Log("BLOCK %d: %v = %v", depth, *token, tValue)
				block = append(block, tValue)
			}
		}

		if depth == 0 {
			break
		}
	}
	e.vm.Log("BLOCK TOTAL: %v", block)
	return block, nil
}
func (e *Evaluator) evalBlock(block []*Token) []*Token {
	depth := 0

	split := make([]*Token, 0)
	section := make([]*Token, 0)
	for i := 0; i < len(block); i++ {
		if block[i] == nil {
			continue
		}
		e.vm.Log("------- %d EVALUATING %v", depth, *block[i])
		switch block[i].Type {
		case tLEFTPAREN:
			if depth > 0 {
				section = append(section, block[i])
				e.vm.Log("evalBlock %d: appending section with %v", depth, *block[i])
			}
			depth++
			e.vm.Log("depth: %v", depth)
		case tRIGHTPAREN:
			if depth > 0 {
				section = append(section, block[i])
				e.vm.Log("evalBlock %d: appending section with %v", depth, *block[i])
			}
			depth--
			e.vm.Log("depth: %v", depth)
		case tSEPARATOR:
			e.vm.Log("evalBlock %d: ,", depth)
			tValue, _, err := NewEvaluator(e.vm, section).Eval(true)
			if err != nil {
				e.vm.Log("evalBlock %d: eval error: %v", depth, err)
				return block
			}
			split = append(split, tValue)
			e.vm.Log("evalBlock %d: evaluated as %v", depth, *tValue)
			section = make([]*Token, 0)
		default:
			section = append(section, block[i])
			e.vm.Log("evalBlock %d: appending section with %v", depth, *block[i])
		}

		if depth < 0 {
			break
		}
	}
	if len(section) > 0 {
		tValue, _, err := NewEvaluator(e.vm, section).Eval(true)
		if err != nil {
			return block
		}
		split = append(split, tValue)
		e.vm.Log("evalBlock: evaluated final section as %v", *tValue)
	}

	return split
}
func (e *Evaluator) error(format string, params ...interface{}) error {
	lines := "?@?"
	if e.pos >= 0 && e.pos < len(e.tokens) {
		if e.tokens[e.pos] != nil {
			lines = fmt.Sprintf("%d@%d", e.tokens[e.pos].LineNumber, e.tokens[e.pos].LinePos)
		}
	}

	format = fmt.Sprintf("eval %s:\n- %s", lines, format)
	if params != nil {
		return fmt.Errorf(format, params...)
	}
	return fmt.Errorf(format)
}