CREATE TABLE accounts
(
    id         bigserial PRIMARY KEY,
    owner      varchar NOT NULL,
    currency   varchar NOT NULL,
    balance    bigint  NOT NULL,
    created_at timestamptz DEFAULT (now())
);


CREATE TABLE users
(
    username            varchar PRIMARY KEY,
    password            varchar NOT NULL,
    email               varchar NOT NULL UNIQUE,
    full_name           varchar NOT NULL,
    role                varchar NOT NULL DEFAULT 'user',
    password_changed_at timestamptz      DEFAULT '0001-01-01 00:00:00z',
    created_at          timestamptz      DEFAULT (now())
);

CREATE INDEX ON accounts (owner);

ALTER TABLE accounts
    ADD CONSTRAINT "owner_currency_key" UNIQUE (owner, currency);

ALTER TABLE accounts
    ADD FOREIGN KEY (owner) REFERENCES users (username);

