setlocal
%~d0
cd %~p0

if exist VERSION (
  for /f "delims=" %%i in (VERSION) do set zhmVersion=%%i
) else (
  set zhmVersion=devel
)
for /f "delims=" %%i in ('git log -1 --pretty^=format:"%%h(%%ai"') do set zhmRevision=%%i
set zhmRevision=%zhmRevision:~0,18%)
set FLAGS=-ldflags "-X main.Version=%zhmVersion% -X main.Revision=%zhmRevision%"

go install %FLAGS%

endlocal
