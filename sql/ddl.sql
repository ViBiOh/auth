CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- clean
DROP TABLE IF EXISTS auth.login_profile;
DROP TABLE IF EXISTS auth.profile;
DROP TABLE IF EXISTS auth.login;

DROP SEQUENCE IF EXISTS auth.profile_seq;
DROP SEQUENCE IF EXISTS auth.login_seq;

DROP INDEX IF EXISTS login_profile_login_id;
DROP INDEX IF EXISTS profile_id;
DROP INDEX IF EXISTS profile_id;
DROP INDEX IF EXISTS login_login;

DROP SCHEMA IF EXISTS auth;

-- schema
CREATE SCHEMA auth;

-- user
CREATE SEQUENCE auth.login_seq;
CREATE TABLE auth.login (
  id BIGINT NOT NULL DEFAULT nextval('auth.login_seq'),
  login TEXT NOT NULL,
  password TEXT NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE auth.login_seq OWNED BY auth.login.id;

CREATE UNIQUE INDEX login_id ON auth.login(id);
CREATE UNIQUE INDEX login_login ON auth.login(login);

-- profile
CREATE SEQUENCE auth.profile_seq;
CREATE TABLE auth.profile (
  id BIGINT NOT NULL DEFAULT nextval('auth.profile_seq'),
  name TEXT NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE auth.profile_seq OWNED BY auth.profile.id;

CREATE UNIQUE INDEX profile_id ON auth.profile(id);

-- login_profile
CREATE TABLE auth.login_profile (
  login_id BIGINT NOT NULL REFERENCES auth.login(id) ON DELETE CASCADE,
  profile_id BIGINT NOT NULL REFERENCES auth.profile(id) ON DELETE CASCADE,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE UNIQUE INDEX login_profile_login_id ON auth.login_profile(login_id);
