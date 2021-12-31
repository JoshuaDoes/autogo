#cs
Example script for interpreter testing
Use at your own risk
#ce
;#Debug
$src = FileOpenDialog("Select a file to read...", @ScriptDir, "Any file (*)")
MsgBox(0, "AutoGo - Source", $src)
$dst = FileSaveDialog("Choose where to save it to...", @ScriptDir, "Any file (*)")
MsgBox(0, "AutoGo - Destination", $dst)

$data = FileRead($src)
MsgBox(0, $src, $data)
#Debug
$code = FileWrite($dst, $data)
ConsoleWrite("Wrote data: " & $code & @CRLF)