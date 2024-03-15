ALTER TABLE IF exists "sessions" DROP CONSTRAINT IF exists "sessions_username_fkey";
DROP TABLE IF exists "sessions";
