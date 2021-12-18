CREATE TABLE IF NOT EXISTS "tbl_user" (
   "Id"	INTEGER PRIMARY KEY AUTOINCREMENT,
   "Username"	TEXT NOT NULL UNIQUE,
   "HashedPassword"	BLOB NOT NULL,
   "IsSuper" TINYINT,
   "CreatedAt" DATETIME DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS "tbl_record" (
  "Id"	INTEGER PRIMARY KEY AUTOINCREMENT,
  "Sha1"	TEXT,
  "Content"	TEXT,
  "ContentType" TEXT,
  "TransType" INTEGER,
  "OutputContent" TEXT,
  "SrcLang" TEXT,
  "DesLang" TEXT,
  "FileName" TEXT,
  "DirRandId" TEXT,
  "State" INTEGER,
  "StateDescribe" TEXT,
  "Error" TEXT,
  "UserId" INTEGER,
  "CreateAt" DATETIME DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS tbl_record_sha1_idx ON tbl_record(Sha1);
