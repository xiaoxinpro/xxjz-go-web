package importsql

import (
	"database/sql"
	"log"
	"regexp"
	"strings"
)

// MySQLToSQLite converts MySQL dump DDL/DML to SQLite-compatible SQL.
// Returns a slice of statements; for each CREATE TABLE, a DROP TABLE IF EXISTS is prepended so import is idempotent.
func MySQLToSQLite(mysql string) []string {
	var out []string
	lines := strings.Split(mysql, "\n")
	var current strings.Builder
	flush := func() {
		s := strings.TrimSpace(current.String())
		if s != "" && !strings.HasPrefix(s, "--") {
			stmts := convertOneMySQLToSQLite(s)
			out = append(out, stmts...)
		}
		current.Reset()
	}
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			// 空行也追加到缓冲区，避免多行 INSERT 被截断
			current.WriteString("\n")
			continue
		}
		// 遇到要跳过的行时，先 flush 已积累的语句，避免把下一句和未 flush 的 INSERT 拼在一起
		if strings.HasPrefix(trimmed, "SET ") ||
			strings.HasPrefix(trimmed, "/*!") ||
			strings.HasPrefix(strings.ToUpper(trimmed), "LOCK TABLES") ||
			strings.HasPrefix(strings.ToUpper(trimmed), "UNLOCK TABLES") {
			if current.Len() > 0 {
				flush()
			}
			continue
		}
		current.WriteString(line)
		// 按“语句结束”判断：整行去掉空白后以 ; 结尾才 flush，避免行尾空格导致 INSERT 不 flush
		if strings.HasSuffix(trimmed, ";") {
			current.WriteString("\n")
			flush()
		} else {
			current.WriteString("\n")
		}
	}
	flush()
	return out
}

// knownTableSuffixes: known table name suffixes (longest first) to detect user's prefix.
var knownTableSuffixes = []string{
	"account_transfer", "account_image", "account_class", "account_funds",
	"user_config", "user_login", "user_push", "account", "user",
}

// detectTablePrefix extracts table name from DROP/CREATE/INSERT statement and returns prefix (e.g. "jizhang_" or "xxjz_").
// Returns empty if not found or already xxjz_. App expects xxjz_ so other prefixes will be normalized.
func detectTablePrefix(s string) string {
	// First backticked identifier in statement is typically the table name
	re := regexp.MustCompile("`([^`]+)`")
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return ""
	}
	tableName := matches[1]
	for _, suffix := range knownTableSuffixes {
		if strings.HasSuffix(tableName, suffix) {
			prefix := strings.TrimSuffix(tableName, suffix)
			if prefix != "" && strings.HasSuffix(prefix, "_") {
				return prefix
			}
			if prefix != "" {
				return prefix + "_"
			}
			break
		}
	}
	return ""
}

