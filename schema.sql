DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS stop;
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title TEXT,
    content TEXT NOT NULL,
    pub_time TIMESTAMP DEFAULT now(),
    link TEXT UNIQUE
);
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    news_id INT REFERENCES posts(id) ON DELETE CASCADE,
    parent_id INT REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL DEFAULT 'empty',
    PubTime BIGINT NOT NULL DEFAULT extract (epoch from now())
);
CREATE TABLE IF NOT EXISTS stop (
    id SERIAL PRIMARY KEY,
    stop_list TEXT
);
INSERT INTO comments(news_id,content)  VALUES (1,'комментарий');
INSERT INTO comments(news_id,content)  VALUES (2,'ups  проверка');
INSERT INTO stop (stop_list) VALUES ('qwerty');
INSERT INTO stop (stop_list) VALUES ('йцукен');
INSERT INTO stop (stop_list) VALUES ('zxvbnm');