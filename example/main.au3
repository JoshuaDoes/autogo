Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg & @CRLF)
	SetError(0)
EndFunc

If 0 < 1 Then
	ConsoleWriteLn("0 < 1: True")
Else
	ConsoleWriteLn("0 < 1: False")
EndIf

If 1 > 0 Then
	ConsoleWriteLn("1 > 0: True")
Else
	ConsoleWriteLn("1 > 0: False")
EndIf

If @ScriptName = "main.au3" Then
	ConsoleWriteLn("@ScriptName = main.au3: True")
Else
	ConsoleWriteLn("@ScriptName = main.au3: False")
EndIf