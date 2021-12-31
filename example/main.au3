#cs
Example script for interpreter testing
Use at your own risk
#ce
;#Debug
$src = FileOpenDialog("Select a file to read...", @ScriptDir, "Any file (*)")
MsgBox(0, "AutoGo - Source", $src)
$dir = FileSelectFolder("Choose a folder for fun...", "", 0, @ScriptDir)
$dst = FileSaveDialog("Choose where to save it to...", $dir, "Any file (*)")
MsgBox(0, "AutoGo - Destination", $dst)

ConsoleWrite("Writing: " & $src & @CRLF & "-> " & $dst & @CRLF)
;FileWrite($dst, FileRead($src))