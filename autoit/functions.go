package autoit

import (
	"os"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sqweek/dialog"
)

var (
	stdFunctions = map[string]*Function{
		"string": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "expression"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				return NewToken(tSTRING, args["expression"].String()), nil
			},
		},
		"number": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "expression"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				return NewToken(tNUMBER, args["expression"].Float64()), nil
			},
		},
		"binary": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "expression"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				return NewToken(tBINARY, args["expression"].Bytes()), nil
			},
		},
		"consolewrite": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "text"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				fmt.Fprint(os.Stdout, args["text"].String())
				vm.stdout += args["text"].String()
				return nil, nil
			},
		},
		"consolewriteerror": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "error"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				fmt.Fprint(os.Stderr, args["error"].String())
				vm.stderr += args["error"].String()
				return nil, nil
			},
		},
		"fileopendialog": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "title"},
				&FunctionArg{Name: "initDir"},
				&FunctionArg{Name: "filter"},
				&FunctionArg{Name: "options", DefaultValue: NewToken(tNUMBER, "0")},
				&FunctionArg{Name: "defaultName", DefaultValue: NewToken(tSTRING, "")},
				&FunctionArg{Name: "hwnd", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				filterSplit := strings.Split(args["filter"].String(), "(")
				if len(filterSplit) < 2 {
					return nil, vm.Error("invalid file filter `%s`: %v", args["filter"].String(), filterSplit)
				}
				filterSplit2 := strings.Split(filterSplit[1], ")")
				filters := strings.Split(filterSplit2[0], "|")

				file, err := dialog.File().Title(args["title"].String()).SetStartDir(args["initDir"].String()).Filter(args["filter"].String(), filters...).Load()
				if err != nil {
					if err == dialog.ErrCancelled {
						return NewToken(tSTRING, ""), nil
					}
					return nil, err
				}
				return NewToken(tSTRING, file), nil
			},
		},
		"filesavedialog": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "title"},
				&FunctionArg{Name: "initDir"},
				&FunctionArg{Name: "filter"},
				&FunctionArg{Name: "options", DefaultValue: NewToken(tNUMBER, "0")},
				&FunctionArg{Name: "defaultName", DefaultValue: NewToken(tSTRING, "")},
				&FunctionArg{Name: "hwnd", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				filterSplit := strings.Split(args["filter"].String(), "(")
				if len(filterSplit) < 2 {
					return nil, vm.Error("invalid file filter `%s`: %v", args["filter"].String(), filterSplit)
				}
				filterSplit2 := strings.Split(filterSplit[1], ")")
				filters := strings.Split(filterSplit2[0], "|")

				file, err := dialog.File().Title(args["title"].String()).SetStartDir(args["initDir"].String()).Filter(args["filter"].String(), filters...).Save()
				if err != nil {
					if err == dialog.ErrCancelled {
						return NewToken(tSTRING, ""), nil
					}
					return nil, err
				}
				return NewToken(tSTRING, file), nil
			},
		},
		"fileselectfolder": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "dialogText"},
				&FunctionArg{Name: "rootDir"},
				&FunctionArg{Name: "flag", DefaultValue: NewToken(tNUMBER, "0")},
				&FunctionArg{Name: "initialDir", DefaultValue: NewToken(tSTRING, "")},
				&FunctionArg{Name: "hwnd", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				directory, err := dialog.Directory().Title(args["dialogText"].String()).SetStartDir(args["initialDir"].String()).Browse()
				if err != nil {
					if err == dialog.ErrCancelled {
						return NewToken(tSTRING, ""), nil
					}
					return nil, err
				}
				return NewToken(tSTRING, directory), nil
			},
		},
		"fileread": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "file"},
				&FunctionArg{Name: "count", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				fileData, err := os.ReadFile(args["file"].String())
				if err != nil {
					return nil, err
				}
				if args["count"].Int() > 0 {
					return NewToken(tBINARY, fileData[:args["count"].Int()]), nil
				}
				return NewToken(tBINARY, fileData), nil
			},
		},
		"filewrite": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "file"},
				&FunctionArg{Name: "data"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				err := os.WriteFile(args["file"].String(), args["data"].Bytes(), 0666)
				if err != nil {
					return NewToken(tNUMBER, "0"), err
				}
				return NewToken(tNUMBER, "1"), nil
			},
		},
		"inetread": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "url"},
				&FunctionArg{Name: "options", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				resp, err := http.Get(args["url"].String())
				if err != nil {
					return nil, err
				}
				defer resp.Body.Close()
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil, err
				}
				return NewToken(tBINARY, data), nil
			},
		},
		"msgbox": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "flag"},
				&FunctionArg{Name: "title"},
				&FunctionArg{Name: "text"},
				&FunctionArg{Name: "timeout", DefaultValue: NewToken(tNUMBER, "0")},
				&FunctionArg{Name: "hwnd", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				switch args["flag"].Int() {
				case 0: //$MB_OK
					dialog.Message("%s", args["text"].String()).Title(args["title"].String()).Info()
					return NewToken(tNUMBER, "1"), nil //$IDOK
				case 4: //$MB_YESNO
					yesno := dialog.Message("%s", args["text"].String()).Title(args["title"].String()).YesNo()
					if yesno {
						return NewToken(tNUMBER, "6"), nil //$IDYES
					}
					return NewToken(tNUMBER, "7"), nil //$IDNO
				case 16: //$MB_ICONERROR
					dialog.Message("%s", args["text"].String()).Title(args["title"].String()).Error()
					return NewToken(tNUMBER, "1"), nil //$IDOK
				}
				return nil, nil
			},
		},
		"timerdiff": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "handle"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				then := time.Unix(0, args["handle"].Int64() * int64(time.Millisecond))
				now := time.Now()
				diff := now.Sub(then)
				vm.Log("timerdiff: then(%s) now(%s) diff(%s)", then, now, diff)
				return NewToken(tNUMBER, diff.Milliseconds()), nil
			},
		},
		"timerinit": &Function{
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				return NewToken(tNUMBER, time.Now().UnixNano() / int64(time.Millisecond)), nil
			},
		},
		/*"consolewriteline": &Function{
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
	function, exists := vm.funcs[strings.ToLower(funcName)]
	if !exists {
		function, exists = stdFunctions[strings.ToLower(funcName)]
		if !exists {
			return nil, vm.Error("undefined function %s", funcName)
		}
	}

	if len(args) > len(function.Args) {
		for _, arg := range args {
			vm.Log("arg: %v", arg)
		}
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
	if function.Block != nil {
		/*vmFunc, err := NewAutoItTokenVM(vm.scriptPath, function.Block, vm.parentScope)
		if err != nil {
			return nil, err
		}
		vmFunc.Logger = vm.Logger

		vmFunc.funcs = vm.funcs
		vmFunc.vars = vm.vars*/

		vmFuncPtr := *vm
		vmFunc := &vmFuncPtr
		vmFunc.running = false
		vmFunc.suspended = false
		vmFunc.pos = 0
		vmFunc.tokens = function.Block
		vmFunc.parentScope = vm

		preprocess := vmFunc.Preprocess()
		if preprocess != nil {
			return nil, preprocess
		}

		for i := 0; i < len(function.Args); i++ {
			vm.Log("func block: %d", i)
			if i < len(args) {
				vm.Log("set func value %s = %v", function.Args[i].Name, args[i])
				vmFunc.SetVariable(function.Args[i].Name, args[i])
			} else {
				vm.Log("set func value %s = %v", function.Args[i].Name, function.Args[i].DefaultValue)
				vmFunc.SetVariable(function.Args[i].Name, function.Args[i].DefaultValue)
			}
		}

		err := vmFunc.Run()
		if err != nil {
			return nil, vm.Error("error running function block: %v", err)
		}

		return vmFunc.returnValue, nil
	}
	return nil, vm.Error("no handler for function %s", funcName)
}