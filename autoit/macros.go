package autoit

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"path/filepath"
)

func (vm *AutoItVM) GetMacro(macro string) (*Token, error) {
	switch strings.ToLower(macro) {
	case "autoitexe":
		osExecutable, err := os.Executable()
		return NewToken(tSTRING, osExecutable), err
	case "autoitpid":
		pid := os.Getpid()
		return NewToken(tNUMBER, pid), nil
	case "autoitversion":
		return NewToken(tSTRING, "3.3.15.4"), nil //Fake the targeted version
	case "autoitx64":
		if strings.Contains(runtime.GOARCH, "64") {
			return NewToken(tNUMBER, 1)
		}
		return NewToken(tNUMBER, 0), nil
	case "cr":
		return NewToken(tSTRING, "\r"), nil
	case "crlf":
		return NewToken(tSTRING, "\r\n"), nil
	case "lf":
		return NewToken(tSTRING, "\n"), nil
	case "scriptdir":
		return NewToken(tSTRING, filepath.Dir(vm.scriptPath)), nil
	case "scriptfullpath":
		return NewToken(tSTRING, vm.scriptPath), nil
	case "scriptname":
		return NewToken(tSTRING, filepath.Base(vm.scriptPath)), nil
	}
	return nil, vm.Error("illegal macro: %s", macro)
}