CREATE EXTENSION IF NOT EXISTS citext;

CREATE UNLOGGED TABLE users (
    nickname CITEXT PRIMARY KEY,
    fullname TEXT NOT NULL,
    about TEXT,
    email CITEXT UNIQUE
);

CREATE UNLOGGED TABLE forum (
    slug    CITEXT PRIMARY KEY,
    title   TEXT,
    "user"  CITEXT,
    posts   BIGINT DEFAULT 0,
    threads BIGINT DEFAULT 0,

    FOREIGN KEY ("user") REFERENCES "users" (nickname)
);

CREATE UNLOGGED TABLE thread (
    id      SERIAL PRIMARY KEY,
    author  CITEXT,
    created TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    forum   CITEXT,
    message TEXT NOT NULL,
    slug    CITEXT UNIQUE,
    title   TEXT NOT NULL,
    votes   INT                      DEFAULT 0,

    FOREIGN KEY (author) REFERENCES "users" (nickname),
    FOREIGN KEY (forum)  REFERENCES "forum" (slug)
);

CREATE UNLOGGED TABLE post (
    id       BIGSERIAL PRIMARY KEY,
    author   CITEXT NOT NULL,
    created  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    forum    CITEXT,
    message  TEXT   NOT NULL,
    isEdited BOOLEAN                  DEFAULT FALSE,
    parent   BIGINT                   DEFAULT 0,
    thread   INT,
    Path     BIGINT[]                 DEFAULT ARRAY []::INTEGER[],

    FOREIGN KEY (author) REFERENCES "users"  (nickname),
    FOREIGN KEY (forum)  REFERENCES "forum"  (slug),
    FOREIGN KEY (thread) REFERENCES "thread" (id),
    FOREIGN KEY (parent) REFERENCES "post"   (id)
);

CREATE UNLOGGED TABLE votes
(
    nickname  citext,
    voice     INT,
    thread_id INT,

    FOREIGN KEY (nickname) REFERENCES "users" (nickname),
    FOREIGN KEY (thread_id) REFERENCES "thread" (id),
    UNIQUE (nickname, thread_id)
);

CREATE UNLOGGED TABLE users_forum
(
    nickname CITEXT NOT NULL,
    fullname TEXT NOT NULL,
    about    TEXT,
    email    CITEXT,
    slug     citext NOT NULL,

    FOREIGN KEY (nickname) REFERENCES "users" (nickname),
    FOREIGN KEY (slug) REFERENCES "forum" (slug),
    UNIQUE (nickname, slug)
);

CREATE INDEX all_users_forum ON users_forum (nickname, fullname, about, email);
CLUSTER users_forum USING all_users_forum;
CREATE INDEX nickname_users_forum ON users_forum using hash (nickname);
CREATE INDEX f_a_e_users_forum ON users_forum (fullname, about, email);

CREATE OR REPLACE FUNCTION update_user_forum() RETURNS TRIGGER AS
$update_users_forum$
DECLARE
    m_fullname CITEXT;
    m_about    CITEXT;
    m_email CITEXT;
BEGIN
    SELECT fullname, about, email FROM users WHERE nickname = NEW.author INTO m_fullname, m_about, m_email;
    INSERT INTO users_forum (nickname, fullname, about, email, slug)
    VALUES (NEW.author, m_fullname, m_about, m_email, NEW.forum) on conflict do nothing;
    return NEW;
end
$update_users_forum$
LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION insert_votes() RETURNS TRIGGER AS
$update_users_forum$
BEGIN
    UPDATE thread SET votes = (votes + NEW.voice) WHERE id=NEW.thread_id;
    return NEW;
end
$update_users_forum$
LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION update_votes() RETURNS TRIGGER AS
$update_users_forum$
BEGIN
    IF OLD.voice <> NEW.voice THEN
        UPDATE thread SET votes = (votes + NEW.Voice*2) WHERE id=NEW.thread_id;
    END IF;
    return NEW;
