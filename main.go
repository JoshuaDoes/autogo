package main

import (
	"fmt"
	"os"

	"github.com/JoshuaDoes/autogo/autoit"
)

func runVM(scriptPath string, script []byte) error {
	vm, err := autoit.NewAutoItVM(scriptPath, script, nil)
	if err != nil {
		return err
	}

	return vm.Run()
}

func main() {
	if len(os.Args) > 1 {
		for _, scriptFile := range os.Args[1:] {
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
		err := runVM(os.Args[0], []byte(";INTERNAL SCRIPT\n#include \"main.au3\""))
		if err != nil {
			fmt.Println(err)
		}
	}
}