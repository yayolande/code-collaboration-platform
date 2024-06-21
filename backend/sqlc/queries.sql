
-- name: GetUsers :many
SELECT * FROM users ;

-- name: AddUser :one
INSERT INTO users (username, password, email, status) VALUES (?, ?, ?, ?)
RETURNING * ;

-- name: LoginUser :one
SELECT * FROM users 
  WHERE username = ? AND password = ? 
  LIMIT 1 ;

-- name: AddPost :one
INSERT INTO posts (user_id, language_id, code, comment, post_date)
   VALUES (?, ?, ?, ?, ?) RETURNING * ;

-- name: AddPostIntoTree :one
INSERT INTO posts_tree (post_id, parent_post_id)
   VALUES (?, ?) RETURNING * ;

-- name: GetPostsFromRoot :many
SELECT * FROM posts_tree t 
   INNER JOIN posts p ON t.post_id = p.post_id 
   WHERE t.post_id = ? OR t.parent_post_id = ?
   ORDER BY t.parent_post_id ASC ;
   /* Later on, the post should also return the user info (username, user_id) */
   /* INNER JOIN users u ON u.user_id = p.user_id */
   /* Not working now because there is no user in db yet */
