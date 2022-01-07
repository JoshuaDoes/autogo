Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg)
	If @OSType = "WIN32_NT" Then
		ConsoleWrite(@CR)
	EndIf
	ConsoleWrite(@LF)
	Return SetError(1, 0, 1) ;Make useful for exit
EndFunc

;#Debug
Local $var, $var2 = "line 2", $var3, $var4 = "line 4"
ConsoleWriteLn($var)
ConsoleWriteLn($var2)
ConsoleWriteLn($var3)
ConsoleWriteLn($var4)