package autoit

import (
	"time"
)

func init() {
	//Time management
	stdFunctions["sleep"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "delay"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			time.Sleep(time.Millisecond * time.Duration(args["delay"].Int64()))
			return nil, nil
		},
	}
	stdFunctions["timerdiff"] = &Function{
		Args: []*FunctionArg{
			&FunctionArg{Name: "handle"},
		},
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			now := time.Now()
			then := vm.GetHandle(args["handle"].Handle()).(time.Time)
			diff := now.Sub(then)
			vm.Log("timerdiff: then(%s) now(%s) diff(%s)", then, now, diff)
			return NewToken(tNUMBER, float64(diff.Nanoseconds()) / 1000000), nil
		},
	}
	stdFunctions["timerinit"] = &Function{
		Func: func(vm *AutoItVM, args map[string]*Token) (*Token, error) {
			return vm.AddHandle(time.Now()), nil
		},
	}
}