DROP TABLE IF EXISTS public.user_otps;

ALTER TABLE public.users ALTER COLUMN password_hash SET NOT NULL;
ALTER TABLE public.users ALTER COLUMN is_active SET DEFAULT true;

ALTER TABLE public.users
DROP COLUMN IF EXISTS username,
DROP COLUMN IF EXISTS google_id,
DROP COLUMN IF EXISTS auth_provider,
DROP COLUMN IF EXISTS approval_status;
