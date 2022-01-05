;#Debug
$iTime = 2000
$iStartTime = TimerInit()
Sleep($iTime)
$iEndTime = TimerDiff($iStartTime)

$iEndSeconds = $iEndTime / 1000

;#Debug
ConsoleWriteLn(@OSType)
;#Debug
If $iEndTime < $iTime Then
	If @ScriptName = "main.au3" Then
		ConsoleWriteLn("Likely using AutoIt, slept for " & $iTime & "ms and really slept for " & $iEndTime & "ms (approx " & $iEndSeconds & " seconds)")
	Else
		ConsoleWriteLn("Guaranteed using AutoIt, slept for " & $iTime & "ms and really slept for " & $iEndTime & "ms (approx " & $iEndSeconds & " seconds)")
	EndIf
ElseIf $iEndTime > $iTime Then
	ConsoleWriteLn("Possibly using AutoGo, slept for " & $iTime & "ms and really slept for " & $iEndTime & "ms (approx " & $iEndSeconds & " seconds)")
Else
	ConsoleWriteLn("Guaranteed using magics, slept for " & $iTime & "ms")
EndIf

Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg & @CRLF)
	SetError(0)
EndFunc