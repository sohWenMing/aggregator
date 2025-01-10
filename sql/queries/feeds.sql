-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES($1, $2, $3, $4, $5, $6)
RETURNING *;


-- name: ResetFeeds :exec
DELETE from feeds;


-- name: GetFeeds :many
SELECT feeds.name AS FeedName, feeds.url AS FeedUrl, users.name as UserName
  FROM feeds
  JOIN users
    ON feeds.user_id = users.id;


-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES($1, $2, $3, $4, $5)
    RETURNING *
)

SELECT 
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
    FROM inserted_feed_follow
    JOIN feeds
      ON inserted_feed_follow.feed_id = feeds.id
    JOIN users
      ON inserted_feed_follow.user_id = users.id;

-- name: GetFeedIdByURL :one
SELECT feeds.id
  FROM feeds
  WHERE feeds.url = $1;
  

-- name: GetFeedFollowForUser :many
SELECT feeds.name as feed_name, feeds.url as feed_url, users.name as user_name
  FROM feed_follows
  JOIN users
    ON feed_follows.user_id = users.id
  JOIN feeds
    ON feed_follows.feed_id = feeds.id
  WHERE feed_follows.user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE from feed_follows
WHERE feed_follows.user_id = $1
  AND feed_follows.feed_id IN (
    SELECT feed_follows.feed_id
      FROM feed_follows
      JOIN feeds
        ON feed_follows.feed_id = feeds.id
     WHERE feeds.url = $2
  );

-- name: MarkFetchedFeed :exec
UPDATE feeds
SET updated_at = $1, last_fetched_at = $1
WHERE feeds.id = $2;

-- name: GetNextFeedToFetch :one
SELECT feeds.*
  FROM feeds
  ORDER BY feeds.last_fetched_at NULLS FIRST
  LIMIT 1;
 