end
$update_users_forum$
LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION update_count_of_threads() RETURNS TRIGGER AS
$update_users_forum$
BEGIN
    UPDATE forum SET threads = threads + 1 WHERE slug=NEW.forum;
    return NEW;
end
$update_users_forum$
LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION update_path() RETURNS TRIGGER AS
$update_path$
DECLARE
    parentPath         BIGINT[];
    first_parent_thread INT;
BEGIN
    IF (NEW.parent IS NULL) THEN
        NEW.path := array_append(new.path, new.id);
    ELSE
        SELECT path FROM post WHERE id = new.parent INTO parentPath;
        SELECT thread FROM post WHERE id = parentPath[1] INTO first_parent_thread;
        IF NOT FOUND OR first_parent_thread != NEW.thread THEN
            RAISE EXCEPTION 'parent is from different thread' USING ERRCODE = '00409';
        end if;

        NEW.path := NEW.path || parentPath || new.id;
    end if;
    UPDATE forum SET posts = posts + 1 WHERE forum.slug = new.forum;
    RETURN new;
end
$update_path$
LANGUAGE plpgsql;

CREATE INDEX path_ ON post (path);

CREATE TRIGGER add_thread_in_forum
    BEFORE INSERT
    ON thread
    FOR EACH ROW EXECUTE PROCEDURE update_count_of_threads();

CREATE TRIGGER add_voice
    BEFORE INSERT
    ON votes
    FOR EACH ROW EXECUTE PROCEDURE insert_votes();

CREATE TRIGGER edit_voice
    BEFORE UPDATE
    ON votes
    FOR EACH ROW EXECUTE PROCEDURE update_votes();

CREATE TRIGGER update_path_trigger
    BEFORE INSERT
    ON post
    FOR EACH ROW EXECUTE PROCEDURE update_path();

CREATE TRIGGER thread_insert_user_forum
    AFTER INSERT
    ON thread
    FOR EACH ROW EXECUTE PROCEDURE update_user_forum();

CREATE TRIGGER post_insert_user_forum
    AFTER INSERT
    ON post
    FOR EACH ROW EXECUTE PROCEDURE update_user_forum();


CREATE INDEX IF NOT EXISTS user_nickname  ON users using hash (nickname);
CREATE INDEX IF NOT EXISTS user_email     ON users using hash (email);
CREATE INDEX IF NOT EXISTS forum_slug     ON forum using hash (slug);
CREATE INDEX IF NOT EXISTS thr_slug       ON thread using hash (slug);
CREATE INDEX IF NOT EXISTS thr_date       ON thread (created);
CREATE INDEX IF NOT EXISTS thr_forum      ON thread using hash (forum);
CREATE INDEX IF NOT EXISTS thr_forum_date ON thread (forum, created);
CREATE INDEX IF NOT EXISTS post_id_path   ON post (id, (path[1]));
CREATE INDEX IF NOT EXISTS post_path1     ON post ((path[1]));
CREATE INDEX IF NOT EXISTS post_thread_id ON post (thread, id);
CREATE INDEX IF NOT EXISTS post_thr_id    ON post (thread);
CREATE INDEX IF NOT EXISTS post_thread_id_path1_parent ON post (thread, id, (path[1]), parent);
CREATE INDEX IF NOT EXISTS post_thread_path_id ON post (thread, path, id);
CREATE INDEX IF NOT EXISTS post_path1_path_id_desc ON post ((path[1]) DESC, path, id);
CREATE INDEX IF NOT EXISTS post_path1_path_id_asc ON post ((path[1]) DESC, path, id);

CREATE UNIQUE INDEX IF NOT EXISTS  vote_unique ON votes (nickname, thread_id);
CREATE UNIQUE INDEX IF NOT EXISTS  forum_users_unique ON users_forum (slug, nickname);
