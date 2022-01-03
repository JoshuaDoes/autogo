#cs
Example script for interpreter testing
Use at your own risk
#ce
;#debug
$url = "http://www.autoitscript.com/autoit3/files/beta/update.dat"
consoleWrite("Reading " & $URL & " ..." & @CrLf)
$data = inetRead($url)
Consolewrite($data)
msgBox(0, "AutoIt: update.dat", $DATA)