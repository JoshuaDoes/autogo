#cs
Example script for interpreter testing
Use at your own risk
#ce
$sMsg = "Hello, world! I'm running from " & @AutoItExe & " with a PID of " & @AutoItPID & "!"
ConsoleWrite($sMsg & @CRLF)

#Debug
$sErr = "This isn't really AutoIt v" & @AutoItVersion & " :c"
ConsoleWriteError($sErr & @CRLF)