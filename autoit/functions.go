package autoit

import (
	"strings"

	//"github.com/sqweek/dialog"
)

var (
	stdFunctions = map[string]*Function{}
)

//Function holds an AutoIt function
type Function struct {
	Args []*FunctionArg                                     //Ordered list of arguments for calls
	Func func(*AutoItVM, map[string]*Token) (*Token, error) //Stores a Go func binding for calls
	Block []*Token                                          //Stores a token block to step on calls
}

//FunctionArg holds an AutoIt function argument
type FunctionArg struct {
	Name string         //Accessed by Function.Block as $Name
	DefaultValue *Token //Leave nil to require a value to be set by the caller 
}

//FunctionCall holds an AutoIt function call
type FunctionCall struct {
	Name string      //Name of the function to call
	Block [][]*Token //List of token blocks to evaluate each argument for the function call
}

func (vm *AutoItVM) GetFunction(fc *FunctionCall) *Function {
	function, exists := stdFunctions[strings.ToLower(fc.Name)]
	if !exists {
		function, exists = vm.funcs[strings.ToLower(fc.Name)]
		if !exists {
			if vm.parentScope != nil {
				function = vm.parentScope.GetFunction(fc)
			}
			if function == nil {
				return nil
			}
		}
	}
	return function
}

func (vm *AutoItVM) HandleCall(fc *FunctionCall) (*Token, error) {
	function := vm.GetFunction(fc)
	if function == nil {
		return nil, vm.Error("undefined function %s", fc.Name)
	}

	if fc.Block == nil {
		fc.Block = make([][]*Token, 0)
	}
	if len(fc.Block) > len(function.Args) {
		return nil, vm.Error("%s(%d) called with too many args (%d)", fc.Name, len(function.Args), len(fc.Block))
	}

	funcArgs := make(map[string]*Token)
	minimumArgs := len(function.Args)
	for i := 0; i < len(function.Args); i++ {
		if i < len(fc.Block) {
			vm.Log("funcArgs %d: evaluating...", i)
			tValue, _, err := NewEvaluator(vm, fc.Block[i]).Eval(true)
			if err != nil {
				return nil, err
			}
			if tValue.Type == tDEFAULT {
				if function.Args[i].DefaultValue != nil {
					tValue = function.Args[i].DefaultValue
				} else {
					return nil, vm.Error("%s has no default value for argument %d/%d", fc.Name, i+1, len(function.Args))
				}
			}
			funcArgs[function.Args[i].Name] = tValue
			vm.Log("funcArgs %d: %s = %v", i, function.Args[i].Name, tValue)
		} else {
			if function.Args[i].DefaultValue != nil && minimumArgs == len(function.Args) {
				minimumArgs = i-1
			}
			funcArgs[function.Args[i].Name] = function.Args[i].DefaultValue
		}
	}
	if len(fc.Block) < minimumArgs {
		return nil, vm.Error("%s(%d) called with less than required args (%d/%d)", fc.Name, len(function.Args), len(fc.Block), minimumArgs)
	}

	if function.Func != nil {
		vm.SetError(0)
		vm.SetExtended(0)
		vm.SetReturnValue(NewToken(tNUMBER, 0))
		return function.Func(vm, funcArgs)
	}
	if function.Block != nil {
		vmFunc, _ := vm.ExtendVM(function.Block, false)
		vmFunc.numParams = len(fc.Block)

		for i := 0; i < len(function.Args); i++ {
			vm.Log("func block: %d", i)
			if i < len(fc.Block) {
				vm.Log("set func value %s = %v", function.Args[i].Name, funcArgs[function.Args[i].Name])
				vmFunc.SetVariable(function.Args[i].Name, funcArgs[function.Args[i].Name])
			} else {
				vm.Log("set default func value %s = %v", function.Args[i].Name, function.Args[i].DefaultValue)
				vmFunc.SetVariable(function.Args[i].Name, function.Args[i].DefaultValue)
			}
		}

		err := vmFunc.Run()
		if err != nil {
			return nil, vm.Error("error running function block: %v", err)
		}

		vm.SetError(vmFunc.GetError())
		vm.SetExtended(vmFunc.GetExtended())
		vm.SetReturnValue(vmFunc.GetReturnValue())
		return vmFunc.returnValue, nil
	}
	return nil, vm.Error("no handler for function %s", fc.Name)
}