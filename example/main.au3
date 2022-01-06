Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg & @CRLF)
	SetError(1)
EndFunc

If 0 < 1 Then
	ConsoleWriteLn("0 < 1: True")
	If 1 > 0 Then
		ConsoleWriteLn("1 > 0: True")
		If @ScriptName = "main.au3" Then
			ConsoleWriteLn("@ScriptName = main.au3: True")
		Else
			ConsoleWriteLn("@ScriptName = main.au3: False")
		EndIf
	Else
		ConsoleWriteLn("1 > 0: False")
	EndIf
Else
	ConsoleWriteLn("0 < 1: False")
EndIf