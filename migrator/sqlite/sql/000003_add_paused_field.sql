-- Add is_paused field to channels table for pause/resume functionality
ALTER TABLE channels ADD COLUMN is_paused BOOLEAN DEFAULT 0;

-- Create index for efficient querying of active non-paused channels
CREATE INDEX IF NOT EXISTS idx_channels_active_paused ON channels(is_active, is_paused);