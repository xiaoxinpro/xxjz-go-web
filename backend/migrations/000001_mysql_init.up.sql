-- MySQL schema for xxjz
CREATE TABLE IF NOT EXISTS xxjz_account (
  acid int(11) unsigned NOT NULL AUTO_INCREMENT,
  acmoney double(9,2) unsigned NOT NULL,
  acclassid int(11) NOT NULL,
  actime int(11) NOT NULL,
  acremark varchar(255) NOT NULL DEFAULT '',
  jiid int(11) NOT NULL,
  zhifu int(11) NOT NULL,
  fid int(11) NOT NULL DEFAULT -1,
  PRIMARY KEY (acid),
  KEY idx_jiid (jiid),
  KEY idx_actime (actime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS xxjz_account_class (
  classid int(11) NOT NULL AUTO_INCREMENT,
  classname varchar(255) NOT NULL,
  classtype int(1) NOT NULL,
  ufid int(11) NOT NULL,
  sort int(11) NOT NULL DEFAULT 255,
  PRIMARY KEY (classid),
  KEY idx_ufid (ufid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS xxjz_account_funds (
  fundsid int(11) NOT NULL AUTO_INCREMENT,
  fundsname varchar(255) NOT NULL,
  uid int(11) NOT NULL,
  sort int(11) NOT NULL DEFAULT 255,
  PRIMARY KEY (fundsid),
  KEY idx_uid (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS xxjz_account_image (
  id int(11) unsigned NOT NULL AUTO_INCREMENT,
  uid int(11) unsigned NOT NULL,
  acid int(11) unsigned DEFAULT NULL,
  name varchar(32) NOT NULL,
  type varchar(32) NOT NULL,
  size int(11) unsigned NOT NULL,
  ext varchar(8) NOT NULL,
  md5 varchar(32) NOT NULL,
  savepath varchar(32) NOT NULL,
  savename varchar(32) NOT NULL,
  time int(11) unsigned NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS xxjz_account_transfer (
  tid int(11) unsigned NOT NULL AUTO_INCREMENT,
  uid int(11) unsigned NOT NULL,
  money double(9,2) unsigned NOT NULL,
  source_fid int(11) unsigned NOT NULL,
  target_fid int(11) unsigned NOT NULL,
  time int(11) unsigned NOT NULL,
  mark varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (tid),
  KEY idx_uid (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS xxjz_user (
  uid int(11) NOT NULL AUTO_INCREMENT,
  username varchar(32) NOT NULL,
  password varchar(32) NOT NULL,
  email varchar(255) NOT NULL,
  utime int(11) NOT NULL,
  PRIMARY KEY (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS xxjz_user_config (
  cid int(11) NOT NULL AUTO_INCREMENT,
  uid int(11) NOT NULL,
  config_name varchar(32) NOT NULL,
  config_key varchar(32) NOT NULL,
  config_value varchar(32) NOT NULL,
  time int(11) NOT NULL,
  PRIMARY KEY (cid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS xxjz_user_login (
  lid int(11) NOT NULL AUTO_INCREMENT,
  uid int(11) NOT NULL,
  login_name varchar(32) NOT NULL,
  login_id varchar(32) NOT NULL,
  login_key varchar(32) NOT NULL,
  login_token varchar(32) NOT NULL,
  PRIMARY KEY (lid),
  KEY idx_login (login_name, login_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS xxjz_user_push (
  pid int(11) NOT NULL AUTO_INCREMENT,
  uid int(11) NOT NULL,
  push_name varchar(32) NOT NULL DEFAULT 'Weixin',
  push_id varchar(64) NOT NULL,
  push_mark varchar(32) DEFAULT NULL,
  time int(11) NOT NULL,
  PRIMARY KEY (pid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
