;#Debug
$iTime = 2000
$iStartTime = TimerInit()
Sleep($iTime)
$iEndTime = TimerDiff($iStartTime)

$iEndSeconds = $iEndTime / 1000

If $iEndTime < $iTime Then
	ConsoleWriteLn("Likely using AutoIt, slept for " & $iTime & " milliseconds and really slept for " & $iEndTime & " milliseconds (approx " & $iEndSeconds & " seconds)")
ElseIf $iEndTime > $iTime Then
	ConsoleWriteLn("Possibly using AutoGo, slept for " & $iTime & " milliseconds and really slept for " & $iEndTime & " milliseconds (approx " & $iEndSeconds & " seconds)")
Else
	ConsoleWriteLn("Guaranteed using magics, slept for " & $iTime & " milliseconds")
EndIf

Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg & @CRLF)
EndFunc