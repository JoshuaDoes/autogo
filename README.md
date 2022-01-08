# AutoGo
### An AutoIt interpreter and cross-platform runtime package

* WARNING: AutoGo is not yet feature complete. Some syntax features are still missing, the lexing logic may have flaws, and the internal function list is nowhere near ready for running production level scripts. Any functionality missing from AutoGo but present in AutoIt is considered a bug.

## What is AutoIt?
[AutoIt](https://www.autoitscript.com/) was developed as a procedural scripting language for Windows by Jonathan Bennett and first appeared in January of 1999. It began as a statement-driven language designed to automate user interactions, but over time and especially in AutoIt v3 it became a general purpose language with a syntax similar to BASIC and a focus on providing access to all of the Windows API through internal functions. The main feature that made AutoIt simpler to develop for versus other scripting languages was the ability to store various useful data types in what's internally a variant data type, to remain type agnostic in most expressions (except for where conversions are clearly needed, such as treating hex strings as binary).

## What is AutoGo?
AutoGo is a project I began at the tail end of 2021 in the quest to teach myself how to write a programming language. I decided that since I have a long history of AutoIt in my early programming and it's simple enough for any seasoned developer to understand, it would make for a good specification to try adhering to as efficiently as I know how using my own favorite language, Go. Why auto it (Windows) when you can auto Go (not just Windows)?

## Is AutoGo Windows-only?
No, and that's where things get interesting. See, AutoIt is as good of a language as it is because internally it has a heavy reliance on wrapping the functionality provided by various Windows APIs only available on Windows (and in most cases exposing API calls 1:1 with the internal function signature). But how can we make these same calls on other Go-supported platforms such as Linux?

Enter winelib. Or at least, when I get around to it. For the time being I'll use as much of Go's stdlib (or better packages) as I can while implementing the majority of AutoIt's internal functions, but eventually there will come a point where a function is clearly doing something specific to Windows. I chose winelib after a friend suggested it because it shares 100% of its code with wine directly and acts as a drop-in replacement (with some small work on top) for any program utilizing the Windows API headers. If all goes to plan, AutoGo's only dependency on non-Windows platforms will be wine itself, and all internal functions will transparently compile against either WinAPI on Windows or otherwise WineAPI (what I'll from here on out refer to winelib as) where faster Go alternatives are unavailable.

## How can I use it?
Because I made the decision to write AutoGo as a package at heart instead of just a program, there are three ways to use it: Installing the runtime, then either executing it with .au3 script files or executing it in interactive mode, or importing the package and spawning an unlimited number of interactable AutoIt VMs that can evaluate any given script block.

### Installing the runtime to your GOBIN directory
`go install github.com/JoshuaDoes/autogo/cmd/autogo`

### Executing an AutoIt script
`autogo "Hello World.au3"`

### Running in interactive mode (NOTE: not yet implemented!)
`autogo -i` OR `autogo /i`

### Using AutoIt in Go
AutoGo exposes the internal lexer functionality and runtime methods to any package which includes it, allowing a complete deep dive into an AutoIt script using Go code. For example, to run an internal hello world script you could do the following to load it into a lexer and print the tokens, then prefix it with a debug flag and create an AutoIt virtual machine from the altered script:

```Go
package main

import (
	"fmt"

	"github.com/JoshuaDoes/autogo/autoit"
)

var script = []byte(`
$sNoun = "world"
ConsoleWrite("Hello, " & $sNoun & "!" & @CRLF)
`)

func main() {
	lexer := autoit.NewLexer(script)
	tokens, err := lexer.GetTokens()
	if err != nil {
		panic(err)
	}
	for i, token := range tokens {
		fmt.Printf("Token %d: %v\n", i, *token)
	}

	fmt.Println("Starting VM...")
	debugScript := append([]byte(`#Debug`), script...)
	vm, err := autoit.NewAutoItVM("Hello World.au3", debugScript, nil)
	if err != nil {
		panic(err)
	}

	err = vm.Run()
	if err != nil {
		panic(err)
	}
}
```
Output (may be a little out of date):
```
PS C:\..\go\src\github.com\JoshuaDoes\autogo\example> .\windows.ps1
Token 0: {EOL  0 0}
Token 1: {VARIABLE sNoun 1 0}
Token 2: {OP = 1 6}
Token 3: {STRING world 1 8}
Token 4: {EOL  1 16}
Token 5: {CALL ConsoleWrite 2 0}
Token 6: {BLOCK ( 2 12}
Token 7: {STRING Hello,  2 13}
Token 8: {OP & 2 22}
Token 9: {VARIABLE sNoun 2 24}
Token 10: {OP & 2 31}
Token 11: {STRING ! 2 33}
Token 12: {OP & 2 37}
Token 13: {MACRO CRLF 2 39}
Token 14: {BLOCKEND ) 2 45}
Token 15: {EOL  2 46}
Starting VM...
vm: FLAG Debug
vm: VARIABLE sNoun
vm: evaluating {VARIABLE sNoun 2 0}
vm: evaluating {STRING world 2 8}
vm: $sNoun = {STRING world 2 8}
vm: CALL ConsoleWrite
vm: evaluating {CALL ConsoleWrite 3 0}
vm: BLOCK 1
vm: evaluating {STRING Hello,  3 13}
vm: evaluating {VARIABLE sNoun 3 24}
vm: expecting value for: {VARIABLE sNoun 3 24}
vm: append: {STRING Hello, world 0 0}
vm: evaluating {STRING ! 3 33}
vm: append: {STRING Hello, world! 0 0}
vm: evaluating {MACRO CRLF 3 39}
vm: append: {STRING Hello, world!
 0 0}
vm: BLOCK 1: {STRING Hello,  3 13} = &{STRING Hello, world!
 0 0}
vm: BLOCK REACHED EOL
vm: ------- 0 EVALUATING {STRING Hello, world!
 0 0}
vm: evalBlock 0: appending section with {STRING Hello, world!
 0 0}
vm: ------- 0 EVALUATING {BLOCKEND  0 0}
vm: depth: -1
vm: evaluating {STRING Hello, world!
 0 0}
vm: evalBlock: evaluated final section as {STRING Hello, world!
 0 0}
Hello, world!
```

AutoGo provides a simple cmd implementation as the primary example. It either reads scripts from file that were passed in from the command line and executes them, or it executes its own internal script if you don't pass any arguments (which currently attempts to run "main.au3" relative to the current directory, and exits with error code 1 if it fails to find it). This cmd may later be upgraded to be a drop-in replacement for the AutoIt SciTE environment's various tools.

### License
The source code for the AutoGo package is released and licensed under the Mozilla Public License Version 2.0. See [LICENSE](https://github.com/JoshuaDoes/autogo/blob/master/LICENSE) for more details. TL;DR: You are required to share your modifications to AutoGo no matter if you're releasing source code that depends on your modified version or if you're releasing a compiled program that sources from your modified version. However, you're free to use any other license you wish for code that uses AutoGo, such as in a compiled AutoIt script (using AutoGo instead of the original AutoIt compiler) or when importing AutoGo into another Go program, or any other scenario involving the use of this package.

### Donating
If AutoGo finds any use in your world, I'd appreciate at least a drink if you're willing!

[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://paypal.me/JoshuaDoes)

### Contributing
AutoGo follows a specific set of guidelines when developing for various pieces of the total package. This is to ensure that existing accurate scripts continue to work as expected with as little regressions as possible. When I'm ready to start accepting help with adding all the missing functionality, [CONTRIBUTING](https://github.com/JoshuaDoes/autogo/blob/master/CONTRIBUTING.md) will be ready for you to find out how to make your pull requests acceptable!