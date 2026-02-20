-- SQLite schema for xxjz (table prefix xxjz_)
CREATE TABLE IF NOT EXISTS xxjz_account (
  acid INTEGER PRIMARY KEY AUTOINCREMENT,
  acmoney REAL NOT NULL,
  acclassid INTEGER NOT NULL,
  actime INTEGER NOT NULL,
  acremark TEXT NOT NULL DEFAULT '',
  jiid INTEGER NOT NULL,
  zhifu INTEGER NOT NULL,
  fid INTEGER NOT NULL DEFAULT -1
);

CREATE TABLE IF NOT EXISTS xxjz_account_class (
  classid INTEGER PRIMARY KEY AUTOINCREMENT,
  classname TEXT NOT NULL,
  classtype INTEGER NOT NULL,
  ufid INTEGER NOT NULL,
  sort INTEGER NOT NULL DEFAULT 255
);

CREATE TABLE IF NOT EXISTS xxjz_account_funds (
  fundsid INTEGER PRIMARY KEY AUTOINCREMENT,
  fundsname TEXT NOT NULL,
  uid INTEGER NOT NULL,
  sort INTEGER NOT NULL DEFAULT 255
);

CREATE TABLE IF NOT EXISTS xxjz_account_image (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uid INTEGER NOT NULL,
  acid INTEGER,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  size INTEGER NOT NULL,
  ext TEXT NOT NULL,
  md5 TEXT NOT NULL,
  savepath TEXT NOT NULL,
  savename TEXT NOT NULL,
  time INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS xxjz_account_transfer (
  tid INTEGER PRIMARY KEY AUTOINCREMENT,
  uid INTEGER NOT NULL,
  money REAL NOT NULL,
  source_fid INTEGER NOT NULL,
  target_fid INTEGER NOT NULL,
  time INTEGER NOT NULL,
  mark TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS xxjz_user (
  uid INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL,
  password TEXT NOT NULL,
  email TEXT NOT NULL,
  utime INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS xxjz_user_config (
  cid INTEGER PRIMARY KEY AUTOINCREMENT,
  uid INTEGER NOT NULL,
  config_name TEXT NOT NULL,
  config_key TEXT NOT NULL,
  config_value TEXT NOT NULL,
  time INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS xxjz_user_login (
  lid INTEGER PRIMARY KEY AUTOINCREMENT,
  uid INTEGER NOT NULL,
  login_name TEXT NOT NULL,
  login_id TEXT NOT NULL,
  login_key TEXT NOT NULL,
  login_token TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS xxjz_user_push (
  pid INTEGER PRIMARY KEY AUTOINCREMENT,
  uid INTEGER NOT NULL,
  push_name TEXT NOT NULL DEFAULT 'Weixin',
  push_id TEXT NOT NULL,
  push_mark TEXT,
  time INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_account_jiid ON xxjz_account(jiid);
CREATE INDEX IF NOT EXISTS idx_account_actime ON xxjz_account(actime);
CREATE INDEX IF NOT EXISTS idx_class_ufid ON xxjz_account_class(ufid);
CREATE INDEX IF NOT EXISTS idx_funds_uid ON xxjz_account_funds(uid);
CREATE INDEX IF NOT EXISTS idx_transfer_uid ON xxjz_account_transfer(uid);
CREATE INDEX IF NOT EXISTS idx_user_login_id ON xxjz_user_login(login_name, login_id);
