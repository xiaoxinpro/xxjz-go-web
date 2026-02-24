package importsql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMySQLToSQLite_DropCreate(t *testing.T) {
	// Use + to include backticks (raw string cannot contain backtick)
	sql := "DROP TABLE IF EXISTS `xxjz_user`;\n" +
		"CREATE TABLE `xxjz_user` (\n" +
		"  `uid` int(11) NOT NULL AUTO_INCREMENT,\n" +
		"  `username` varchar(32) NOT NULL,\n" +
		"  PRIMARY KEY (`uid`)\n" +
		") ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;\n"
	stmts := MySQLToSQLite(sql)
	assert.GreaterOrEqual(t, len(stmts), 2)
	var createStmt string
	for _, s := range stmts {
		if strings.Contains(s, "CREATE TABLE") {
			createStmt = s
			break
		}
	}
	assert.Contains(t, createStmt, "CREATE TABLE")
	assert.Contains(t, createStmt, "AUTOINCREMENT")
	assert.NotContains(t, createStmt, "ENGINE")
}

func TestMySQLToSQLite_InsertIncluded(t *testing.T) {
	// 模拟 dump：LOCK / 注释 / INSERT / 注释 / UNLOCK，确认 INSERT 被解析并保留
	sql := "LOCK TABLES `jizhang_user` WRITE;\n" +
		"/*!40000 ALTER TABLE `jizhang_user` DISABLE KEYS */;\n" +
		"INSERT INTO `jizhang_user` VALUES (1,'admin','x','a@b.com',1421555782);\n" +
		"/*!40000 ALTER TABLE `jizhang_user` ENABLE KEYS */;\n" +
		"UNLOCK TABLES;\n"
	stmts := MySQLToSQLite(sql)
	var insertFound bool
	for _, s := range stmts {
		if strings.HasPrefix(strings.TrimSpace(s), "INSERT INTO") {
			insertFound = true
			assert.Contains(t, s, "xxjz_user")
			assert.Contains(t, s, "admin")
			break
		}
	}
	assert.True(t, insertFound, "INSERT 语句应被解析并出现在结果中")
}

func TestMySQLToSQLite_InsertWithBOM(t *testing.T) {
	// 带 UTF-8 BOM 的 INSERT 不应被丢弃
	const utf8BOM = "\xef\xbb\xbf"
	sql := utf8BOM + "INSERT INTO `jizhang_user` VALUES (1,'a','b','c',0);\n"
	stmts := MySQLToSQLite(sql)
	assert.Len(t, stmts, 1)
	assert.True(t, strings.Contains(stmts[0], "INSERT INTO"))
	assert.True(t, strings.Contains(stmts[0], "xxjz_user"))
}

func TestMySQLToSQLite_InsertBacktickInValues(t *testing.T) {
	// VALUES 内含有反引号时，只替换 INSERT INTO ... VALUES 前的标识符，不替换值里的反引号，避免产生非法 "
	sql := "INSERT INTO `jizhang_account` VALUES (1,2.5,3,4,'remark`with`backtick',5,6,-1);\n"
	stmts := MySQLToSQLite(sql)
	assert.Len(t, stmts, 1)
	// 表名应为双引号形式
	assert.Contains(t, stmts[0], "\"xxjz_account\"")
	// 值里的反引号应保留，不能变成双引号
	assert.Contains(t, stmts[0], "remark`with`backtick")
	assert.NotContains(t, stmts[0], "remark\"with\"backtick")
}

func TestMySQLToSQLite_InsertEscapedSingleQuote(t *testing.T) {
	// MySQL 用 \' 表示字符串内单引号，需转为 SQLite 的 ''
	sql := "INSERT INTO `jizhang_account` VALUES (1,2.5,3,4,'it\\'s ok',5,6,-1);\n"
	stmts := MySQLToSQLite(sql)
	assert.Len(t, stmts, 1)
	assert.Contains(t, stmts[0], "it''s ok")
	assert.NotContains(t, stmts[0], "\\'")
}
