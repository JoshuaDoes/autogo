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
	switch tSource.Type {
	case tMACRO:
		tDest, err := e.vm.GetMacro(tSource.String())
		e.vm.Log("macro: %v", *tDest)
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
	}

	tOp := e.readToken()
	if tOp != nil {
		tSourceValue, _, err := NewEvaluator(e.vm, []*Token{tSource}).Eval(true)
		if err != nil {
			return nil, err
		}

		switch tOp.Type {
		case tOP:
			switch tOp.String() {
			case "&":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("could not append value: %v", err)
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
					return nil, e.error("error getting value to sum: %v", err)
				}

				tDest := NewToken(tNUMBER, tSourceValue.Float64() + tValue.Float64())
				e.vm.Log("sum: %v", *tDest)
				return e.mergeValue(tDest)
			case "-":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("error getting value to subtract: %v", err)
				}

				tDest := NewToken(tNUMBER, tSourceValue.Float64() - tValue.Float64())
				e.vm.Log("subtract: %v", *tDest)
				return e.mergeValue(tDest)
			case "*":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("error getting value to multiply: %v", err)
				}

				tDest := NewToken(tNUMBER, tSourceValue.Float64() * tValue.Float64())
				e.vm.Log("multiply: %v", *tDest)
				return e.mergeValue(tDest)
			case "/":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("error getting value to divide: %v", err)
				}

				tDest := NewToken(tNUMBER, tSourceValue.Float64() / tValue.Float64())
				e.vm.Log("divide: %v", *tDest)
				return e.mergeValue(tDest)
			case "<":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("error getting value to compare less than: %v", err)
				}

				tDest := NewToken(tBOOLEAN, tSourceValue.Float64() < tValue.Float64())
				e.vm.Log("less than: %v", *tDest)
				return e.mergeValue(tDest)
			case ">":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("error getting value to compare greater than: %v", err)
				}

				tDest := NewToken(tBOOLEAN, tSourceValue.Float64() > tValue.Float64())
				e.vm.Log("greater than: %v", *tDest)
				return e.mergeValue(tDest)
			case "=":
				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				e.move(tRead)
				if err != nil {
					return nil, e.error("error getting value to compare equals: %v", err)
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
				return nil, e.error("error getting value to compare bool: %v", err)
			}

			tDest := NewToken(tBOOLEAN, tSourceValue.Bool() && tValue.Bool())
			e.vm.Log("compare bool: %v", *tDest)
			return e.mergeValue(tDest)
		case tOR:
			tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("error getting value to or bool: %v", err)
			}

			tDest := NewToken(tBOOLEAN, tSourceValue.Bool() || tValue.Bool())
			e.vm.Log("or bool: %v", *tDest)
			return e.mergeValue(tDest)
		case tEOL, tBLOCKEND, tSEPARATOR, tTHEN:
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
	case tBLOCK:
		/*
		block := e.readBlock()
		if block == nil {
			e.move(-1)
			return nil, e.pos, e.error("illegal block")
		}
		*/
		return nil, e.pos, e.error("block not implemented")
	case tSTRING, tNUMBER, tBOOLEAN, tNOT, tBINARY, tMACRO, tHANDLE:
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

			vmIf, preprocess := e.vm.ExtendVM(blockIf)
			if preprocess != nil {
				return nil, e.pos, preprocess
			}
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

		vmIf, preprocess := e.vm.ExtendVM(blockIf)
		if preprocess != nil {
			return nil, e.pos, preprocess
		}
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

		vmSwitch, preprocess := e.vm.ExtendVM(blockSwitch)
		if preprocess != nil {
			return nil, e.pos, preprocess
		}
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
			e.vm.Log("expecting value for: %v", *tEval)
			tVariable := e.vm.GetVariable(tEval.String())
			if tVariable == nil {
				return nil, e.pos, e.error("undeclared global variable $%s", tEval.String())
			}
			tValue, err := e.mergeValue(tVariable)
			if err != nil {
				return nil, e.pos, err
			}
			return tValue, e.pos, nil
		}

		e.move(-1)
		for {
			tVariable := e.readToken()
			if tVariable == nil {
				return nil, e.pos, e.error("expected variable declaration, instead got: %v", *tVariable)
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
						return nil, e.pos+tRead, e.error("error getting value for variable declaration of $%s: %v", tVariable.String(), err)
					}

					e.vm.SetVariable(tVariable.String(), tValue)
					e.vm.Log("$%s = %v:%v", tVariable.String(), *e.tokens[e.pos], *tValue)
					continue
				default:
					e.move(-1)
					return nil, e.pos, e.error("illegal operator following variable$%s: %s", tVariable.String(), tOp.String())
				}
			case tSEPARATOR:
				e.vm.SetVariable(tVariable.String(), NewToken(tSTRING, ""))
				continue
			case tEOL, tBLOCKEND:
				e.move(-1)
				e.vm.SetVariable(tVariable.String(), NewToken(tSTRING, ""))
				e.vm.Log("$%s = \"\"", tVariable.String())
				return nil, e.pos, nil
			default:
				e.move(-1)
				return nil, e.pos, e.error("illegal token following variable $%s: %v", tVariable.String(), *tOp)
			}
		}
	case tCALL:
		callTokens, err := e.readBlock()
		if err != nil {
			return nil, e.pos, err
		}
		if callTokens == nil {
			if !expectValue {
				return nil, e.pos, e.error("call requires parameter block")
			}
			return tEval, e.pos, nil //Return a pointer to the call rather than evaluating its value
		}
		callParams := e.evalBlock(callTokens)
		if len(callParams) > 1 {
			if callParams[len(callParams)-1] == nil || callParams[len(callParams)-1].Type == tBLOCKEND {
				callParams = callParams[:len(callParams)-1] //Strip leftover BLOCKEND from somewhere, TODO: find it
			}
		}

		tValue, err := e.vm.HandleFunc(tEval.String(), callParams)
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
func (e *Evaluator) move(pos int) {
	e.pos += pos
}
func (e *Evaluator) readToken() *Token {
	if e.pos >= len(e.tokens) {
		return nil
	}
	defer e.move(1)
	return e.tokens[e.pos]
}
func (e *Evaluator) readBlock() ([]*Token, error) {
	block := make([]*Token, 0)
	depth := 0

	for {
		token := e.readToken()
		if token == nil {
			e.vm.Log("BLOCK REACHED EOF")
			return block, nil
		}

		switch token.Type {
		case tBLOCK:
			if depth > 0 {
				block = append(block, token)
			}
			depth++
			e.vm.Log("BLOCK+ %d", depth)
		case tBLOCKEND:
			if depth > 0 {
				block = append(block, token)
			}
			depth--
			e.vm.Log("BLOCK- %d", depth)
		case tEOL:
			if depth > 0 {
				for i := 0; i < depth; i++ {
					block = append(block, NewToken(tBLOCKEND, ""))
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
			if depth <= 0 {
				e.move(-1)
				return block, nil
			}

			tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos-1:]).Eval(true)
			e.move(tRead-1)
			if err != nil {
				e.vm.Log("BLOCK %d ERR: %v, %v", depth, *token, err)
				e.move(-1)
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
		case tBLOCK:
			if depth > 0 {
				section = append(section, block[i])
				e.vm.Log("evalBlock %d: appending section with %v", depth, *block[i])
			}
			depth++
			e.vm.Log("depth: %v", depth)
		case tBLOCKEND:
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
	if params != nil {
		return fmt.Errorf("evaluator: " + format, params...)
	}
	return fmt.Errorf("evaluator: " + format)
}