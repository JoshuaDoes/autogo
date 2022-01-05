package autoit

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
			return NewToken(tNUMBER, 1), nil
		}
		return NewToken(tNUMBER, 0), nil
	case "compiled":
		return NewToken(tNUMBER, 0), nil
	case "computername":
		hostname, err := os.Hostname()
		return NewToken(tSTRING, hostname), err
	case "cpuarch":
		switch runtime.GOARCH {
		case "amd64":
			return NewToken(tSTRING, "X64"), nil
		case "i386":
			return NewToken(tSTRING, "X86"), nil
		}
		return NewToken(tSTRING, runtime.GOARCH), nil
	case "cr":
		return NewToken(tSTRING, "\r"), nil
	case "crlf":
		return NewToken(tSTRING, "\r\n"), nil
	case "error":
		return NewToken(tNUMBER, vm.error), nil
	case "exitcode":
		return NewToken(tNUMBER, vm.exitCode), nil
	case "exitmethod":
		return NewToken(tCALL, vm.exitMethod), nil
	case "extended":
		return NewToken(tNUMBER, vm.extended), nil
	case "hour":
		return NewToken(tNUMBER, time.Now().Hour()), nil
	case "ipaddress1":
		netInterfaces, err := net.Interfaces()
		if err == nil && len(netInterfaces) > 0 {
			addresses, err := netInterfaces[0].Addrs()
			if err == nil && len(addresses) > 0 {
				var ip net.IP
				switch v := addresses[0].(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				return NewToken(tSTRING, ip.String()), nil
			}
		}
		return NewToken(tSTRING, "0.0.0.0"), nil
	case "ipaddress2":
		netInterfaces, err := net.Interfaces()
		if err == nil && len(netInterfaces) > 1 {
			addresses, err := netInterfaces[1].Addrs()
			if err == nil && len(addresses) > 1 {
				var ip net.IP
				switch v := addresses[0].(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				return NewToken(tSTRING, ip.String()), nil
			}
		}
		return NewToken(tSTRING, "0.0.0.0"), nil
	case "ipaddress3":
		netInterfaces, err := net.Interfaces()
		if err == nil && len(netInterfaces) > 2 {
			addresses, err := netInterfaces[2].Addrs()
			if err == nil && len(addresses) > 2 {
				var ip net.IP
				switch v := addresses[0].(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				return NewToken(tSTRING, ip.String()), nil
			}
		}
		return NewToken(tSTRING, "0.0.0.0"), nil
	case "ipaddress4":
		netInterfaces, err := net.Interfaces()
		if err == nil && len(netInterfaces) > 3 {
			addresses, err := netInterfaces[3].Addrs()
			if err == nil && len(addresses) > 3 {
				var ip net.IP
				switch v := addresses[0].(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				return NewToken(tSTRING, ip.String()), nil
			}
		}
		return NewToken(tSTRING, "0.0.0.0"), nil
	case "lf":
		return NewToken(tSTRING, "\n"), nil
	case "mday":
		return NewToken(tNUMBER, time.Now().Day()), nil
	case "min":
		return NewToken(tNUMBER, time.Now().Minute()), nil
	case "mon":
		return NewToken(tNUMBER, time.Now().Month()), nil
	case "msec":
		return NewToken(tNUMBER, float64(time.Now().UnixNano()) / 1000000), nil
	case "numparams":
		return NewToken(tNUMBER, vm.numParams), nil
	case "osarch":
		switch runtime.GOARCH {
		case "amd64":
			return NewToken(tSTRING, "X64"), nil
		case "i386":
			return NewToken(tSTRING, "X86"), nil
		case "ia64":
			return NewToken(tSTRING, "IA64"), nil
		}
		return NewToken(tSTRING, runtime.GOARCH), nil
	case "ostype":
		switch runtime.GOOS {
		case "windows":
			return NewToken(tSTRING, "WIN32_NT"), nil
		}
		return NewToken(tSTRING, runtime.GOOS), nil
	case "scriptdir":
		return NewToken(tSTRING, filepath.Dir(vm.scriptPath)), nil
	case "scriptfullpath":
		return NewToken(tSTRING, vm.scriptPath), nil
	case "scriptname":
		return NewToken(tSTRING, filepath.Base(vm.scriptPath)), nil
	case "sec":
		return NewToken(tNUMBER, time.Now().Second()), nil
	case "tab":
		return NewToken(tSTRING, "\t"), nil
	case "tempdir":
		return NewToken(tSTRING, os.TempDir()), nil
	case "wday":
		return NewToken(tNUMBER, int(time.Now().Weekday()) + 1), nil
	case "workingdir":
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return NewToken(tSTRING, wd), nil
	case "yday":
		return NewToken(tNUMBER, time.Now().YearDay()), nil
	case "year":
		return NewToken(tNUMBER, time.Now().Year()), nil
	}
	return nil, vm.Error("illegal macro: %s", macro)
}