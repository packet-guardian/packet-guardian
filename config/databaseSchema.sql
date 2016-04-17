DROP TABLE "device";
CREATE TABLE "device" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "mac" TEXT NOT NULL,
    "username" TEXT NOT NULL,
    "registered_from" TEXT DEFAULT '',
    "platform" TEXT DEFAULT '',
    "expires" INTEGER DEFAULT (0),
    "date_registered" INTEGER NOT NULL,
    "user_agent" TEXT DEFAULT '',
    "blacklisted" INTEGER DEFAULT (0)
);

DROP TABLE "user";
CREATE TABLE "user" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "username" TEXT NOT NULL,
    "password" TEXT DEFAULT '',
    "device_limit" INTEGER DEFAULT (-1),
    "default_expiration" INTEGER DEFAULT (0),
    "expiration_type" INTEGER DEFAULT (1),
    "can_manage" INTEGER DEFAULT (1),
    "valid_start" INTEGER DEFAULT (0),
    "valid_end" INTEGER DEFAULT (0),
    "valid_forever" INTEGER DEFAULT (1)
);

INSERT INTO "user" ("username", "password") VALUES ("admin", "$2a$10$rZfN/gdXZdGYyLtUb6LF.eHOraDes3ibBECmWic2I3SocMC0L2Lxa");
INSERT INTO "user" ("username", "password") VALUES ("helpdesk", "$2a$10$ICCdq/OyZBBoNPTRmfgntOnujD6INGv7ZAtA/Xq6JIdRMO65xCuNC");

DROP TABLE "blacklist";
CREATE TABLE "blacklist" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "value" TEXT NOT NULL UNIQUE ON CONFLICT IGNORE
);
