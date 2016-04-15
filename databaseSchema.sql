CREATE TABLE "device" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "mac" TEXT NOT NULL,
    "username" TEXT NOT NULL,
    "regIP" TEXT DEFAULT '',
    "platform" TEXT DEFAULT '',
    "subnet" TEXT DEFAULT '',
    "expires" INTEGER DEFAULT (0),
    "dateRegistered" INTEGER NOT NULL,
    "userAgent" TEXT DEFAULT ''
);

CREATE TABLE "user" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "username" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "deviceLimit" INTEGER DEFAULT (10),
    "expires" INTEGER DEFAULT (0),
    "canManage" INTEGER DEFAULT (1),
    "validAfter" INTEGER DEFAULT (0),
    "validBefore" INTEGER DEFAULT (0)
);

INSERT INTO "user" ("username", "password") VALUES ("admin", "$2a$10$qTSqBy7YI8YVMNT0Ozl99uImx4jEYgUKJrA4qJcnffMmpOB3mOcEq");

CREATE TABLE "blacklist" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "value" TEXT NOT NULL UNIQUE ON CONFLICT IGNORE
);