// convertOneMySQLToSQLite converts one MySQL statement to SQLite. Returns one or more statements;
// for CREATE TABLE it returns [DROP TABLE IF EXISTS tablename, CREATE TABLE ...] so import is idempotent.
func convertOneMySQLToSQLite(s string) []string {
	s = strings.TrimSpace(s)
	// 去掉 UTF-8 BOM，否则首条语句可能被误判为非 INSERT/CREATE/DROP 而丢弃
	const utf8BOM = "\xef\xbb\xbf"
	s = strings.TrimPrefix(s, utf8BOM)
	// Normalize table prefix: user may use jizhang_, xxjz_, or custom prefix; app expects xxjz_.
	if prefix := detectTablePrefix(s); prefix != "" && prefix != "xxjz_" {
		s = strings.ReplaceAll(s, "`"+prefix, "`xxjz_")
	}
	upper := strings.ToUpper(s)
	// DROP TABLE -> keep
	if strings.HasPrefix(upper, "DROP TABLE") {
		s = regexp.MustCompile("`([^`]+)`").ReplaceAllString(s, "$1")
		return []string{strings.TrimSpace(s)}
	}
	// CREATE TABLE
	if strings.HasPrefix(upper, "CREATE TABLE") {
		s = regexp.MustCompile("`([^`]+)`").ReplaceAllString(s, "\"$1\"")
		s = regexp.MustCompile(`\bint\(\d+\)\s+unsigned\b`).ReplaceAllString(s, "INTEGER")
		s = regexp.MustCompile(`\bint\s+unsigned\b`).ReplaceAllString(s, "INTEGER")
		s = regexp.MustCompile(`\bint\(\d+\)`).ReplaceAllString(s, "INTEGER")
		s = regexp.MustCompile(`\bint\b`).ReplaceAllString(s, "INTEGER")
		s = regexp.MustCompile(`\bdouble\(\d+,\d+\)\s+unsigned\b`).ReplaceAllString(s, "REAL")
		s = regexp.MustCompile(`\bdouble\(\d+,\d+\)`).ReplaceAllString(s, "REAL")
		s = regexp.MustCompile(`\bvarchar\(\d+\)`).ReplaceAllString(s, "TEXT")
		s = regexp.MustCompile(`\bunsigned\b`).ReplaceAllString(s, "")
		s = regexp.MustCompile(`AUTO_INCREMENT`).ReplaceAllString(s, "AUTOINCREMENT")
		// SQLite: AUTOINCREMENT only with INTEGER PRIMARY KEY on the same column.
		s = regexp.MustCompile(`\"([^\"]+)\"\s+INTEGER\s+NOT NULL\s+AUTOINCREMENT`).ReplaceAllString(s, "\"$1\" INTEGER PRIMARY KEY AUTOINCREMENT")
		s = regexp.MustCompile(`\"([^\"]+)\"\s+INTEGER\s+AUTOINCREMENT`).ReplaceAllString(s, "\"$1\" INTEGER PRIMARY KEY AUTOINCREMENT")
		s = regexp.MustCompile(`,\s*PRIMARY KEY\s*\(\s*\"[^\"]+\"\s*\)`).ReplaceAllString(s, "")
		s = regexp.MustCompile(`,\s*KEY\s+\"[^\"]*\"\s*\([^)]+\)`).ReplaceAllString(s, "")
		s = regexp.MustCompile(`,\s*UNIQUE KEY\s+\"[^\"]*\"\s*\([^)]+\)`).ReplaceAllString(s, "")
		s = regexp.MustCompile(`DEFAULT '-1'`).ReplaceAllString(s, "DEFAULT -1")
		s = regexp.MustCompile(`DEFAULT '(\d+)'`).ReplaceAllString(s, "DEFAULT $1")
		s = regexp.MustCompile(`\s*ENGINE=\w+\s*.*$`).ReplaceAllString(s, "")
		s = regexp.MustCompile(`\s*CHARACTER SET \w+`).ReplaceAllString(s, "")
		s = regexp.MustCompile(`\s*COLLATE \w+`).ReplaceAllString(s, "")
		s = strings.TrimSpace(s)
		// Prepend DROP TABLE IF EXISTS so import works even when tables already exist (e.g. from migrations).
		if m := regexp.MustCompile(`CREATE TABLE\s+"([^"]+)"`).FindStringSubmatch(s); len(m) >= 2 {
			return []string{"DROP TABLE IF EXISTS " + m[1], s}
		}
		return []string{s}
	}
	// INSERT: 只对 "INSERT INTO ... VALUES" 之前的标识符部分做反引号替换；VALUES 部分做 MySQL→SQLite 转义
	if strings.HasPrefix(upper, "INSERT INTO") {
		idx := strings.Index(upper, " VALUES ")
		if idx >= 0 {
			identPart := s[:idx]
			valuesPart := s[idx:]
			identPart = regexp.MustCompile("`([^`]+)`").ReplaceAllString(identPart, "\"$1\"")
			// MySQL 字符串内用 \' 表示单引号，SQLite 用 ''；不转换会导致 SQLite 把 \' 后的 ' 当成字符串结束，产生非法 token
			valuesPart = strings.ReplaceAll(valuesPart, "\\'", "''")
			s = identPart + valuesPart
		} else {
			s = regexp.MustCompile("`([^`]+)`").ReplaceAllString(s, "\"$1\"")
		}
		return []string{strings.TrimSpace(s)}
	}
	return nil
}

// RunSQLiteStatements executes statements on db (SQLite). Skips empty and errors on first real error.
// On error, logs the failing statement index, a short snippet, and for long INSERTs the approximate location of problematic characters.
func RunSQLiteStatements(db *sql.DB, statements []string) error {
	const snippetLen = 300
	const contextWindow = 100 // 定位问题时前后各显示的字节数
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			snippet := stmt
			if len(snippet) > snippetLen {
				snippet = snippet[:snippetLen] + "..."
			}
			log.Printf("[import] statement [%d/%d] failed: %v\n  snippet: %s", i+1, len(statements), err, snippet)
			// 超长 INSERT 时在 VALUES 部分定位可能引发 "unrecognized token" 的字符（" 或 \）
			if len(stmt) > 5000 && strings.Contains(strings.ToUpper(stmt), "INSERT INTO") && strings.Contains(stmt, " VALUES ") {
				logInsertProblemLocation(stmt, contextWindow)
			}
			return err
		}
	}
	return nil
}

