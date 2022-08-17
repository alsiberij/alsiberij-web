CREATE TABLE bans
(
    "bannedUserId"    INTEGER                             NOT NULL,
    "reason"          VARCHAR(512)                        NOT NULL,
    "activeUntil"     TIMESTAMP                           NOT NULL,
    "createdByUserId" INTEGER                             NOT NULL,
    "createdAt"       TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX ON bans ("bannedUserId");