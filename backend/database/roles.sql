-- DB hazırlık: superuser ile bir kez:
-- CREATE DATABASE radikal_hesap;

-- ⚠️ GÜVENLİK UYARISI (PRODUCTION):
-- Aşağıdaki 'app_pass' / 'sys_pass' yalnızca LOKAL geliştirme içindir.
-- Canlı ortamda bu script'i çalıştırmadan ÖNCE şifreleri güçlü, rastgele
-- değerlerle değiştirin ve aynı değerleri backend .env içindeki
-- DATABASE_URL / SYSTEM_DATABASE_URL ile eşitleyin. Mevcut bir kurulumda
-- değiştirmek için: ALTER ROLE radikal_app WITH PASSWORD '...';

-- Sonra radikal_hesap içinde bu script çalıştırılır:
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'radikal_app') THEN
        CREATE ROLE radikal_app WITH LOGIN PASSWORD 'app_pass';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'radikal_system') THEN
        CREATE ROLE radikal_system WITH LOGIN PASSWORD 'sys_pass' BYPASSRLS;
    END IF;
END
$$;

-- Schema yetkileri
GRANT USAGE, CREATE ON SCHEMA public TO radikal_app, radikal_system;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO radikal_app, radikal_system;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO radikal_app, radikal_system;

-- Gelecekte postgres ve radikal_system tarafından oluşturulacak tüm nesneler için yetkiler
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO radikal_app, radikal_system;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO radikal_app, radikal_system;
ALTER DEFAULT PRIVILEGES FOR ROLE radikal_system IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO radikal_app, radikal_system;
ALTER DEFAULT PRIVILEGES FOR ROLE radikal_system IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO radikal_app, radikal_system;
