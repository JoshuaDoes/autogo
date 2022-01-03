#cs
Example script for interpreter testing
Use at your own risk
#ce
;#Debug
ConsoleWrite($asdf)
ConsoleWrite("Reading @ScriptFullPath ..." & @CRLF)
$script = FileRead(@ScriptFullPath)
ConsoleWrite(@ScriptName & ":" & @CRLF & $script & @CRLF)
MsgBox(0, "AutoGo: " & @ScriptName, $script)

$url = "http://www.autoitscript.com/autoit3/files/beta/update.dat"
ConsoleWrite("Reading " & $url & " ..." & @CRLF)
$data = InetRead($url)
ConsoleWrite($data)
MsgBox(0, "AutoIt: update.dat", $data)