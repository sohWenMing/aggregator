-- +goose Up
CREATE TABLE feed_follows (
    id uuid PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,    
    user_id uuid NOT NULL,
    feed_id uuid NOT NULL, 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    CONSTRAINT user_id_to_feed_id UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;