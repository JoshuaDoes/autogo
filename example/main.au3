#cs
Example script for interpreter testing
Use at your own risk
#ce
Local $sValue = String(3 * 5)
Global $iValue = 4 * 9 + 0.1

ConsoleWriteLn($sValue)
ConsoleWriteLn($iValue)
ConsoleWriteLn($sValue + $iValue)

Func ConsoleWriteLn($sMsg = "")
	Local $sNewMsg = $sMsg & @CRLF
	ConsoleWrite($sNewMsg)
EndFunc