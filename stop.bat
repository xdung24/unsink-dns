@echo off
setlocal
set "SCRIPT_DIR=%~dp0"
echo Attempting to run unsinkdns.exe from "%SCRIPT_DIR%"...
if exist "%SCRIPT_DIR%unsinkdns.exe" (
    "%SCRIPT_DIR%unsinkdns.exe" --stop
) else (
    echo unsinkdns.exe not found in "%SCRIPT_DIR%".
    echo Trying to run unsinkdns.exe from PATH...
    unsinkdns.exe --stop
)
pause