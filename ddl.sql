CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- clean
DROP TABLE IF EXISTS login;
DROP TABLE IF EXISTS profile;

DROP SEQUENCE IF EXISTS login_seq;
DROP SEQUENCE IF EXISTS profile_seq;

DROP INDEX IF EXISTS login_id;
DROP INDEX IF EXISTS login_login;
DROP INDEX IF EXISTS profile_id;

-- user
CREATE SEQUENCE login_seq;
CREATE TABLE login (
  id BIGINT NOT NULL DEFAULT nextval('login_seq'),
  login TEXT NOT NULL,
  password TEXT NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE login_seq OWNED BY login.id;

CREATE UNIQUE INDEX login_id ON login(id);
CREATE UNIQUE INDEX login_login ON login(login);

-- profile
CREATE SEQUENCE profile_seq;
CREATE TABLE profile (
  id BIGINT NOT NULL DEFAULT nextval('profile_seq'),
  name TEXT NOT NULL,
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);
ALTER SEQUENCE profile_seq OWNED BY profile.id;

CREATE UNIQUE INDEX profile_id ON profile(id);

-- login_profile
CREATE TABLE login_profile (
  login_id BIGINT NOT NULL REFERENCES login(id),
  profile_id BIGINT NOT NULL REFERENCES profile(id),
  creation_date TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE UNIQUE INDEX login_profile_login_id ON login_profile(login_id);

-- data
INSERT INTO profile (name) VALUES ('admin');
