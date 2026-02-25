package service

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/loginlock"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/mail"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/repository"
)

var emailRegex = regexp.MustCompile(`^[-a-zA-Z0-9_.]+@([0-9A-Za-z][0-9A-Za-z-]+\.)+[A-Za-z]{2,5}$`)

const msgLocked = "你的账号已被锁定，请联系管理员解锁！"

type UserService struct {
	cfg       *config.Config
	repo      *repository.UserRepo
	lockStore *loginlock.Store
}

func NewUserService(cfg *config.Config, repo *repository.UserRepo, lockStore *loginlock.Store) *UserService {
	return &UserService{cfg: cfg, repo: repo, lockStore: lockStore}
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
// Uses login lock: after LoginTimes failed attempts (per username, TTL 3600s), returns locked message.
func (s *UserService) UserLogin(username, password string) (ok bool, uid int64, uname string, shell string) {
	if s.lockStore != nil && s.cfg.User.LoginTimes > 0 {
		if s.lockStore.Count(username) >= s.cfg.User.LoginTimes {
			return false, 0, msgLocked, ""
		}
	}
	u, err := s.repo.GetByUsername(username)
	if err != nil || u == nil {
		if s.lockStore != nil && s.cfg.User.LoginTimes > 0 {
			s.lockStore.Add(username)
		}
		return false, 0, "用户名或密码错误！", ""
	}
	hash := MD5(password)
	if u.Password != hash {
		if s.lockStore != nil && s.cfg.User.LoginTimes > 0 {
			s.lockStore.Add(username)
		}
		return false, 0, "用户名或密码错误！", ""
	}
	if s.lockStore != nil {
		s.lockStore.Clear(username)
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

// Forget token: base64(username + "." + md5(username+password+endtime) + "." + endtime)
const forgetTokenValidSec = 7200 // 2h

func (s *UserService) RequestPasswordReset(email string) (ok bool, msg string) {
	if !emailRegex.MatchString(email) {
		return false, "邮箱格式不正确！"
	}
	if !mail.IsConfigured(&s.cfg.Mail) {
		return false, "找回密码功能未开放"
	}
	u, err := s.repo.GetByEmail(strings.TrimSpace(email))
	if err != nil || u == nil {
		return false, "该邮箱未注册过账号！"
	}
	endtime := time.Now().Unix() + forgetTokenValidSec
	endtimeStr := strconv.FormatInt(endtime, 10)
	checkCode := MD5(u.Username + "+" + u.Password + "+" + endtimeStr)
	token := base64.StdEncoding.EncodeToString([]byte(u.Username + "." + checkCode + "." + endtimeStr))
	baseURL := strings.TrimRight(s.cfg.App.BaseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:5173" // dev fallback
	}
	resetLink := baseURL + "/#/forget/reset?p=" + token
	subject := "找回密码 - 小歆记账APP"
	body := "<br>" + u.Username + "：<br />请点击下面的链接，按流程进行密码重设。（两小时内有效）<br><a href=\"" + resetLink + "\">确认密码找回</a></br><pre>" + resetLink + "</pre></br>"
	if err := mail.Send(&s.cfg.Mail, u.Email, subject, body); err != nil {
		return false, "服务器出错，请稍后再试！"
	}
	return true, "找回密码的链接已发送至您的邮箱！"
}

// VerifyForgetToken parses p and checks expiry and checkCode. Returns username and true if valid.
func (s *UserService) VerifyForgetToken(p string) (username string, ok bool) {
	dec, err := base64.StdEncoding.DecodeString(p)
	if err != nil {
		return "", false
	}
	parts := strings.SplitN(string(dec), ".", 3)
	if len(parts) != 3 {
		return "", false
	}
	username = strings.TrimSpace(parts[0])
	endtime, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
	if err != nil || time.Now().Unix() > endtime {
		return "", false
	}
	u, err := s.repo.GetByUsername(username)
	if err != nil || u == nil {
		return "", false
	}
	endtimeStr := parts[2]
	checkCode := MD5(u.Username + "+" + u.Password + "+" + endtimeStr)
	if parts[1] != checkCode {
		return "", false
	}
	return username, true
}

// ResetPasswordWithToken resets password using valid token p. New password min 4 chars.
func (s *UserService) ResetPasswordWithToken(p, newPassword string) (ok bool, msg string) {
	if len(newPassword) < 4 {
		return false, "密码格式错误！"
	}
	username, valid := s.VerifyForgetToken(p)
	if !valid {
		return false, "找回密码链接错误或已过期，请重新获取链接或联系管理员！"
	}
	u, _ := s.repo.GetByUsername(username)
	if u == nil {
		return false, "用户不存在"
	}
	if err := s.repo.UpdatePassword(u.UID, MD5(newPassword)); err != nil {
		return false, "更新失败"
	}
	return true, "OK，修改成功！"
}

// GetUserByWeixinOpenID returns uid, username, shell for the user bound to this openid, or (0, "", "") if not bound.
func (s *UserService) GetUserByWeixinOpenID(openid string) (uid int64, uname string, shell string) {
	uid, err := s.repo.GetUIDByWeixinOpenID(openid)
	if err != nil || uid <= 0 {
		return 0, "", ""
	}
	u, err := s.repo.GetByUID(uid)
	if err != nil || u == nil {
		return 0, "", ""
	}
	return u.UID, u.Username, MD5(u.Username+u.Password)
}

// WeixinBind binds openid to uid. Returns (true, nil) or (false, error message).
func (s *UserService) WeixinBind(uid int64, openid, sessionKey, unionid string) (ok bool, msg string) {
	existing, _ := s.repo.GetUIDByWeixinOpenID(openid)
	if existing > 0 {
		return false, "绑定出错，该微信已绑定。"
	}
	if err := s.repo.InsertWeixinLogin(uid, openid, sessionKey, unionid); err != nil {
		return false, "绑定失败"
	}
	return true, "绑定成功"
}

// WeixinRegistBind runs RegistShell then binds openid to the new user. Returns (ok, msg, uid).
func (s *UserService) WeixinRegistBind(username, password, email, openid, sessionKey, unionid string) (ok bool, msg string, uid int64) {
	existing, _ := s.repo.GetUIDByWeixinOpenID(openid)
	if existing > 0 {
		return false, "绑定出错，该微信被已绑定。", 0
	}
	ok, msg, uid = s.RegistShell(username, password, email)
	if !ok {
		return false, msg, 0
	}
	bindOk, bindMsg := s.WeixinBind(uid, openid, sessionKey, unionid)
	if !bindOk {
		return false, bindMsg, uid
	}
	return true, "注册并绑定成功。", uid
}
