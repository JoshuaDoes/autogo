#cs
Example script for interpreter testing
Use at your own risk
#ce
;#debug
$timeStart = TimerInit()
$iNum = 13 + 3.2 * 2 / 8.0
$timeEnd = TimerDiff($timeStart)
ConsoleWrite("Evaluated (13 + 3.2 * 2 / 8.0) as (" & $iNum & ") in " & $timeEnd & " milliseconds" & @CRLF)

$timeStart = TimerInit()
$updateData = InetRead("http://www.autoitscript.com/autoit3/files/beta/update.dat")
$timeEnd = TimerDiff($timeStart)
$updateData = String($updateData)
ConsoleWrite("Found update in " & $timeEnd & " milliseconds:" & @CRLF & $updateData)