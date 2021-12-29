package autoit

import (
	"os"
	"strconv"
	"runtime"
)

func (vm *AutoItVM) GetMacro(macro string) (*Token, error) {
	switch macro {
	case "AutoItExe":
		osExecutable, err := os.Executable()
		return NewToken(tSTRING, osExecutable), err
	case "AutoItPID":
		pid := os.Getpid()
		return NewToken(tNUMBER, strconv.Itoa(pid)), nil
	case "AutoItVersion":
		return NewToken(tSTRING, "3.3.15.4"), nil //Fake the targeted version
	case "AutoItX64":
		if runtime.GOARCH == "amd64" {
			return NewToken(tNUMBER, "1"), nil
		}
		return NewToken(tNUMBER, "0"), nil
	case "CR":
		return NewToken(tSTRING, "\r"), nil
	case "LF":
		return NewToken(tSTRING, "\n"), nil
	case "CRLF":
		return NewToken(tSTRING, "\r\n"), nil
	}
	return nil, vm.Error("illegal macro: %s", macro)
}