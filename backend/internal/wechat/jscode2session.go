package wechat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const jscode2sessionURL = "https://api.weixin.qq.com/sns/jscode2session"

// JSCode2SessionResponse is the WeChat API response.
type JSCode2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// JSCode2Session exchanges js_code for openid and session_key. appid and secret are WeChat mini-program credentials.
func JSCode2Session(appid, secret, jsCode string) (openid, sessionKey, unionid string, err error) {
	if appid == "" || secret == "" {
		return "", "", "", fmt.Errorf("wechat not configured")
	}
	u, _ := url.Parse(jscode2sessionURL)
	q := u.Query()
	q.Set("appid", appid)
	q.Set("secret", secret)
	q.Set("js_code", jsCode)
	q.Set("grant_type", "authorization_code")
	u.RawQuery = q.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}
	var data JSCode2SessionResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return "", "", "", err
	}
	if data.ErrCode != 0 {
		return "", "", "", fmt.Errorf("wechat api: %d %s", data.ErrCode, data.ErrMsg)
	}
	return data.OpenID, data.SessionKey, data.UnionID, nil
}
