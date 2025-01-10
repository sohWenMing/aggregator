-- name: CreatePost :one
INSERT INTO posts (
    id, 
    created_at, 
    updated_at, 
    title, 
    url, 
    description,
    published_at,
    feed_id)
VALUES($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostsForUser :many
SELECT 
    posts.title AS post_title,
    posts.url AS post_url,
    posts.description AS post_description,
    posts.published_at AS post_published_on,
    feeds.name AS feed_name
  FROM users
  JOIN feed_follows
    ON users.id = feed_follows.user_id
  JOIN feeds
    ON feed_follows.feed_id = feeds.id
  JOIN posts
    ON feed_follows.feed_id = posts.feed_id
 WHERE users.id = $1
 ORDER BY posts.published_at DESC
  LIMIT $2;
