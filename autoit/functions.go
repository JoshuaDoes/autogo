package autoit

import (
	"fmt"
	"os"
)

var (
	stdFunctions = map[string]*Function{
		"ConsoleWrite": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "sMsg"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				fmt.Fprint(os.Stdout, args["sMsg"].Data)
				return nil, nil
			},
		},
		"ConsoleWriteError": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "sError"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				fmt.Fprint(os.Stderr, args["sError"].Data)
				return nil, nil
			},
		},
		/*"ConsoleWriteLine": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "sMsg"},
			},
			Block: []*Token{
				NewToken(tCALL, "ConsoleWrite"),
				NewToken(tBLOCK, ""),
				NewToken(tVARIABLE, "sMsg"),
				NewToken(tOP, "&"),
				NewToken(tMACRO, "CRLF"),
				NewToken(tBLOCKEND, ""),
			},
		},*/
	}
)

type Function struct {
	Args []*FunctionArg                                     //Ordered list of arguments for calls
	Func func(*AutoItVM, map[string]*Token) (*Token, error) //Stores a Go func binding for calls
	Block []*Token                                          //Stores a token block to step on calls
}

type FunctionArg struct {
	Name string         //Accessed by function block as $Name
	DefaultValue *Token //leave nil to require
}

func (vm *AutoItVM) HandleFunc(funcName string, args []*Token) (*Token, error) {
	function, exists := stdFunctions[funcName]
	if !exists {
		return nil, vm.Error("undefined function %s", funcName)
	}

	if len(args) > len(function.Args) {
		return nil, vm.Error("%s(%d) called with too many args (%d)", funcName, len(function.Args), len(args))
	}

	funcArgs := make(map[string]*Token)
	minimumArgs := len(function.Args)
	for i := 0; i < len(function.Args); i++ {
		if i < len(args) {
			funcArgs[function.Args[i].Name] = args[i]
		} else {
			if function.Args[i].DefaultValue != nil && minimumArgs == len(function.Args) {
				minimumArgs = i-1
			}
			funcArgs[function.Args[i].Name] = function.Args[i].DefaultValue
		}
	}
	if len(args) < minimumArgs {
		return nil, vm.Error("%s(%d) called with less than required args (%d/%d)", funcName, len(function.Args), len(args), minimumArgs)
	}

	if function.Func != nil {
		return function.Func(vm, funcArgs)
	}
	return nil, vm.Error("no handler for function %s", funcName)
}