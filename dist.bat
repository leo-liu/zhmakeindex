setlocal
set FLAGS="-ldflags=-s"

set GOOS=windows
set GOARCH=386
go build %FLAGS% -o release/windows_x86/zhmakeindex.exe
set GOARCH=amd64
go build %FLAGS% -o release/windows_x64/zhmakeindex.exe

set GOOS=linux
set GOARCH=386
go build %FLAGS% -o release/linux_x86/zhmakeindex
set GOARCH=amd64
go build %FLAGS% -o release/linux_x64/zhmakeindex

set GOOS=darwin
set GOARCH=386
go build %FLAGS% -o release/darwin_x86/zhmakeindex
set GOARCH=amd64
go build %FLAGS% -o release/darwin_x64/zhmakeindex

endlocal