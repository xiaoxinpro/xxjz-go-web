package service

import (
	"crypto/md5"
	"encoding/hex"
	"regexp"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

var emailRegex = regexp.MustCompile(`^[-a-zA-Z0-9_.]+@([0-9A-Za-z][0-9A-Za-z-]+\.)+[A-Za-z]{2,5}$`)

type UserService struct {
	cfg *config.Config
	repo *repository.UserRepo
}

func NewUserService(cfg *config.Config, repo *repository.UserRepo) *UserService {
	return &UserService{cfg: cfg, repo: repo}
}

// IsInitialized returns true if at least one user exists (first-run check).
func (s *UserService) IsInitialized() (bool, error) {
	n, err := s.repo.CountUsers()
	return n > 0, err
}

func MD5(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// UserLogin checks credentials and returns (ok, uid, uname, shell) or (false, 0, errorMessage, "").
func (s *UserService) UserLogin(username, password string) (ok bool, uid int64, uname string, shell string) {
	u, err := s.repo.GetByUsername(username)
	if err != nil || u == nil {
		return false, 0, "用户名或密码错误！", ""
	}
	hash := MD5(password)
	if u.Password != hash {
		return false, 0, "用户名或密码错误！", ""
	}
	shell = MD5(u.Username + u.Password)
	return true, u.UID, u.Username, shell
}

// UserShell validates session: md5(username+password) == key.
func (s *UserService) UserShell(username, key string) bool {
	u, err := s.repo.GetByUsername(username)
	if err != nil || u == nil {
		return false
	}
	expected := MD5(u.Username + u.Password)
	return key == expected
}

func (s *UserService) GetUserEmail(uid int64, all bool) (string, error) {
	email, err := s.repo.GetEmail(uid)
	if err != nil {
		return "", err
	}
	if all {
		return email, nil
	}
	// mask: first 2 chars + ... + domain
	if len(email) <= 5 {
		return email, nil
	}
	at := 0
	for i, c := range email {
		if c == '@' {
			at = i
			break
		}
	}
	if at == 0 {
		return email, nil
	}
	before := email[:2]
	if 2 > at {
		before = email[:at]
	}
	return before + "..." + email[at:], nil
}

func (s *UserService) IsDemoUser(username string) bool {
	return username == s.cfg.User.Demo.Username
}

func (s *UserService) RegistShell(username, password, email string) (ok bool, msg string, uid int64) {
	if len(username) < 2 {
		return false, "用户名不合法！", 0
	}
	if len(password) < 4 {
		return false, "密码长度过短，请重新输入新密码！", 0
	}
	if !emailRegex.MatchString(email) {
		return false, "邮箱格式有误，请重新输入邮箱。", 0
	}
	exists, _ := s.repo.UsernameExists(username, 0)
	if exists {
		return false, "用户名已存在，请更换用户名再试！", 0
	}
	exists, _ = s.repo.EmailExists(email)
	if exists {
		return false, "该邮箱已注册过，如忘记密码请尝试找回密码。", 0
	}
	id, err := s.repo.Create(username, MD5(password), email)
	if err != nil {
		return false, "写入数据库出错(>_<)", 0
	}
	return true, "新账号注册成功!", id
}

func (s *UserService) UpdateUsername(uid int64, newUsername, email, password string) (ok bool, msg string) {
	u, err := s.repo.GetByUID(uid)
	if err != nil || u == nil {
		return false, "用户不存在"
	}
	if u.Username == s.cfg.User.Demo.Username {
		return false, "抱歉Demo账号无法进行用户名修改！"
	}
	if MD5(password) != u.Password {
		return false, "验证密码失败，请重新输入登录密码！"
	}
	curEmail, _ := s.repo.GetEmail(uid)
	if curEmail != email {
		return false, "验证邮箱失败，请输入注册时填写的Email。"
	}
	if len(newUsername) < 2 {
		return false, "用户名不合法！"
	}
	exists, _ := s.repo.UsernameExists(newUsername, uid)
	if exists {
		return false, "用户名已存在，请更换用户名再试！"
	}
	if err := s.repo.UpdateUsername(uid, newUsername); err != nil {
		return false, "更新失败"
	}
	return true, newUsername
}

func (s *UserService) UpdatePassword(uid int64, oldPass, newPass string) (ok bool, msg string) {
	u, err := s.repo.GetByUID(uid)
	if err != nil || u == nil {
		return false, "用户不存在"
	}
	if u.Username == s.cfg.User.Demo.Username {
		return false, "抱歉Demo账号无法进行密码修改！"
	}
	if MD5(oldPass) != u.Password {
		return false, "验证密码失败，请重新输入登录密码！"
	}
	if len(newPass) < 6 {
		return false, "密码长度过短，请重新输入新密码！"
	}
	if err := s.repo.UpdatePassword(uid, MD5(newPass)); err != nil {
		return false, "更新失败"
	}
	return true, u.Username
}
