;#Debug

#cs ;>> Needs work still
;Definitions
Local $aArray[2] = [1, "Example"]
Local $mMap[]
Local $pPtr = Ptr(-1)
Local $hWnd = WinGetHandle(AutoItWinGetTitle())
Local $oObject = ObjCreate("Scripting.Dictionary")
Local $tStruct = DllStructCreate("wchar[256]")

;Write lines
				"$aArray:" & @TAB & @TAB & VarGetType($aArray) & @CRLF & _
				"$mMap:" & @TAB & @TAB & VarGetType($mMap) & @CRLF & _
				"$pPtr:" & @TAB & @TAB & VarGetType($pPtr) & @CRLF & _
				"$hWnd:" & @TAB & @TAB & VarGetType($hWnd) & @CRLF & _
				"$oObject:" & @TAB & VarGetType($oObject) & @CRLF & _
				"$tStruct:" & @TAB & @TAB & VarGetType($tStruct) & @CRLF & _
				"MsgBox:" & @TAB & @TAB & VarGetType(MsgBox) & @CRLF & _
#ce ;<< Needs work still

;Working now
Local $dBinary = Binary("0x00204060")
Local $bBoolean = False
Local $iInt = 1
Local $fFloat = 2.0
Local $sString = "Some text"
Local $vKeyword = Default
Local $nNull = Null
Local $fuUserFunc = Test
Local $fuFunc = ConsoleWrite

ConsoleWrite("----" & @CRLF & _
				"Variable Types" & @CRLF & @CRLF & _
				"$dBinary:" & @TAB & @TAB & VarGetType($dBinary) & @CRLF & _
				"$nNull:" & @TAB & @TAB & VarGetType($nNull) & @CRLF & _
				"$bBoolean:" & @TAB & VarGetType($bBoolean) & @CRLF & _
				"$iInt:" & @TAB & @TAB & VarGetType($iInt) & @CRLF & _
				"$fFloat:" & @TAB & @TAB & VarGetType($fFloat) & @CRLF & _
				"$sString:" & @TAB & @TAB & VarGetType($sString) & @CRLF & _
				"$vKeyword:" & @TAB & VarGetType($vKeyword) & @CRLF & _
				"Func 'Test':" & @TAB & VarGetType(Test) & @CRLF & _
				"$fuUserFunc:" & @TAB & VarGetType($fuUserFunc) & @CRLF & _
				"$fuFunc:" & @TAB & @TAB & VarGetType($fuFunc) & @CRLF & _
"----" & @CRLF)

Test("asdf123", Default)

Func Test($msg1, $msg2 = "asdf")
	ConsoleWrite($msg1 & @CRLF)
	ConsoleWrite($msg2 & @CRLF)
EndFunc   ;==>Test