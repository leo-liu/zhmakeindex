setlocal
for /f "delims=" %%i in ('hg parent --template "{rev}({node|short})"') do set Revision=%%i
set FLAGS=-ldflags "-s -w -X main.Revision %Revision%"

go install %FLAGS%

endlocal
