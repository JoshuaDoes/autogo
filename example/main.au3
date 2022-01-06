Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg & @CRLF)
	SetError(1)
EndFunc

ConsoleWriteLn("Test")
Switch @error
	Case 0
		ConsoleWriteLn("0")
	Case 1
		ConsoleWriteLn("1")
	Case Else
		ConsoleWriteLn(@error)
EndSwitch

Switch @OSType
	Case "linux"
		ConsoleWriteLn("Hello there, free penguin!")
	Case "darwin"
		ConsoleWriteLn("Hello there, Apple person!")
	Case "windows", "WIN32_NT"
		ConsoleWriteLn("Hello there, Microsoft person!")
	Case "android"
		ConsoleWriteLn("Hello there, Google person!")
	Case "ios"
		ConsoleWriteLn("Hello there, Apple mobile person!")
	Case Else
		ConsoleWriteLn("Unrecognized platform: " & @OSType)
EndSwitch