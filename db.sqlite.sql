CREATE TABLE IF NOT EXISTS "users" (
	"id"	TEXT NOT NULL UNIQUE,
	"DisplayName"	TEXT NOT NULL,
	"Mail"	TEXT NOT NULL UNIQUE,
	"JobTitle"	TEXT,
	"UserPrincipalName"	TEXT,
	"AccessToken"	TEXT NOT NULL UNIQUE,
	"AccessTokenActive"	INTEGER,
	"RefreshToken"	TEXT,
	PRIMARY KEY("id")
);

--  Id              string
-- 	User            string
-- 	Subject         string
-- 	SenderEmail     string
-- 	SenderName      string
-- 	HasAttachments  bool
-- 	BodyPreview     string
-- 	BodyType        string
-- 	BodyContent     string
-- 	ToRecipient     string
-- 	ToRecipientName string
CREATE TABLE IF NOT EXISTS "mails" (
   "id"	TEXT NOT NULL UNIQUE,
   "User"	TEXT ,
   "Subject"	TEXT ,
   "SenderEmail"	TEXT ,
   "SenderName"	TEXT ,
   "Attachments"	TEXT ,
   "BodyPreview"	TEXT ,
   "BodyType"	TEXT ,
   "BodyContent"	TEXT,
   "ToRecipient"	TEXT,
   "ToRecipientName"	TEXT,
   "Date"       DATETIME,
   PRIMARY KEY("id")
);

CREATE TABLE IF NOT EXISTS "attachment" (
    "attachmentId"	TEXT,
    "mailId"	TEXT ,
    "filename"	TEXT ,
    "filepath" TEXT,
    PRIMARY KEY("attachmentId")
);