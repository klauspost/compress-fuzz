SET CORPUS=flate
del /Q fuzz-build.zip
del /Q %CORPUS%\crashers\*.*
del /Q %CORPUS%\suppressions\*.*

go-fuzz-build -o=fuzz-build.zip -func=Fuzz .
go-fuzz -bin=fuzz-build.zip -workdir=%CORPUS% -procs=4
