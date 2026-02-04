-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    last_seen TIMESTAMP NOT NULL
);

-- Create user_activities table
CREATE TABLE IF NOT EXISTS user_activities (
    feed_id VARCHAR(255) PRIMARY KEY,
    action_text_template TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create user_activity_subject_referring table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS user_activity_subject_referring (
    feed_id VARCHAR(255) NOT NULL,
    referring_type VARCHAR(50) NOT NULL,
    referring_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    PRIMARY KEY (feed_id, referring_id),
    FOREIGN KEY (feed_id) REFERENCES user_activities(feed_id) ON DELETE CASCADE
);

-- Create user_activity_object_referring table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS user_activity_object_referring (
    feed_id VARCHAR(255) NOT NULL,
    referring_type VARCHAR(50) NOT NULL,
    referring_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    PRIMARY KEY (feed_id, referring_id),
    FOREIGN KEY (feed_id) REFERENCES user_activities(feed_id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_user_activity_subject_referring_feed_id ON user_activity_subject_referring(feed_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_object_referring_feed_id ON user_activity_object_referring(feed_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_subject_referring_id ON user_activity_subject_referring(referring_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_object_referring_id ON user_activity_object_referring(referring_id);
