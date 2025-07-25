-- Create scheduler_configs table
CREATE TABLE IF NOT EXISTS scheduler_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL UNIQUE,
    notification_time TEXT DEFAULT '09:00',
    active_days TEXT DEFAULT '[1,2,3,4,5]',
    is_enabled BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
);

-- Create index for efficient querying of enabled scheduler configs
CREATE INDEX IF NOT EXISTS idx_scheduler_configs_enabled ON scheduler_configs(is_enabled);

-- Remove scheduler-related fields from channels table (if they exist)
-- Note: SQLite doesn't support DROP COLUMN, so we'll leave them for backward compatibility
-- but won't use them in the new architecture