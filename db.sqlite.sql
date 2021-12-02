CREATE TABLE IF NOT EXISTS "tbl_user" (
   "Id"	INTEGER PRIMARY KEY AUTOINCREMENT,
   "Username"	TEXT NOT NULL UNIQUE,
   "HashedPassword"	BLOB NOT NULL,
   "IsSuper" TINYINT,
   "CreatedAt" TIMESTAMP DEFAULT (datetime(CURRENT_TIMESTAMP,'localtime'))
);
