SET CORPUS=flate
del /Q fuzz-build.zip
del /Q %CORPUS%\crashers\*.*
del /Q %CORPUS%\suppressions\*.*

go version >log.txt
go env >>log.txt
go-fuzz-build -o=fuzz-build.zip -func=Fuzz .
REM go-fuzz -minimize=5s -v=4 -bin=fuzz-build.zip -workdir=%CORPUS% -procs=5 2>>log.txt
go-fuzz -minimize=5s -bin=fuzz-build.zip -workdir=%CORPUS% -procs=6
