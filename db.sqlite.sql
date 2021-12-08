CREATE TABLE IF NOT EXISTS "tbl_user" (
   "Id"	INTEGER PRIMARY KEY AUTOINCREMENT,
   "Username"	TEXT NOT NULL UNIQUE,
   "HashedPassword"	BLOB NOT NULL,
   "IsSuper" TINYINT,
   "CreatedAt" DATETIME DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS "tbl_record" (
  "Id"	INTEGER PRIMARY KEY AUTOINCREMENT,
  "Md5"	TEXT,
  "Content"	TEXT,
  "ContentType" TEXT,
  "OutputContent" TEXT,
  "SrcLang" TEXT,
  "DesLang" TEXT,
  "FileName" TEXT,
  "FileSrcDir" TEXT,
  "FileDesDir" TEXT,
  "State" INTEGER,
  "StateDescribe" TEXT,
  "Error" TEXT,
  "UserId" INTEGER,
  "CreateAt" DATETIME DEFAULT (datetime('now','localtime'))
);

