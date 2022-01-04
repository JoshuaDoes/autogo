package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/JoshuaDoes/autogo/autoit"
)

func runVM(scriptPath string, script []byte) error {
	vm, err := autoit.NewAutoItScriptVM(scriptPath, script, nil)
	if err != nil {
		return err
	}

	return vm.Run()
}

func main() {
	if len(os.Args) > 1 {
		scripts := make([]string, 0)
		for _, arg := range os.Args[1:] {
			switch strings.ToLower(arg) {
			case "/errorstdout":
				os.Stderr = os.Stdout
			default:
				scripts = append(scripts, arg)
			}
		}
		for _, scriptFile := range scripts {
			script, err := os.ReadFile(scriptFile)
			if err != nil {
				panic(err)
			}

			err = runVM(scriptFile, script)
			if err != nil {
				fmt.Println(err)
			}
		}
	} else {
		err := runVM("main.au3", []byte(";INTERNAL SCRIPT\n#include \"main.au3\""))
		if err != nil {
			fmt.Println(err)
		}
	}
}