@echo off
setlocal
set "BLUESKY_PROJECT=%~dp0.."
if "%~2" NEQ "" goto invalid
if /I "%~1"=="run" goto run
if /I "%~1"=="off" goto off
if "%~1"=="" goto help
if /I "%~1"=="help" goto help
goto invalid

:run
call docker info --format "{{.ServerVersion}}" >nul 2>&1
if errorlevel 1 (
  echo Open Docker Desktop first, then run: bluesky run
  exit /b 1
)
call docker compose --project-directory "%BLUESKY_PROJECT%" -f "%BLUESKY_PROJECT%\docker-compose.yml" --profile app up -d --build
exit /b %errorlevel%

:off
call docker compose --project-directory "%BLUESKY_PROJECT%" -f "%BLUESKY_PROJECT%\docker-compose.yml" --profile app --profile production-tools down
exit /b %errorlevel%

:help
echo bluesky run  - Build and start the local application.
echo bluesky off  - Stop the local application and keep database volumes.
exit /b 0

:invalid
echo Usage: bluesky run ^| off
exit /b 2
