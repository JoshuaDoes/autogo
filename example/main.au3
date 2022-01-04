$iStartTime = TimerInit()
Sleep(2000)
$iEndTime = TimerDiff($iStartTime)
ConsoleWriteLn("Execution time: " & $iEndTime)

Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg & @CRLF)
EndFunc