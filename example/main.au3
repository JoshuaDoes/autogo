;#Debug
$iTime = 2000
$iStartTime = TimerInit()
Sleep($iTime)
$iEndTime = TimerDiff($iStartTime)

$iEndSeconds = $iEndTime / 1000

If @ScriptName = "main.au3" Then
	ConsoleWriteLn("You gotta be using AutoGo!")
EndIf
;#Debug
If $iEndTime > $iTime Then
	ConsoleWriteLn("Possibly using AutoGo, slept for " & $iTime & "ms and really slept for " & $iEndTime & "ms (approx " & $iEndSeconds & " seconds)")
ElseIf $iEndTime < $iTime Then
	If @ScriptName = "main.au3" Then
		ConsoleWriteLn("Likely using AutoIt, slept for " & $iTime & "ms and really slept for " & $iEndTime & "ms (approx " & $iEndSeconds & " seconds)")
	Else
		ConsoleWriteLn("Guaranteed using AutoIt, slept for " & $iTime & "ms and really slept for " & $iEndTime & "ms (approx " & $iEndSeconds & " seconds)")
	EndIf
Else
	ConsoleWriteLn("Guaranteed using magics, slept for " & $iTime & "ms")
EndIf

Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg & @CRLF)
EndFunc