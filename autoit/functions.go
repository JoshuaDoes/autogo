package autoit

import (
	"fmt"
	"os"
)

func (vm *AutoItVM) HandleFunc(funcName string, params []*Token) (*Token, error) {
	switch funcName {
	case "ConsoleWrite":
		return ConsoleWrite(params[0].Data), nil
	case "ConsoleWriteError":
		return ConsoleWriteError(params[0].Data), nil
	}
	return nil, vm.Error("unknown function: %s", funcName)
}

func ConsoleWrite(msg string) *Token {
	fmt.Fprint(os.Stdout, msg)
	return nil
}
func ConsoleWriteError(msg string) *Token {
	fmt.Fprint(os.Stderr, msg)
	return nil
}