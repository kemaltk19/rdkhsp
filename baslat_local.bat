@echo off
echo Radikal Hesap Projesi (Lokal - Docker'siz) Baslatiliyor...
cd /d "%~dp0"

echo 1. PostgreSQL veritabani Docker üzerinden baslatiliyor (zaten calisiyorsa devam eder)...
docker-compose up -d db

echo 2. Backend (Go) baslatiliyor...
start "Radikal Hesap - Backend" cmd /k "cd backend && go run main.go"

echo 3. Frontend (Node/Vite) baslatiliyor...
start "Radikal Hesap - Frontend" cmd /k "cd frontend && npm run dev"

echo Proje bilesenleri ayri pencerelerde baslatildi. Kapatmak istediginizde o pencereleri kapatiniz.
pause
