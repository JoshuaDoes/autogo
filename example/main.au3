Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg)
	If @OSType = "WIN32_NT" Then
		ConsoleWrite(@CR)
	EndIf
	ConsoleWrite(@LF)
	Return SetError(1, 0, 1) ;Make useful for exit
EndFunc

;#Debug
$file = FileOpen("main.au3")
If @error Then
	Exit ConsoleWriteLn("Error opening file")
EndIf
ConsoleWriteLn("Handle: " & $file)
$data = FileRead($file)
If @error Then
	Exit ConsoleWriteLn("Error reading file")
EndIf
ConsoleWriteLn("Data:" & @CRLF & $data)