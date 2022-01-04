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
	tOp := e.readToken()
	if tOp == nil {
		return tSource, nil
	}

	switch tOp.Type {
	case tOP:
		switch tOp.String() {
		case "&":
			tValue, tRead, err := NewEvaluator(e.vm, []*Token{e.tokens[e.pos]}).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("could not append value: %v", err)
			}
			if tValue == nil {
				return nil, e.error("could not append nil value")
			}

			tDest := NewToken(tSTRING, tSource.String())
			tDest.Data += tValue.String()
			e.vm.Log("append: %v", *tDest)
			return e.mergeValue(tDest)
		case "+":
			tValue, tRead, err := NewEvaluator(e.vm, []*Token{e.tokens[e.pos]}).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("error getting value to sum: %v", err)
			}

			tDest := NewToken(tNUMBER, tSource.Float64() + tValue.Float64())
			e.vm.Log("sum: %v", *tDest)
			return e.mergeValue(tDest)
		case "-":
			tValue, tRead, err := NewEvaluator(e.vm, []*Token{e.tokens[e.pos]}).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("error getting value to subtract: %v", err)
			}

			tDest := NewToken(tNUMBER, tSource.Float64() - tValue.Float64())
			e.vm.Log("subtract: %v", *tDest)
			return e.mergeValue(tDest)
		case "*":
			tValue, tRead, err := NewEvaluator(e.vm, []*Token{e.tokens[e.pos]}).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("error getting value to multiply: %v", err)
			}

			tDest := NewToken(tNUMBER, tSource.Float64() * tValue.Float64())
			e.vm.Log("multiply: %v", *tDest)
			return e.mergeValue(tDest)
		case "/":
			tValue, tRead, err := NewEvaluator(e.vm, []*Token{e.tokens[e.pos]}).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("error getting value to divide: %v", err)
			}

			tDest := NewToken(tNUMBER, tSource.Float64() / tValue.Float64())
			e.vm.Log("divide: %v", *tDest)
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

		tDest := NewToken(tBOOLEAN, tSource.Bool() && tValue.Bool())
		e.vm.Log("compare bool: %v", *tDest)
		return e.mergeValue(tDest)
	case tOR:
		tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
		e.move(tRead)
		if err != nil {
			return nil, e.error("error getting value to or bool: %v", err)
		}

		tDest := NewToken(tBOOLEAN, tSource.Bool() || tValue.Bool())
		e.vm.Log("or bool: %v", *tDest)
		return e.mergeValue(tDest)
	case tEOL, tBLOCKEND:
		e.move(-1)
		return tSource, nil
	case tSEPARATOR:
		e.move(-1)
		return tSource, nil
	default:
		e.move(-1)
		return nil, e.error("illegal token following value to merge: %v", *tOp)
	}
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
	case tSTRING, tNUMBER, tBOOLEAN, tBINARY:
		if !expectValue {
			return nil, e.pos, e.error("unexpected value")
		}

		tValue, err := e.mergeValue(tEval)
		if err != nil {
			return nil, e.pos, err
		}
		return tValue, e.pos, nil
	case tNOT:
		if !expectValue {
			return nil, e.pos, e.error("unexpected not bool")
		}

		tBool := e.readToken()
		if tBool == nil {
			return nil, e.pos, e.error("expected bool for not")
		}

		tValue, err := e.mergeValue(tBool)
		if err != nil {
			return nil, e.pos, err
		}
		return NewToken(tBOOLEAN, !tValue.Bool()), e.pos, nil
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

		tOp := e.readToken()
		if tOp == nil {
			e.move(-1)
			return e.vm.GetVariable(tEval.String()), e.pos, nil
		}

		switch tOp.Type {
		case tOP:
			switch tOp.String() {
			case "=":
				if expectValue {
					return nil, e.pos, e.error("illegal variable declaration when expecting value")
				}

				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				if err != nil {
					return nil, e.pos+tRead, e.error("error getting value for variable declaration: %v", err)
				}

				e.vm.SetVariable(tEval.String(), tValue, true)
				e.vm.Log("$%s = %v", tEval.String(), *tValue)
				return nil, e.pos+tRead, nil
			default:
				e.move(-1)
				return nil, e.pos, e.error("illegal operator following variable: %s", tOp.String())
			}
		case tEOL, tBLOCKEND:
			e.move(-1)
			return e.vm.GetVariable(tEval.String()), e.pos, nil //Return the evaluated value
		default:
			e.move(-1)
			return nil, e.pos, e.error("illegal token following variable: %v", *tOp)
		}
	case tMACRO:
		tValue, err := e.vm.GetMacro(tEval.String())
		if err != nil {
			return nil, e.pos, err
		}
		tValue, err = e.mergeValue(tValue)
		return tValue, e.pos, err
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
			if callParams[len(callParams)-1].Type == tBLOCKEND {
				callParams = callParams[:len(callParams)-1] //Strip leftover BLOCKEND from somewhere, TODO: find it
			}
		}

		tValue, err := e.vm.HandleFunc(tEval.String(), callParams)
		return tValue, e.pos, err
	case tFUNCRETURN:
		tValue := e.readToken()
		if tValue == nil {
			return nil, e.pos, nil
		}
		if tValue.Type == tEOL {
			e.move(-1)
			return nil, e.pos, nil
		}

		tValue, err := e.mergeValue(tValue)
		if err != nil {
			return nil, e.pos, err
		}

		e.vm.returnValue = tValue
		if expectValue {
			return tValue, e.pos, nil
		}
		return nil, e.pos, nil
	case tEOL:
		if expectValue {
			return nil, e.pos, e.error("illegal end of line when expecting value")
		}
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
			depth++
			if depth > 1 {
				block = append(block, token)
			}
			e.vm.Log("BLOCK %d", depth)
		case tBLOCKEND:
			depth--
			if depth > 1 {
				block = append(block, token)
			}
			e.vm.Log("BLOCK %d", depth)
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
			if depth == 0 {
				e.move(-1)
				return block, nil
			}

			tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos-1:]).Eval(true)
			e.move(tRead)
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
		e.vm.Log("------- %d EVALUATING %v", depth, *block[i])
		switch block[i].Type {
		case tBLOCK:
			depth++
			e.vm.Log("depth: %v", depth)
			if depth > 0 {
				section = append(section, block[i])
				e.vm.Log("evalBlock %d: appending section with %v", depth, *block[i])
			}
		case tBLOCKEND:
			depth--
			e.vm.Log("depth: %v", depth)
			if depth > 0 {
				section = append(section, block[i])
				e.vm.Log("evalBlock %d: appending section with %v", depth, *block[i])
			}
		case tSEPARATOR:
			e.vm.Log("evalBlock %d: ,")
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
		return fmt.Errorf("eval: " + format, params...)
	}
	return fmt.Errorf("eval: " + format)
}