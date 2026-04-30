@echo off
REM 加载环境变量
if exist .env (
    for /f "usebackq tokens=*" %%i in (`.env`) do set %%i
)

REM 设置默认值（如果.env文件不存在）
if "%FAYHUB_DB_PASSWORD%"=="" set FAYHUB_DB_PASSWORD=fayhub123
if "%FAYHUB_JWT_SECRET%"=="" set FAYHUB_JWT_SECRET=fayhub_jwt_secret_dev_2026

cd /d d:\kaifa\2026\saas\FayHub\backend
go run cmd/main.go
