ALTER TABLE users
    DROP COLUMN "isBanned";

CREATE TABLE bans
(
    "userId"          INTEGER                             NOT NULL,
    "isActive"        BOOLEAN   DEFAULT TRUE              NOT NULL,
    "activeUntil"     DATE                                NOT NULL,
    "createdAt"       TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "createdByUserId" INTEGER                             NOT NULL,
    "reason"          VARCHAR(512)                        NOT NULL
);

CREATE INDEX ON bans("userId");