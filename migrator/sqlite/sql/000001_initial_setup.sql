-- Create channels table
CREATE TABLE IF NOT EXISTS channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slack_channel_id TEXT UNIQUE NOT NULL,
    slack_channel_name TEXT NOT NULL,
    slack_team_id TEXT NOT NULL,
    notification_time TEXT DEFAULT '09:00',
    active_days TEXT DEFAULT '["Monday","Tuesday","Thursday","Friday"]',
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL,
    slack_user_id TEXT NOT NULL,
    slack_user_name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    UNIQUE(channel_id, slack_user_id)
);

-- Create rotations table
CREATE TABLE IF NOT EXISTS rotations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    presented_at DATE NOT NULL,
    was_presenter BOOLEAN DEFAULT 1,
    skipped_reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_channels_active ON channels(is_active);
CREATE INDEX IF NOT EXISTS idx_channels_team ON channels(slack_team_id);
CREATE INDEX IF NOT EXISTS idx_users_channel ON users(channel_id, is_active);
CREATE INDEX IF NOT EXISTS idx_rotations_channel_date ON rotations(channel_id, presented_at);
CREATE INDEX IF NOT EXISTS idx_rotations_user ON rotations(user_id);