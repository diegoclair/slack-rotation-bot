-- Add last_presenter column to users table
ALTER TABLE users ADD COLUMN last_presenter BOOLEAN DEFAULT 0;

-- Create index for better performance
CREATE INDEX IF NOT EXISTS idx_users_last_presenter ON users(channel_id, last_presenter);

-- Drop rotations table as it's no longer needed
DROP TABLE IF EXISTS rotations;

-- Remove unused indexes
DROP INDEX IF EXISTS idx_rotations_channel_date;
DROP INDEX IF EXISTS idx_rotations_user;