-- Add last_logout_at to users table for tracking token invalidation
ALTER TABLE public.users
ADD COLUMN last_logout_at timestamp without time zone DEFAULT NULL;

-- Create index for faster queries
CREATE INDEX idx_users_last_logout_at ON public.users(last_logout_at);
