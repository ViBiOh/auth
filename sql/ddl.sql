-- clean
DROP TABLE IF EXISTS auth.discord;
DROP TABLE IF EXISTS auth.github;
DROP TABLE IF EXISTS auth.basic;
DROP TABLE IF EXISTS auth.user_profile;
DROP TABLE IF EXISTS auth.profile;
DROP TABLE IF EXISTS auth.user;

DROP INDEX IF EXISTS discord_login;
DROP INDEX IF EXISTS discord_user_id;
DROP INDEX IF EXISTS github_login;
DROP INDEX IF EXISTS github_user_id;
DROP INDEX IF EXISTS basic_user_id;
DROP INDEX IF EXISTS user_profile_user_id;
DROP INDEX IF EXISTS profile_id;
DROP INDEX IF EXISTS user_id;

DROP SCHEMA IF EXISTS auth CASCADE;

-- schema
CREATE SCHEMA auth;

-- user
CREATE TABLE auth.user (
  id            TEXT                     NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE          DEFAULT now()
);

CREATE UNIQUE INDEX user_id ON auth.user(id);

-- profile
CREATE TABLE auth.profile (
  id            TEXT                     NOT NULL,
  name          TEXT                     NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE          DEFAULT now()
);

CREATE UNIQUE INDEX profile_id ON auth.profile(id);

-- user_profile
CREATE TABLE auth.user_profile (
  user_id       TEXT                     NOT NULL REFERENCES auth.user(id)    ON DELETE CASCADE,
  profile_id    TEXT                     NOT NULL REFERENCES auth.profile(id) ON DELETE CASCADE,
  creation_date TIMESTAMP WITH TIME ZONE          DEFAULT now()
);

CREATE UNIQUE INDEX user_profile_user_id ON auth.user_profile(user_id);

-- basic
CREATE TABLE auth.basic (
  user_id       TEXT                     NOT NULL REFERENCES auth.user(id) ON DELETE CASCADE,
  login         TEXT                     NOT NULL,
  password      TEXT                     NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE          DEFAULT now()
);

CREATE UNIQUE INDEX basic_user_id ON auth.basic(user_id);

-- github
CREATE TABLE auth.github (
  user_id       TEXT                     NOT NULL REFERENCES auth.user(id) ON DELETE CASCADE,
  id            BIGINT                   NOT NULL,
  login         TEXT                     NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE          DEFAULT now()
);

CREATE UNIQUE INDEX github_user_id ON auth.github(user_id);
CREATE INDEX github_login   ON auth.github(login);

-- discord
CREATE TABLE auth.discord (
  user_id       TEXT                     NOT NULL REFERENCES auth.user(id) ON DELETE CASCADE,
  id            TEXT                     NOT NULL,
  username      TEXT                     NOT NULL,
  avatar        TEXT                     NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE          DEFAULT now()
);

CREATE UNIQUE INDEX discord_user_id ON auth.discord(user_id);
CREATE INDEX discord_login   ON auth.discord(id);
