@echo off

title Building PwdMan...

echo Building...

cd src
go build -o ../release/PwdMan.exe .

echo Build done!

pause
