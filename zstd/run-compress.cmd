SET CORPUS=compress
SET GO111MODULE=off
del /Q fuzz-build.zip
del /Q %CORPUS%\crashers\*.*
del /Q %CORPUS%\suppressions\*.*
SET /a PROCS=%NUMBER_OF_PROCESSORS%*3/4

REM go-fuzz-build -o=fuzz-build.zip -func=FuzzCompress .
go-fuzz-build -tags=datadog -o=fuzz-build.zip -func=FuzzCompressRef .


:LOOP
go run ../timeout.go -duration=15m go-fuzz -timeout=60 -minimize=5s -bin=fuzz-build.zip -workdir=%CORPUS% -procs=%PROCS%
GOTO LOOP