// logInsertProblemLocation 在 INSERT 的 VALUES 部分查找首个 " 或 \，输出其偏移与前后文，便于在 62 万字节等超长语句中定位问题。
func logInsertProblemLocation(stmt string, window int) {
	upper := strings.ToUpper(stmt)
	idx := strings.Index(upper, " VALUES ")
	if idx < 0 {
		return
	}
	valuesPart := stmt[idx:]
	for j, r := range valuesPart {
		if r == '"' || r == '\\' {
			start := j - window
			if start < 0 {
				start = 0
			}
			end := j + window
			if end > len(valuesPart) {
				end = len(valuesPart)
			}
			context := valuesPart[start:end]
			// 避免 log 里出现不可打印字符
			context = strings.ToValidUTF8(context, "?")
			log.Printf("[import] first problematic char in VALUES at byte offset ~%d (relative to VALUES): %q\n  context: ...%s...", j, r, context)
			return
		}
	}
	// 未找到 " 或 \，可能错误在别处，仅输出 VALUES 长度
	log.Printf("[import] INSERT VALUES length: %d bytes, no \" or \\ found in VALUES part", len(valuesPart))
}

// EnsureImportSchema 在导入后为旧版 MySQL 表补齐缺失的 sort 列（当前迁移表有 sort，导入的 dump 可能没有）。
func EnsureImportSchema(db *sql.DB) error {
	for _, table := range []string{"xxjz_account_class", "xxjz_account_funds"} {
		rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
		if err != nil {
			log.Printf("[import] EnsureImportSchema: PRAGMA table_info(%s) err=%v", table, err)
			continue
		}
		hasSort := false
		for rows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dflt sql.NullString
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				rows.Close()
				return err
			}
			if name == "sort" {
				hasSort = true
				break
			}
		}
		rows.Close()
		if hasSort {
			continue
		}
		_, err = db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN sort INTEGER NOT NULL DEFAULT 255`)
		if err != nil {
			if strings.Contains(err.Error(), "no such table") {
				log.Printf("[import] EnsureImportSchema: table %s not found, skip", table)
				continue
			}
			log.Printf("[import] EnsureImportSchema: ALTER TABLE %s ADD COLUMN sort err=%v", table, err)
			return err
		}
		log.Printf("[import] EnsureImportSchema: added column sort to %s", table)
	}
	return nil
}

// NormalizeImportOwnership 将分类表、资金账户表全部归属到「首个用户」（xxjz_user 最小 uid）。
// 导入流程已不再调用：导入保留 dump 中的 ufid/uid，多用户各自看到自己的数据。
// 仅在你需要「把所有分类/资金账户划给某一个用户」时，可手动调用或执行等价 SQL。
func NormalizeImportOwnership(db *sql.DB) error {
	var firstUID int64
	err := db.QueryRow(`SELECT uid FROM xxjz_user ORDER BY uid LIMIT 1`).Scan(&firstUID)
	if err != nil {
		log.Printf("[import] NormalizeImportOwnership: get first user failed: %v", err)
		return err
	}
	log.Printf("[import] NormalizeImportOwnership: firstUID=%d", firstUID)
	resClass, err := db.Exec(`UPDATE xxjz_account_class SET ufid = ?`, firstUID)
	if err != nil {
		log.Printf("[import] NormalizeImportOwnership: UPDATE account_class failed: %v", err)
		return err
	}
	nClass, _ := resClass.RowsAffected()
	log.Printf("[import] NormalizeImportOwnership: xxjz_account_class updated ufid=%d, rows=%d", firstUID, nClass)
	resFunds, err := db.Exec(`UPDATE xxjz_account_funds SET uid = ?`, firstUID)
	if err != nil {
		log.Printf("[import] NormalizeImportOwnership: UPDATE account_funds failed: %v", err)
		return err
	}
	nFunds, _ := resFunds.RowsAffected()
	log.Printf("[import] NormalizeImportOwnership: xxjz_account_funds updated uid=%d, rows=%d", firstUID, nFunds)
	return nil
}
