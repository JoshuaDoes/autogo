#cs
Example script for interpreter testing
Use at your own risk
#ce
Local $sValue = String(3 * 5)
Global $iValue = 4 * 9 + 0.1

ConsoleWriteLn($sValue)
ConsoleWriteLn($iValue)

Func ConsoleWriteLn($msg = "")
	ConsoleWrite($msg & @CRLF)
EndFunc