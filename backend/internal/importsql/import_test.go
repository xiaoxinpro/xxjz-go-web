package importsql

import (
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
	assert.Contains(t, stmts[0], "DROP TABLE")
	assert.Contains(t, stmts[1], "CREATE TABLE")
	assert.Contains(t, stmts[1], "AUTOINCREMENT")
	assert.NotContains(t, stmts[1], "ENGINE")
}
