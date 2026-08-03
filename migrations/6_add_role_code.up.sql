-- Add code column to roles table
ALTER TABLE public.roles ADD COLUMN code VARCHAR(50);

-- Make code unique and not null
ALTER TABLE public.roles ADD CONSTRAINT roles_code_unique UNIQUE (code);

-- Insert default roles
INSERT INTO public.roles (name, code, description) VALUES
('Super Admin', 'super_admin', 'Full system access'),
('Admin', 'admin', 'Administrative access'),
('Church Staff', 'church_staff', 'Church staff member'),
('Public', 'public', 'Public user')
ON CONFLICT DO NOTHING;
