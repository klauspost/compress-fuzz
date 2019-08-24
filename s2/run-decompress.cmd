SET CORPUS=decompress
del /Q fuzz-build.zip
del /Q %CORPUS%\crashers\*.*
del /Q %CORPUS%\suppressions\*.*

go-fuzz-build -o=fuzz-build.zip -tags=noasm -func=FuzzDecompress .
go-fuzz -bin=fuzz-build.zip -workdir=%CORPUS% -procs=10
