package autoit

import (
	"io/ioutil"
	"net/http"
)

func init() {
	stdFunctions["inetread"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "url"},
			&FunctionArg{Name: "options", DefaultValue: NewToken(tNUMBER, 0)},
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
	}
}