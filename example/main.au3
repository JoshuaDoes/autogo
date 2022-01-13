;#Debug
$time = TimerInit()
$end = TimerDiff($time)
$value = 5 * 5

AddFive()
AddFive()
AddFive()
AddFive()
AddFive()
AddFive()
AddFive()
AddFive()
AddFive()
AddFive()

Func ConsoleWriteLn($sMsg = "")
        ConsoleWrite($sMsg)
        If @OSType = "WIN32_NT" Then
                ConsoleWrite(@CR)
        EndIf
        ConsoleWrite(@LF)
        Return SetError(1, 0, 1) ;Make useful for exit
EndFunc

Func AddFive()
        $value = $value + 5
        $end = TimerDiff($time)
        ConsoleWriteLn("Got result " & $value & " in " & $end & "ms")
EndFunc