SET CORPUS=decompress
SET GO111MODULE=off
del /Q fuzz-build-%CORPUS%.zip
del /Q %CORPUS%\crashers\*.*
del /Q %CORPUS%\suppressions\*.*
SET /a PROCS=%NUMBER_OF_PROCESSORS%*3/4

go-fuzz-build -o=fuzz-build-%CORPUS%.zip -func=FuzzDecompress .

:LOOP
go run ../timeout.go -duration=10m go-fuzz -minimize=5s -bin=fuzz-build-%CORPUS%.zip -workdir=%CORPUS% -procs=%PROCS%
GOTO LOOP
