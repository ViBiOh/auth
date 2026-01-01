ALTER TABLE auth."user"
  RENAME COLUMN creation_date TO creation;

ALTER TABLE auth.profile
  RENAME COLUMN creation_date TO creation;

ALTER TABLE auth.user_profile
  RENAME COLUMN creation_date TO creation;

ALTER TABLE auth.basic
  RENAME COLUMN creation_date TO creation;

ALTER TABLE auth.github
  RENAME COLUMN creation_date TO creation;

ALTER TABLE auth.discord
  RENAME COLUMN creation_date TO creation;

CREATE TABLE auth.user_link (
  external_id TEXT                     NOT NULL,
  token       TEXT                     NOT NULL,
  description TEXT                     NOT NULL,
  user_id     TEXT                              REFERENCES auth.user(id) ON DELETE CASCADE,
  creation    TIMESTAMP WITH TIME ZONE          DEFAULT now()
);

CREATE UNIQUE INDEX user_link_external_id ON auth.user_link(external_id);
CREATE        INDEX user_link_user_id     ON auth.user_link(user_id);

