Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg)
	If @OSType = "WIN32_NT" Then
		ConsoleWrite(@CR)
	EndIf
	ConsoleWrite(@LF)
	Return SetError(1, 0, 1) ;Make useful for exit
EndFunc

;#Debug
ConsoleWriteLn("hey there")

Local $mVar[]
$mVar["asdf"] = "data"
$data = $mVar["asdf"]
ConsoleWriteLn($data)