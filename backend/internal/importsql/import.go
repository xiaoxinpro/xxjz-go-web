package importsql

import (
	"database/sql"
	"regexp"
	"strings"
)

// MySQLToSQLite converts MySQL dump DDL/DML to SQLite-compatible SQL.
func MySQLToSQLite(mysql string) []string {
	var out []string
	lines := strings.Split(mysql, "\n")
	var current strings.Builder
	flush := func() {
		s := strings.TrimSpace(current.String())
		if s != "" && !strings.HasPrefix(s, "--") {
			s = convertOneMySQLToSQLite(s)
			if s != "" {
				out = append(out, s)
			}
		}
		current.Reset()
	}
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Skip MySQL-specific
		if strings.HasPrefix(strings.TrimSpace(line), "SET ") ||
			strings.HasPrefix(strings.TrimSpace(line), "/*!") {
			continue
		}
		if strings.HasSuffix(line, ";") {
			current.WriteString(line)
			flush()
		} else {
			current.WriteString(line)
			current.WriteString("\n")
		}
	}
	flush()
	return out
}

func convertOneMySQLToSQLite(s string) string {
	upper := strings.ToUpper(s)
	// DROP TABLE -> keep
	if strings.HasPrefix(upper, "DROP TABLE") {
		s = regexp.MustCompile("`([^`]+)`").ReplaceAllString(s, "$1")
		return s
	}
	// CREATE TABLE
	if strings.HasPrefix(upper, "CREATE TABLE") {
		s = regexp.MustCompile("`([^`]+)`").ReplaceAllString(s, "\"$1\"")
		s = regexp.MustCompile(`\bint\(\d+\)\s+unsigned\b`).ReplaceAllString(s, "INTEGER")
		s = regexp.MustCompile(`\bint\(\d+\)`).ReplaceAllString(s, "INTEGER")
		s = regexp.MustCompile(`\bdouble\(\d+,\d+\)\s+unsigned\b`).ReplaceAllString(s, "REAL")
		s = regexp.MustCompile(`\bdouble\(\d+,\d+\)`).ReplaceAllString(s, "REAL")
		s = regexp.MustCompile(`\bvarchar\(\d+\)`).ReplaceAllString(s, "TEXT")
		s = regexp.MustCompile(`\bunsigned\b`).ReplaceAllString(s, "")
		s = regexp.MustCompile(`AUTO_INCREMENT`).ReplaceAllString(s, "AUTOINCREMENT")
		s = regexp.MustCompile(`DEFAULT '-1'`).ReplaceAllString(s, "DEFAULT -1")
		s = regexp.MustCompile(`DEFAULT '(\d+)'`).ReplaceAllString(s, "DEFAULT $1")
		s = regexp.MustCompile(`\s*ENGINE=\w+\s*.*$`).ReplaceAllString(s, "")
		return strings.TrimSpace(s)
	}
	// INSERT
	if strings.HasPrefix(upper, "INSERT INTO") {
		s = regexp.MustCompile("`([^`]+)`").ReplaceAllString(s, "\"$1\"")
		return s
	}
	return ""
}

// RunSQLiteStatements executes statements on db (SQLite). Skips empty and errors on first real error.
func RunSQLiteStatements(db *sql.DB, statements []string) error {
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
