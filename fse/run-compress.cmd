SET CORPUS=compress
SET GO111MODULE=off
del /Q fuzz-build.zip
del /Q %CORPUS%\crashers\*.*
del /Q %CORPUS%\suppressions\*.*
SET /a PROCS=%NUMBER_OF_PROCESSORS%*3/4

go-fuzz-build -o=fuzz-build.zip -func=FuzzCompress .
go-fuzz -bin=fuzz-build.zip -workdir=%CORPUS% -procs=%PROCS%
