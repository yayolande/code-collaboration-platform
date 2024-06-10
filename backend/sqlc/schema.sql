
CREATE TABLE IF NOT EXISTS users (
  user_id INTEGER PRIMARY KEY AUTOINCREMENT,
  username VARCHAR(255) NOT NULL,
  password TEXT NOT NULL,
  status INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS posts (
  post_id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  language_id INTEGER NOT NULL,
  code TEXT NOT NULL,
  comment TEXT NOT NULL,
  post_date TEXT NOT NULL,

  FOREIGN KEY (language_id) REFERENCES code_languages (language_id),
  FOREIGN KEY (user_id) REFERENCES users (user_id)
);

-- 'parent_post_id': contain a post_id to which to respond. 
-- 'parent_post_id': If the value is '-1', then this post_id is the root
CREATE TABLE IF NOT EXISTS posts_tree (
  post_id INTEGER NOT NULL,
  parent_post_id INTEGER NOT NULL DEFAULT -1,

  PRIMARY KEY (post_id, parent_post_id),
  FOREIGN KEY (post_id) REFERENCES posts (post_id),
  FOREIGN KEY (parent_post_id) REFERENCES posts (post_id)
);

CREATE TABLE IF NOT EXISTS code_languages (
  language_id INTEGER PRIMARY KEY AUTOINCREMENT,
  label VARCHAR(255) NOT NULL
);
