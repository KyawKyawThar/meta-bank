CREATE TABLE sessions
(
    id            uuid PRIMARY KEY,
    username      VARCHAR     NOT NULL,
    refresh_token VARCHAR     NOT NULL,
    user_agent    VARCHAR     NOT NULL,
    client_ip     VARCHAR     NOT NULL,
    is_blocked    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    timestamptz  NOT NULL DEFAULT (now()),
    expired_at    timestamptz NOT NULL
);

ALTER TABLE sessions
    ADD FOREIGN KEY (username) REFERENCES users ("username");