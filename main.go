package main

import (
	"fmt"
	"os"

	"github.com/JoshuaDoes/autogo/autoit"
)

func runVM(script []byte) error {
	vm, err := autoit.NewAutoItVM(script)
	if err != nil {
		return err
	}
	vm.Logger = true

	return vm.Run()
}

func main() {
	if len(os.Args) > 1 {
		for _, scriptFile := range os.Args[1:] {
			script, err := os.ReadFile(scriptFile)
			if err != nil {
				panic(err)
			}

			err = runVM(script)
			if err != nil {
				fmt.Println(err)
			}
		}
	} else {
		err := runVM([]byte(";INTERNAL SCRIPT\n#include \"main.au3\""))
		if err != nil {
			fmt.Println(err)
		}
	}
}