package autoit

import (
	"os"
	"strings"
)

//Preprocess steps through the script and prepares it for execution
func (vm *AutoItVM) Preprocess() error {
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
				includeVM, err := NewAutoItScriptVM(includeFile.String(), includeScript, vm)
				if err != nil {
					return err
				}
				err = includeVM.Preprocess()
				if err != nil {
					return err
				}
				for functionName, function := range includeVM.funcs {
					vm.funcs[functionName] = function
				}
			}
		case tFUNC:
			if !startLine {
				return vm.Error("preprocess: unexpected func")
			}
			startLine = false

			functionName, function, err := vm.PreprocessFunc()
			if err != nil {
				return err
			}

			if _, exists := vm.funcs[functionName]; exists {
				return vm.Error("preprocess: func %s already defined", functionName)
			}
			vm.funcs[functionName] = function
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
func (vm *AutoItVM) PreprocessFunc() (string, *Function, error) {
	tCall := vm.ReadToken()
	if tCall.Type != tCALL {
		return "", nil, vm.Error("preprocessor: expected call after func, instead got: %v", tCall)
	}

	tBlock := vm.ReadToken()
	if tBlock.Type != tLEFTPAREN {
		return "", nil, vm.Error("preprocessor: expected block after func call, instead got: %v", tBlock)
	}

	funcArgs := make([]*FunctionArg, 0)
	for {
		funcArg := &FunctionArg{}

		tVar := vm.ReadToken()
		if tVar == nil || tVar.Type == tRIGHTPAREN || tVar.Type == tEOL {
			break
		}
		if tVar.Type != tVARIABLE {
			return "", nil, vm.Error("preprocessor: expected variable in func call block, instead got: %v", tVar)
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
					return "", nil, vm.Error("preprocessor: expected default value for variable in func call block, instead got: %v", tValue)
				}

				funcArg.DefaultValue = tValue
				funcArgs = append(funcArgs, funcArg)
			default:
				return "", nil, vm.Error("preprocessor: unexpected operator for variable in func call block: %v", tOp)
			}
		default:
			return "", nil, vm.Error("preprocessor: unexpected token in func call block: %v", tOp)
		}
	}

	funcBlock := make([]*Token, 0)
	for {
		tBlock := vm.ReadToken()
		if tBlock == nil {
			return "", nil, vm.Error("preprocessor: unexpected end of func")
		}

		if tBlock.Type == tFUNCEND {
			break
		}
		funcBlock = append(funcBlock, tBlock)
	}

	return strings.ToLower(tCall.String()), &Function{Args: funcArgs, Block: funcBlock}, nil
}