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
		switch tOp.Data {
		case "&":
			tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos:]).Eval(true)
			e.move(tRead)
			if err != nil {
				return nil, e.error("error getting value to append: %v", err)
			}

			tDest := NewToken(tSource.Type, tSource.Data)
			tDest.Data += tValue.Data
			return e.mergeValue(tDest)
		default:
			e.move(-1)
			return nil, e.error("illegal operator following value to merge: %s", tOp.Data)
		}
	case tEOL, tBLOCKEND:
		e.move(-1)
		return tSource, nil
	case tSEPARATOR:
		e.move(-2)
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
	case tSCOPE:
		if expectValue {
			return nil, e.pos, e.error("illegal variable declaration when expecting value")
		}
		return nil, e.pos, e.error("scope not implemented")
	case tVARIABLE:
		if expectValue {
			tValue, err := e.mergeValue(e.vm.vars[tEval.Data])
			if err != nil {
				return nil, e.pos, err
			}
			return tValue, e.pos, nil
		}

		tOp := e.readToken()
		if tOp == nil {
			e.move(-1)
			return e.vm.vars[tEval.Data], e.pos, nil
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

				e.vm.vars[tEval.Data] = tValue
				return nil, e.pos+tRead, nil
			default:
				e.move(-1)
				return nil, e.pos, e.error("illegal operator following variable: %s", tOp.Data)
			}
		case tEOL, tBLOCKEND:
			e.move(-1)
			return e.vm.vars[tEval.Data], e.pos, nil //Return the evaluated value
		default:
			e.move(-1)
			return nil, e.pos, e.error("illegal token following variable: %v", *tOp)
		}
	case tMACRO:
		tValue, err := e.vm.GetMacro(tEval.Data)
		return tValue, e.pos, err
	case tCALL:
		callTokens := e.readBlock()
		if callTokens == nil {
			if !expectValue {
				return nil, e.pos, e.error("call requires parameter block")
			}
			return tEval, e.pos, nil //Return a pointer to the call rather than evaluating its value
		}
		callParams := e.evalBlock(callTokens)

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
func (e *Evaluator) readBlock() []*Token {
	block := make([]*Token, 0)
	depth := -1

	for {
		token := e.readToken()
		if token == nil {
			e.vm.Log("BLOCK REACHED EOF")
			return block
		}

		switch token.Type {
		case tBLOCK:
			depth++
			if depth == 0 {
				depth = 1
			}
			block = append(block, token)
			e.vm.Log("BLOCK %d", depth)
		case tBLOCKEND:
			depth--
			block = append(block, token)
			e.vm.Log("BLOCK %d", depth)
		case tEOL:
			if depth > 0 {
				for i := 0; i < depth; i++ {
					block = append(block, NewToken(tBLOCKEND, ""))
				}
			}

			e.vm.Log("BLOCK REACHED EOL")
			e.move(-1)
			return block
		case tSEPARATOR:
			block = append(block, token)
			e.vm.Log("BLOCK FOUND SEPARATOR")
		default:
			if depth == 0 {
				e.move(-1)
				return block
			}

			tValue, tRead, err := NewEvaluator(e.vm, e.tokens[e.pos-1:]).Eval(true)
			e.move(tRead)
			if err != nil {
				e.vm.Log("BLOCK %d ERR: %v, %v", depth, *token, err)
				e.move(-1)
				return nil
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
	return block
}
func (e *Evaluator) evalBlock(block []*Token) []*Token {
	depth := -1

	split := make([]*Token, 0)
	section := make([]*Token, 0)
	for i := 0; i < len(block); i++ {
		switch block[i].Type {
		case tBLOCK:
			depth++
			if depth == 0 {
				depth = 1
			}
			if depth > 1 {
				section = append(section, block[i])
				e.vm.Log("evalBlock: appending section with %v", *block[i])
			}
		case tBLOCKEND:
			depth--
			if depth >= 1 {
				section = append(section, block[i])
				e.vm.Log("evalBlock: appending section with %v", *block[i])
			}
		case tSEPARATOR:
			e.vm.Log("evalBlock: ,")
			tValue, _, err := NewEvaluator(e.vm, section).Eval(true)
			if err != nil {
				e.vm.Log("evalBlock: eval error: %v", err)
				return block
			}
			split = append(split, tValue)
			e.vm.Log("evalBlock: evaluated as %v", *tValue)
			section = make([]*Token, 0)
		default:
			section = append(section, block[i])
			e.vm.Log("evalBlock: appending section with %v", *block[i])
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
		return fmt.Errorf("eval: " + format, params)
	}
	return fmt.Errorf("eval: " + format)
}