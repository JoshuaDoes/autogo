package autoit

import (
	"io"
	"os"
)

func init() {
	stdFunctions["fileclose"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "filehandle"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			file := vm.GetHandle(args["filehandle"].Handle())
			if file == nil {
				return NewToken(tNUMBER, 0), nil
			}
			file.(*os.File).Close()
			vm.DestroyHandle(args["filehandle"].Handle())
			return NewToken(tNUMBER, 1), nil
		},
	}
	stdFunctions["filedelete"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "filename"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			err := os.Remove(args["filename"].String())
			if err != nil {
				return NewToken(tNUMBER, 0), nil
			}
			return NewToken(tNUMBER, 1), nil
		},
	}
	stdFunctions["fileopen"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "filename"},
			&FunctionArg{Name: "mode", DefaultValue: NewToken(tNUMBER, 0)},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			file, err := os.OpenFile(args["filename"].String(), os.O_RDWR|os.O_APPEND, 0666)
			if err != nil {
				return NewToken(tNUMBER, -1), nil
			}
			handle := vm.AddHandle(file)
			return handle, nil
		},
	}
	stdFunctions["fileread"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "file"},
			&FunctionArg{Name: "count", DefaultValue: NewToken(tNUMBER, 0)},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			file := &os.File{}
			close := false

			if args["file"].Type == tHANDLE {
				file = vm.GetHandle(args["file"].Handle()).(*os.File)
				if file == nil {
					vm.SetError(1)
					return NewToken(tSTRING, ""), nil
				}
			} else {
				openFile, err := os.Open(args["file"].String()) //TODO: os.OpenFile() with flags
				if err != nil {
					vm.SetError(1)
					return NewToken(tSTRING, ""), nil
				}
				file = openFile
				close = true
			}

			fileStat, err := file.Stat()
			if err != nil {
				vm.SetError(1)
				return NewToken(tSTRING, ""), nil
			}

			fileData := make([]byte, fileStat.Size())
			if args["count"].Int() > 0 {
				fileData = make([]byte, args["count"].Int())
			}
			fileRead, err := file.Read(fileData)
			if close {
				file.Close()
			}
			vm.SetExtended(fileRead)
			if err == io.EOF {
				vm.SetError(-1)
				return NewToken(tSTRING, ""), nil
			}
			if err != nil {
				vm.SetError(1)
			}
			return NewToken(tBINARY, fileData), nil
		},
	}
	stdFunctions["filewrite"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "file"},
			&FunctionArg{Name: "data"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			file := &os.File{}
			close := false

			if args["file"].Type == tHANDLE {
				file = vm.GetHandle(args["file"].Handle()).(*os.File)
				if file == nil {
					vm.SetError(1)
					return NewToken(tNUMBER, 0), nil
				}
			} else {
				openFile, err := os.OpenFile(args["file"].String(), os.O_RDWR|os.O_CREATE, 0666)
				if err != nil {
					vm.SetError(1)
					return NewToken(tNUMBER, 0), nil
				}
				file = openFile
				close = true
			}

			_, err := file.Write(args["data"].Bytes())
			if close {
				file.Close()
			}
			if err != nil {
				return NewToken(tNUMBER, 0), nil
			}
			return NewToken(tNUMBER, 1), nil
		},
	}
}