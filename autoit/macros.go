package autoit

import (
	"os"
	"strconv"
)

func (vm *AutoItVM) GetMacro(macro string) (*Token, error) {
	switch macro {
	case "AutoItExe":
		osExecutable, err := os.Executable()
		return NewToken(tSTRING, osExecutable), err
	case "AutoItPID":
		pid := os.Getpid()
		return NewToken(tNUMBER, strconv.Itoa(pid)), nil
	case "CR":
		return NewToken(tSTRING, "\r"), nil
	case "LF":
		return NewToken(tSTRING, "\n"), nil
	case "CRLF":
		return NewToken(tSTRING, "\r\n"), nil
	}
	return nil, vm.Error("illegal macro: %s", macro)
}