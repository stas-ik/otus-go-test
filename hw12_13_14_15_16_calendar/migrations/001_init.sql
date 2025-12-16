-- +goose Up
CREATE TABLE IF NOT EXISTS events (
                                      id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    description TEXT,
    user_id VARCHAR(255) NOT NULL,
    notify_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX idx_events_user_time ON events(user_id, start_time, end_time);

CREATE INDEX idx_events_start_time ON events(start_time);

-- +goose Down
DROP TABLE IF EXISTS events;