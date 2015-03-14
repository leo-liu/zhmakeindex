@echo off
setlocal
%~d0
cd %~p0

echo ==========^> Updating version...
echo \def\zhmVersion{%%> zhm-version.tex
if exist ..\VERSION (
    type ..\VERSION >> zhm-version.tex
) else (
    echo devel%%>> zhm-version.tex
)
echo }>> zhm-version.tex
echo \def\zhmRevision{%%>> zhm-version.tex
git rev-parse --short HEAD >> zhm-version.tex
echo }>> zhm-version.tex

echo ==========^> Compiling document...
xelatex -interaction=batchmode zhmakeindex.tex
echo ==========^> Running BibTeX...
bibtex zhmakeindex
echo ==========^> Running zhmakeindex...
zhmakeindex zhmakeindex
echo ==========^> Recompiling document...
xelatex -interaction=batchmode zhmakeindex.tex
xelatex -interaction=batchmode zhmakeindex.tex

endlocal
