package autoit

import (
	"os"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sqweek/dialog"
)

var (
	stdFunctions = map[string]*Function{
		"ConsoleWrite": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "text"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				fmt.Fprint(os.Stdout, args["text"].Data)
				return nil, nil
			},
		},
		"ConsoleWriteError": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "error"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				fmt.Fprint(os.Stderr, args["error"].Data)
				return nil, nil
			},
		},
		"FileOpenDialog": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "title"},
				&FunctionArg{Name: "initDir"},
				&FunctionArg{Name: "filter"},
				&FunctionArg{Name: "options", DefaultValue: NewToken(tNUMBER, "0")},
				&FunctionArg{Name: "defaultName", DefaultValue: NewToken(tSTRING, "")},
				&FunctionArg{Name: "hwnd", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				filterSplit := strings.Split(args["filter"].Data, "(")
				if len(filterSplit) < 2 {
					return nil, vm.Error("invalid file filter `%s`: %v", args["filter"].Data, filterSplit)
				}
				filterSplit2 := strings.Split(filterSplit[1], ")")
				filters := strings.Split(filterSplit2[0], "|")

				file, err := dialog.File().Title(args["title"].Data).SetStartDir(args["initDir"].Data).Filter(args["filter"].Data, filters...).Load()
				if err != nil {
					if err == dialog.ErrCancelled {
						return NewToken(tSTRING, ""), nil
					}
					return nil, err
				}
				return NewToken(tSTRING, file), nil
			},
		},
		"FileSaveDialog": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "title"},
				&FunctionArg{Name: "initDir"},
				&FunctionArg{Name: "filter"},
				&FunctionArg{Name: "options", DefaultValue: NewToken(tNUMBER, "0")},
				&FunctionArg{Name: "defaultName", DefaultValue: NewToken(tSTRING, "")},
				&FunctionArg{Name: "hwnd", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				filterSplit := strings.Split(args["filter"].Data, "(")
				if len(filterSplit) < 2 {
					return nil, vm.Error("invalid file filter `%s`: %v", args["filter"].Data, filterSplit)
				}
				filterSplit2 := strings.Split(filterSplit[1], ")")
				filters := strings.Split(filterSplit2[0], "|")

				file, err := dialog.File().Title(args["title"].Data).SetStartDir(args["initDir"].Data).Filter(args["filter"].Data, filters...).Save()
				if err != nil {
					if err == dialog.ErrCancelled {
						return NewToken(tSTRING, ""), nil
					}
					return nil, err
				}
				return NewToken(tSTRING, file), nil
			},
		},
		"FileSelectFolder": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "dialogText"},
				&FunctionArg{Name: "rootDir"},
				&FunctionArg{Name: "flag", DefaultValue: NewToken(tNUMBER, "0")},
				&FunctionArg{Name: "initialDir", DefaultValue: NewToken(tSTRING, "")},
				&FunctionArg{Name: "hwnd", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				directory, err := dialog.Directory().Title(args["dialogText"].Data).SetStartDir(args["initialDir"].Data).Browse()
				if err != nil {
					if err == dialog.ErrCancelled {
						return NewToken(tSTRING, ""), nil
					}
					return nil, err
				}
				return NewToken(tSTRING, directory), nil
			},
		},
		"FileRead": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "file"},
				&FunctionArg{Name: "count", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				fileData, err := os.ReadFile(args["file"].Data)
				if err != nil {
					return nil, err
				}
				if args["count"].Int() > 0 {
					return NewToken(tSTRING, string(fileData[:args["count"].Int()])), nil
				}
				return NewToken(tSTRING, string(fileData)), nil
			},
		},
		"FileWrite": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "file"},
				&FunctionArg{Name: "data"},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				err := os.WriteFile(args["file"].Data, args["data"].Bytes(), 0666)
				if err != nil {
					return NewToken(tNUMBER, "0"), err
				}
				return NewToken(tNUMBER, "1"), nil
			},
		},
		"InetRead": &Function{
			Args: []*FunctionArg{
				&FunctionArg{Name: "url"},
				&FunctionArg{Name: "options", DefaultValue: NewToken(tNUMBER, "0")},
			},
			Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
				resp, err := http.Get(args["url"].Data)
				if err != nil {
					return nil, err
				}
				defer resp.Body.Close()
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil, err
				}
				return NewToken(tSTRING, string(data)), nil
			},
		},
		"MsgBox": &Function{
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
					dialog.Message("%s", args["text"].Data).Title(args["title"].Data).Info()
					return NewToken(tNUMBER, "1"), nil //$IDOK
				case 4: //$MB_YESNO
					yesno := dialog.Message("%s", args["text"].Data).Title(args["title"].Data).YesNo()
					if yesno {
						return NewToken(tNUMBER, "6"), nil //$IDYES
					}
					return NewToken(tNUMBER, "7"), nil //$IDNO
				case 16: //$MB_ICONERROR
					dialog.Message("%s", args["text"].Data).Title(args["title"].Data).Error()
					return NewToken(tNUMBER, "1"), nil //$IDOK
				}
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
	return nil, vm.Error("no handler for function %s", funcName)
}