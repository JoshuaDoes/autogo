#cs
Example script for interpreter testing
Use at your own risk
#ce
;#Debug
$iNum = 2 + 3 + 5
ConsoleWrite("Expecting '10': " & $iNum & @CRLF)
$iNum = $iNum + 2 + 8
ConsoleWrite("Expecting '20': " & $iNum & @CRLF)
$iNum = $iNum * 2 * 4 * 2
ConsoleWrite("Expecting '320': " & $iNum & @CRLF)
$iNum = $iNum - 2 - 8 - 200
ConsoleWrite("Expecting '110': " & $iNum & @CRLF)

$iNum = "asdf" + $iNum * 8
ConsoleWrite("Expecting '880': " & $iNum & @CRLF)
$iNum = $iNum + "asdf"
ConsoleWrite("Expecting '880': " & $iNum & @CRLF)
#Debug
$iNum = $iNum * 5 & "lol"
ConsoleWrite("Expecting '4400lol': " & $iNum & @CRLF)

MsgBox(0, "AutoGo", "Final result: " & $iNum)
$iYesNo = MsgBox(4, "AutoGo", "Yes or no?")
MsgBox(16, "AutoGo", "You chose: " & $iYesNo)

$sFilePath = FileOpenDialog("Select a file to read...", @ScriptDir, "Any file (*)")
MsgBox(0, "AutoGo - File", $sFilePath)