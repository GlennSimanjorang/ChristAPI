-- Remove logout tracking from users table
DROP INDEX IF EXISTS idx_users_last_logout_at;
ALTER TABLE public.users
DROP COLUMN IF EXISTS last_logout_at;
