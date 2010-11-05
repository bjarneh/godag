@ECHO OFF
REM ==================================================
REM Build tool for godag on Windows
REM ==================================================
REM This script does not contain the 'cproot' target
REM but hopefully it will build godag on Windows...
REM
REM Copyright (C) 2009 all rights reserved 
REM GNU GENERAL PUBLIC LICENSE VERSION 3.0
REM Author bjarneh@ifi.uio.no


IF "%1" == "install" GOTO SANITY
IF "%1" == "clean"   GOTO CLEAN
IF "%1" == "help"    GOTO HELP


:HELP
ECHO.
ECHO build.bat - utility script for godag
ECHO. 
ECHO This script has 3 legal targets
ECHO.
ECHO [Targets]
ECHO.
ECHO help     -  display this message and exit
ECHO clean    -  del *.8 from src directory + GOBIN\gd 
ECHO install  -  compile and move binary to GOBIN
GOTO END


:CLEAN
ECHO [clean]
DEL src\utilz\*.8
DEL src\start\*.8
DEL src\cmplr\*.8
DEL src\parse\*.8
GOTO END

:SANITY
IF "%GOROOT%" == "" GOTO FAIL
IF "%GOOS%"   == "" GOTO FAIL
IF "%GOARCH%" == "" GOTO FAIL
IF "%GOBIN%"  == "" GOTO FAIL
GOTO BUILD

:BUILD
ECHO [install]
CHDIR src\utilz
8g.exe walker.go
8g.exe handy.go
8g.exe stringset.go
8g.exe stringbuffer.go
CHDIR ..\parse
8g.exe -o gopt.8 option.go gopt.go
cd ..\cmplr
8g.exe -I ..\ dag.go
8g.exe -I ..\ compiler.go
CHDIR ..\start
8g.exe -I ..\ main.go
8l.exe -L ..\ -o ..\..\gd.exe main.8
CHDIR ..\..\
MOVE gd.exe %GOBIN%
GOTO END


:FAIL
ECHO [ERROR] Missing environment variable
ECHO [ERROR] GOROOT = %GOROOT%
ECHO [ERROR] GOARCH = %GOARCH%
ECHO [ERROR] GOBIN  = %GOBIN%
ECHO [ERROR] GOOS   = %GOOS%
GOTO END


:END
@ECHO ON
