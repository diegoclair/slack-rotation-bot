-- Add role field to scheduler_configs table
ALTER TABLE scheduler_configs ADD COLUMN role TEXT DEFAULT 'On duty';