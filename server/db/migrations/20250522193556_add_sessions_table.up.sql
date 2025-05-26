CREATE TABLE sessions (
    "id" varchar(255) PRIMARY KEY NOT NULL,
    "email" varchar(255) NOT NULL,
    "refresh_token" varchar(512) NOT NULL,
    "is_revoked" BOOLEAN NOT NULL DEFAULT false,
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "expires_at" TIMESTAMP NOT NULL
);