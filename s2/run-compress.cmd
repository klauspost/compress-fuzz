SET CORPUS=compress
SET GO111MODULE=off
del /Q fuzz-build.zip
del /Q %CORPUS%\crashers\*.*
del /Q %CORPUS%\suppressions\*.*
REM There is concurrency in the encodes, so divide by 2.
SET /a PROCS=%NUMBER_OF_PROCESSORS%/2
REM MORE AGGRESSIVE:
SET /a PROCS=%NUMBER_OF_PROCESSORS%/4*3

go-fuzz-build -o=fuzz-build.zip -race -func=FuzzCompress .

:LOOP
go run ../timeout.go -duration=5m go-fuzz -minimize=5s -bin=fuzz-build.zip -workdir=%CORPUS% -procs=%PROCS%
GOTO LOOP

