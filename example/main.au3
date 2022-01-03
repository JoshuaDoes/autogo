#cs
Example script for interpreter testing
Use at your own risk
#ce
$timeStart = TimerInit()
$iNum = 13 + 3.2 * 2 / 8.0
$timeEnd = TimerDiff($timeStart)
ConsoleWrite("Evaluated (13 + 3.2 * 2 / 8.0) as (" & $iNum & ") in " & $timeEnd & " milliseconds" & @CRLF)