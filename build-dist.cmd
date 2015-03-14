setlocal
%~d0
cd %~p0

if exist VERSION (
  for /f "delims=" %%i in (VERSION) do set zhmVersion=%%i
) else (
  set zhmVersion=devel
)
for /f "delims=" %%i in ('git rev-parse --short HEAD') do set zhmRevision=%%i
set FLAGS=-ldflags "-X main.Version %zhmVersion% -X main.Revision %zhmRevision%"

set GOOS=windows
set GOARCH=386
go build %FLAGS% -o bin/windows_x86/zhmakeindex.exe
set GOARCH=amd64
go build %FLAGS% -o bin/windows_x64/zhmakeindex.exe

set GOOS=linux
set GOARCH=386
go build %FLAGS% -o bin/linux_x86/zhmakeindex
set GOARCH=amd64
go build %FLAGS% -o bin/linux_x64/zhmakeindex

set GOOS=darwin
set GOARCH=386
go build %FLAGS% -o bin/darwin_x86/zhmakeindex
set GOARCH=amd64
go build %FLAGS% -o bin/darwin_x64/zhmakeindex

endlocal
