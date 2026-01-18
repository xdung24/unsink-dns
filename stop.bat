@echo off
setlocal
set "SCRIPT_DIR=%~dp0"
echo Attempting to run unsink-dns.exe from "%SCRIPT_DIR%"...
if exist "%SCRIPT_DIR%unsink-dns.exe" (
    "%SCRIPT_DIR%unsink-dns.exe" --stop
) else (
    echo unsink-dns.exe not found in "%SCRIPT_DIR%".
    echo Trying to run unsink-dns.exe from PATH...
    unsink-dns.exe --stop
)
pause