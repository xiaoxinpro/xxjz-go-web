package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

func TestMD5(t *testing.T) {
	h := MD5("admin888")
	assert.Equal(t, "7fef6171469e80d32c0559f88b377245", h)
}

func TestUserLogin_Invalid(t *testing.T) {
	// UserLogin requires a non-nil repo (calls GetByUsername). Skip without DB.
	t.Skip("UserLogin requires repository; run integration test with DB")
}

func TestIsDemoUser(t *testing.T) {
	cfg := &config.Config{User: config.UserConfig{Demo: config.DemoAccount{Username: "demo", Password: "x"}}}
	svc := NewUserService(cfg, nil, nil)
	assert.True(t, svc.IsDemoUser("demo"))
	assert.False(t, svc.IsDemoUser("admin"))
}

func TestRegistShell_Validation(t *testing.T) {
	cfg := &config.Config{}
	svc := NewUserService(cfg, nil, nil)
	ok, msg, _ := svc.RegistShell("a", "123", "bad-email")
	assert.False(t, ok)
	assert.NotEmpty(t, msg)
	ok, msg, _ = svc.RegistShell("ab", "12", "a@b.com")
	assert.False(t, ok)
	assert.Contains(t, msg, "密码")
}

// Integration-style test with real DB would use repository with :memory: SQLite
func TestUserService_WithRepo(t *testing.T) {
	// Skip if no DB; optional: use sqlite :memory: + repository
	t.Skip("optional: use in-memory SQLite for full login test")
	_ = repository.NewUserRepo(nil)
}
