#cs
Example script for interpreter testing
Use at your own risk
#ce
$bTrue = True And True
ConsoleWrite("Expecting True: " & $bTrue & @CRLF)
$bFalse = True And False
ConsoleWrite("Expecting False: " & $bFalse & @CRLF)
$bTrue = True Or True
ConsoleWrite("Expecting True: " & $bTrue & @CRLF)
$bTrue = False Or True
ConsoleWrite("Expecting True: " & $bTrue & @CRLF)
$bTrue = True And Not False
ConsoleWrite("Expecting True: " & $bTrue & @CRLF)
$bFalse = True And Not True
ConsoleWrite("Expecting False: " & $bFalse & @CRLF)