package autoit

import (
	"fmt"
	"math"
	"os"
	"strings"
)

func init() {
	//Type conversions
	stdFunctions["binary"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "expression"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			return NewToken(tBINARY, args["expression"].Bytes()), nil
		},
	}
	stdFunctions["number"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "expression"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			return NewToken(tDOUBLE, args["expression"].Float64()), nil
		},
	}
	stdFunctions["string"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "expression"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			return NewToken(tSTRING, string(args["expression"].Bytes())), nil
		},
	}

	//Type checks
	stdFunctions["vargettype"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "expression"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			varType := string(args["expression"].Type)
			switch args["expression"].Type {
			case tBINARY:
				varType = "Binary"
			case tNULL:
				varType = "Keyword"
			case tBOOLEAN:
				varType = "Bool"
			case tSTRING:
				varType = "String"
			case tDEFAULT:
				varType = "Keyword"
			}
			return NewToken(tSTRING, strings.Title(varType)), nil
		},
	}

	//Debugging
	stdFunctions["consolewrite"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "text"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			fmt.Fprint(os.Stdout, args["text"].String())
			vm.stdout += args["text"].String()
			return nil, nil
		},
	}
	stdFunctions["consolewriteerror"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "error"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			fmt.Fprint(os.Stderr, args["error"].String())
			vm.stderr += args["error"].String()
			return nil, nil
		},
	}
	stdFunctions["seterror"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "code"},
			&FunctionArg{Name: "extended", DefaultValue: NewToken(tNUMBER, 0)},
			&FunctionArg{Name: "returnValue", DefaultValue: NewToken(tNUMBER, 0)},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			vm.error = args["code"].Int()
			vm.extended = args["extended"].Int()
			vm.returnValue = args["returnValue"]
			return vm.returnValue, nil
		},
	}
	stdFunctions["setextended"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "extended"},
			&FunctionArg{Name: "returnValue", DefaultValue: NewToken(tNUMBER, 0)},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			vm.extended = args["extended"].Int()
			vm.returnValue = args["returnValue"]
			return vm.returnValue, nil
		},
	}
}

//Impl based on https://stackoverflow.com/a/35164011
//Determines whether a given number has a decimal point value
func isWholeNumber(num float64) bool {
	if num == math.Trunc(num) {
		return true
	}
	return false
	/*
	const epsilon = 1e-9 //Margin of error
	if _, frac := math.Modf(math.Abs(num)); frac < epsilon || frac > 1.0 - epsilon {
		return false
	}
	return true
	*/
}