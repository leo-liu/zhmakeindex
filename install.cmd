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

go install %FLAGS%

endlocal
