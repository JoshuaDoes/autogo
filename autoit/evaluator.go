package autoit

import (
	"fmt"
	"strconv"
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
		switch tOp.Data {
		case "&":
			tValue, tRead, err := NewEvaluator(e.vm, []*Token{e.tokens[e.pos]}).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("could not append value: %v", err)
			}
			if tValue == nil {
				return nil, e.error("could not append nil value")
			}

			tDest := NewToken(tSTRING, tSource.Data)
			tDest.Data += tValue.Data
			e.vm.Log("append: %v", *tDest)
			return e.mergeValue(tDest)
		case "+":
			tValue, tRead, err := NewEvaluator(e.vm, []*Token{e.tokens[e.pos]}).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("error getting value to sum: %v", err)
			}

			tDest := NewToken(tNUMBER, strconv.Itoa(tSource.Int() + tValue.Int()))
			e.vm.Log("sum: %v", *tDest)
			return e.mergeValue(tDest)
		case "-":
			tValue, tRead, err := NewEvaluator(e.vm, []*Token{e.tokens[e.pos]}).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("error getting value to subtract: %v", err)
			}

			tDest := NewToken(tNUMBER, strconv.Itoa(tSource.Int() - tValue.Int()))
			e.vm.Log("subtract: %v", *tDest)
			return e.mergeValue(tDest)
		case "*":
			tValue, tRead, err := NewEvaluator(e.vm, []*Token{e.tokens[e.pos]}).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("error getting value to multiply: %v", err)
			}

			tDest := NewToken(tNUMBER, strconv.Itoa(tSource.Int() * tValue.Int()))
			e.vm.Log("multiply: %v", *tDest)
			return e.mergeValue(tDest)
		default:
			e.move(-1)
			return nil, e.error("illegal operator following value to merge: %s", tOp.Data)
		}
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
	case tSTRING:
		if !expectValue {
			return nil, e.pos, e.error("illegal string when not expecting value")
		}

		tValue, err := e.mergeValue(tEval)
		if err != nil {
			return nil, e.pos, err
		}
		return tValue, e.pos, nil
	case tNUMBER:
		if !expectValue {
			return nil, e.pos, e.error("illegal number when not expecting value")
		}

		tValue, err := e.mergeValue(tEval)
		if err != nil {
			return nil, e.pos, err
		}
		return tValue, e.pos, nil
	case tSCOPE:
		if expectValue {
			return nil, e.pos, e.error("illegal variable declaration when expecting value")
		}
		return nil, e.pos, e.error("scope not implemented")
	case tVARIABLE:
		if expectValue {
			e.vm.Log("expecting value for: %v", *tEval)
			tVariable := e.vm.GetVariable(tEval.Data)
			if tVariable == nil {
				return nil, e.pos, e.error("undeclared global variable $%s", tEval.Data)
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
			return e.vm.GetVariable(tEval.Data), e.pos, nil
		}

		switch tOp.Type {
		case tOP:
			switch tOp.Data {
			case "=":
				if expectValue {
					return nil, e.pos, e.error("illegal variable declaration when expecting value")
				}

				tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
				if err != nil {
					return nil, e.pos+tRead, e.error("error getting value for variable declaration: %v", err)
				}

				e.vm.SetVariable(tEval.Data, tValue)
				e.vm.Log("$%s = %v", tEval.Data, *tValue)
				return nil, e.pos+tRead, nil
			default:
				e.move(-1)
				return nil, e.pos, e.error("illegal operator following variable: %s", tOp.Data)
			}
		case tEOL, tBLOCKEND:
			e.move(-1)
			return e.vm.GetVariable(tEval.Data), e.pos, nil //Return the evaluated value
		default:
			e.move(-1)
			return nil, e.pos, e.error("illegal token following variable: %v", *tOp)
		}
	case tMACRO:
		tValue, err := e.vm.GetMacro(tEval.Data)
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

		tValue, err := e.vm.HandleFunc(tEval.Data, callParams)
		return tValue, e.pos, err
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
	depth := -1

	for {
		token := e.readToken()
		if token == nil {
			e.vm.Log("BLOCK REACHED EOF")
			return block, nil
		}

		switch token.Type {
		case tBLOCK:
			depth++
			if depth == 0 {
				depth = 1
			}
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