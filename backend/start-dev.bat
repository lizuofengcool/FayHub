@echo off
chcp 65001 >nul
echo =========================================
echo   FayHub 后端启动脚本
echo =========================================

set FAYHUB_DB_PASSWORD=fayhub123
set FAYHUB_JWT_SECRET=fayhub_jwt_secret_dev_2026

echo [OK] 环境变量已设置
echo   FAYHUB_DB_PASSWORD=fayhub123
echo   FAYHUB_JWT_SECRET=fayhub_jwt_secret_dev_2026
echo.
echo 正在启动 Go 后端...

go run cmd/main.go
