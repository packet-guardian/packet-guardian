SET SESSION sql_mode = "ANSI,TRADITIONAL";
DROP TABLE IF EXISTS "device";
CREATE TABLE "device" (
    "id" INTEGER PRIMARY KEY AUTO_INCREMENT,
    "mac" TEXT NOT NULL UNIQUE KEY,
    "username" TEXT NOT NULL,
    "registered_from" TEXT,
    "platform" TEXT,
    "expires" INTEGER DEFAULT 0,
    "date_registered" INTEGER NOT NULL,
    "user_agent" TEXT,
    "blacklisted" INTEGER DEFAULT 0,
    "description" TEXT,
    "last_seen" INTEGER NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1;

DROP TABLE IF EXISTS "user";
CREATE TABLE "user" (
    "id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
    "username" TEXT NOT NULL UNIQUE KEY,
    "password" TEXT,
    "device_limit" INTEGER DEFAULT -1,
    "default_expiration" INTEGER DEFAULT 0,
    "expiration_type" INTEGER DEFAULT 1,
    "can_manage" INTEGER DEFAULT 1,
    "valid_start" INTEGER DEFAULT 0,
    "valid_end" INTEGER DEFAULT 0,
    "valid_forever" INTEGER DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=4;

INSERT INTO "user" ("id", "username", "password") VALUES (1, 'admin', '$2a$10$rZfN/gdXZdGYyLtUb6LF.eHOraDes3ibBECmWic2I3SocMC0L2Lxa');
INSERT INTO "user" ("id", "username", "password") VALUES (2, 'helpdesk', '$2a$10$ICCdq/OyZBBoNPTRmfgntOnujD6INGv7ZAtA/Xq6JIdRMO65xCuNC');
INSERT INTO "user" ("id", "username", "password") VALUES (3, 'readonly', '$2a$10$02NG6kQV.4UicpCnz8hyeefBD4JHKAlZToL2K0EN1HV.u6sXpP1Xy');

DROP TABLE IF EXISTS "blacklist";
CREATE TABLE "blacklist" (
    "id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
    "value" VARCHAR(255) NOT NULL UNIQUE KEY,
    "comment" TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1;

DROP TABLE IF EXISTS "lease";
CREATE TABLE "lease" (
    "id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
    "ip" TEXT NOT NULL UNIQUE KEY,
    "mac" TEXT NOT NULL,
    "network" TEXT NOT NULL,
    "start" INTEGER NOT NULL,
    "end" INTEGER NOT NULL,
    "hostname" TEXT NOT NULL,
    "abandoned" INTEGER DEFAULT 0,
    "registered" INTEGER DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1;
