package autoit

import (
	"github.com/JamesHovious/w32"
)

func init() {
	stdFunctions["msgbox"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "flag"},
			&FunctionArg{Name: "title"},
			&FunctionArg{Name: "text"},
			&FunctionArg{Name: "timeout", DefaultValue: NewToken(tNUMBER, 0)}, //Return -1 on timeout
			&FunctionArg{Name: "hwnd", DefaultValue: NewToken(tNUMBER, 0)},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			return NewToken(tNUMBER, w32.MessageBox(args["hwnd"].HWND(), args["text"].String(), args["title"].String(), args["flag"].Uint())), nil
		},
	}
}

func (t *Token) HWND() w32.HWND {
	return w32.HWND(t.Uint())
}