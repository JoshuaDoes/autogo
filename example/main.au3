Func ConsoleWriteLn($sMsg = "")
	ConsoleWrite($sMsg)
	If @OSType = "WIN32_NT" Then
		ConsoleWrite(@CR)
	EndIf
	ConsoleWrite(@LF)
	Return SetError(1, 0, 1) ;Make useful for exit
EndFunc

;#Debug
Example()

Func Example()
    ; Create a variable in Local scope of the filepath that will be read/written to.
    Local $sFilePath = "example.txt"

	; Create a temporary file to write data to.
    If Not FileWrite($sFilePath, "Start of the FileWrite example, line 1. " & @CRLF) Then
        ConsoleWriteLn("An error occurred whilst writing the temporary file.")
        Return False
    EndIf

    ; Open the file for writing (append to the end of a file) and store the handle to a variable.
    Local $hFileOpen = FileOpen($sFilePath)
    If $hFileOpen = -1 Then
        ConsoleWriteLn("An error occurred whilst writing the temporary file.")
        Return False
    EndIf

    ; Write data to the file using the handle returned by FileOpen.
    If Not FileWrite($hFileOpen, "Line 2") Then
        ConsoleWriteLn("An error occurred whilst writing the temporary file.")
	EndIf
    If Not FileWrite($hFileOpen, "This is still line 2 as a new line wasn't appended to the last FileWrite call." & @CRLF) Then
        ConsoleWriteLn("An error occurred whilst writing the temporary file.")
	EndIf
    If Not FileWrite($hFileOpen, "Line 3" & @CRLF) Then
        ConsoleWriteLn("An error occurred whilst writing the temporary file.")
	EndIf
    FileWrite($hFileOpen, "Line 4")
    If Not FileWrite($hFileOpen, "Line 4") Then
        ConsoleWriteLn("An error occurred whilst writing the temporary file.")
	EndIf

    ; Close the handle returned by FileOpen.
    FileClose($hFileOpen)

    ; Display the contents of the file passing the filepath to FileRead instead of a handle returned by FileOpen.
    ConsoleWriteLn("Contents of the file:" & @CRLF & FileRead($sFilePath))

    ; Delete the temporary file.
    FileDelete($sFilePath)
EndFunc   ;==>Example