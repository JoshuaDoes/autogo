Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg)
	If @OSType = "WIN32_NT" Then
		ConsoleWrite(@CR)
	EndIf
	ConsoleWrite(@LF)
	Return SetError(1, 0, 1) ;Make useful for exit
EndFunc

$dBin = Binary("abc")
$sString = String($dBin)
ConsoleWriteLn($sString)