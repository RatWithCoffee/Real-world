DROP TABLE IF EXISTS user_data;
DROP TABLE IF EXISTS session;

CREATE TABLE user_data
(
    id         SERIAL PRIMARY KEY,
    email      VARCHAR     NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    username   VARCHAR     NOT NULL,
    bio        VARCHAR,
    password   bytea       NOT NULL,
    salt       bytea       NOT NULL
);

CREATE TABLE session
(
    session_token varchar,
    user_id       int
);
CREATE INDEX token_index ON session (session_token);


CREATE TABLE article
(
    id          SERIAL PRIMARY KEY,
    slug        VARCHAR     not null,
    user_id     INTEGER REFERENCES user_data (id),
    body        VARCHAR     not null,
    title       VARCHAR     not null,
    description VARCHAR     not null,
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL
);

CREATE INDEX article_creator_index ON article (user_id);
CREATE INDEX article_slug_index ON article (slug);

CREATE TABLE tag
(
    id   SERIAL PRIMARY KEY,
    text VARCHAR NOT NULL
);

INSERT INTO tag (text) VALUES ('golang'),  ('testing'),  ('gomock'),  ('halflife3'), ('coursera');

CREATE TABLE article_tags
(
    article_id INTEGER REFERENCES article (id),
    tag_id     INTEGER REFERENCES tag (id)
);

CREATE INDEX article_index ON article_tags (article_id);
CREATE INDEX tag_index ON article_tags (tag_id);
