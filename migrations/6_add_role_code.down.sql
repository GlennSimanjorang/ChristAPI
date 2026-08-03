-- Remove default roles
DELETE FROM public.roles WHERE code IN ('super_admin', 'admin', 'church_staff', 'public');

-- Remove unique constraint and code column
ALTER TABLE public.roles DROP CONSTRAINT IF EXISTS roles_code_unique;
ALTER TABLE public.roles DROP COLUMN IF EXISTS code;
