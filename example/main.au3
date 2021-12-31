#cs
Example script for interpreter testing
Use at your own risk
#ce
;#Debug
$iNum = 2 + 3
ConsoleWrite("Expecting '5': " & $iNum & @CRLF)
$iNum = $iNum + 2
ConsoleWrite("Expecting '7': " & $iNum & @CRLF)
$iNum = $iNum * 2
ConsoleWrite("Expecting '14': " & $iNum & @CRLF)
$iNum = $iNum - 2
ConsoleWrite("Expecting '12': " & $iNum & @CRLF)

$iNum = "asdf" + $iNum
ConsoleWrite("Expecting '12': " & $iNum & @CRLF)