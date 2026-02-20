package handler

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/config"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/importsql"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/service"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/session"
)

type APIHandler struct {
	cfg       *config.Config
	userSvc   *service.UserService
	statSvc   *service.StatisticService
	fundsSvc  *service.FundsService
	classSvc  *service.ClassService
	db        *sql.DB
}

func NewAPIHandler(cfg *config.Config, userSvc *service.UserService, statSvc *service.StatisticService, fundsSvc *service.FundsService, classSvc *service.ClassService, db *sql.DB) *APIHandler {
	return &APIHandler{cfg: cfg, userSvc: userSvc, statSvc: statSvc, fundsSvc: fundsSvc, classSvc: classSvc, db: db}
}

// getParam returns GET or POST param (API supports both).
func getParam(c *gin.Context, key string) string {
	if c.Request.Method == http.MethodPost {
		if v := c.PostForm(key); v != "" {
			return v
		}
	}
	return c.Query(key)
}

// InitStatus returns whether the app has been initialized (at least one user). No auth. Used by frontend to redirect to /init.
func (h *APIHandler) InitStatus(c *gin.Context) {
	ok, err := h.userSvc.IsInitialized()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"initialized": true, "error": err.Error()}) // assume initialized on error to avoid exposing setup
		return
	}
	c.JSON(http.StatusOK, gin.H{"initialized": ok})
}

// InitSetup creates the first admin user. Only allowed when not initialized. No auth.
func (h *APIHandler) InitSetup(c *gin.Context) {
	initialized, _ := h.userSvc.IsInitialized()
	if initialized {
		c.JSON(http.StatusForbidden, gin.H{"ok": false, "msg": "系统已初始化，请使用管理员登录"})
		return
	}
	var body struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
		Email    string `json:"email" form:"email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		_ = c.ShouldBind(&body) // fallback to form
	}
	if body.Username == "" || body.Password == "" || body.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "msg": "请填写用户名、密码和邮箱"})
		return
	}
	ok, msg, uid := h.userSvc.RegistShell(body.Username, body.Password, body.Email)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"ok": false, "msg": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "msg": "管理员账号创建成功", "uid": uid})
}

// InitImport imports MySQL dump (same as AdminImport). When not initialized, no auth; when initialized, requires admin.
func (h *APIHandler) InitImport(c *gin.Context) {
	initialized, _ := h.userSvc.IsInitialized()
	if initialized {
		sess := sessions.Default(c)
		uid := session.GetUID(sess)
		if uid <= 0 || int64(h.cfg.User.AdminUID) != uid {
			c.JSON(http.StatusForbidden, gin.H{"ok": false, "msg": "需要管理员权限"})
			return
		}
	}
	if h.cfg.Database.Driver != "sqlite" && h.cfg.Database.Driver != "sqlite3" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "msg": "当前仅支持目标为 SQLite 时使用导入"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "msg": "请上传 file 字段的 .sql 文件"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "msg": err.Error()})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 10*1024*1024))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "msg": err.Error()})
		return
	}
	statements := importsql.MySQLToSQLite(string(data))
	if err := importsql.RunSQLiteStatements(h.db, statements); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "msg": "导入失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "msg": "导入成功", "statements": len(statements)})
}

// Login handles POST/GET: username, password, submit -> { uid, uname }
func (h *APIHandler) Login(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	_ = sess.Save()

	username := getParam(c, "username")
	password := getParam(c, "password")
	_ = getParam(c, "submit")

	ok, uid, uname, shell := h.userSvc.UserLogin(username, password)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"uid": 0, "uname": uname})
		return
	}
	sess.Set(session.KeyUID, uid)
	sess.Set(session.KeyUsername, uname)
	sess.Set(session.KeyUserShell, shell)
	if err := sess.Save(); err != nil {
		c.JSON(http.StatusOK, gin.H{"uid": 0, "uname": "登录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"uid": uid, "uname": uname})
}

// Version returns app version and config for clients (no auth).
func (h *APIHandler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title":   h.cfg.App.Title,
		"message": h.cfg.App.Welcome,
		"version": h.cfg.App.Version,
		"wechat":  h.cfg.Wechat.Enable,
		"admin":   gin.H{"uid": h.cfg.User.AdminUID},
		"demo": gin.H{
			"username": h.cfg.User.Demo.Username,
			"password": h.cfg.User.Demo.Password,
		},
		"account": gin.H{
			"PAGE_SIZE":            h.cfg.User.PageSize,
			"MONEY_FORMAT_DECIMALS": h.cfg.Money.FormatDecimals,
			"MONEY_FORMAT_POINT":   h.cfg.Money.FormatPoint,
			"MONEY_FORMAT_THOUSANDS": h.cfg.Money.FormatThousands,
			"MAX_MONEY_VALUE":      h.cfg.Money.MaxValue,
			"MAX_CLASS_NAME":       h.cfg.Limits.MaxClassName,
			"MAX_FUNDS_NAME":       h.cfg.Limits.MaxFundsName,
			"MAX_MARK_VALUE":       h.cfg.Limits.MaxMarkValue,
			"IMAGE_SIZE":           h.cfg.Image.MaxSize,
			"IMAGE_COUNT":          h.cfg.Image.MaxCount,
			"IMAGE_CACHE_URL":      h.cfg.Image.CacheURL,
		},
	})
}

// User handles type=get (get), updataUsername, updataPassword, updataEmail.
func (h *APIHandler) User(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	reqUID, _ := strconv.ParseInt(getParam(c, "uid"), 10, 64)
	typ := getParam(c, "type")
	dataEnc := getParam(c, "data")

	var data map[string]interface{}
	if dataEnc != "" {
		dec, _ := base64.StdEncoding.DecodeString(dataEnc)
		_ = json.Unmarshal(dec, &data)
	}

	out := gin.H{}
	if uid <= 0 || reqUID != uid {
		out["uid"] = 0
		c.JSON(http.StatusOK, out)
		return
	}

	switch typ {
	case "get":
		email, _ := h.userSvc.GetUserEmail(uid, true)
		out["uid"] = uid
		out["username"] = session.GetUsername(sess)
		out["email"] = email
	case "updataUsername":
		username, _ := data["username"].(string)
		email, _ := data["email"].(string)
		password, _ := data["password"].(string)
		ok, msg := h.userSvc.UpdateUsername(uid, username, email, password)
		if ok {
			out["uid"] = uid
			out["username"] = msg
			sess.Set(session.KeyUsername, msg)
			_ = sess.Save()
		} else {
			out["uid"] = 0
			out["username"] = msg
		}
	case "updataPassword":
		oldP, _ := data["old"].(string)
		newP, _ := data["new"].(string)
		ok, msg := h.userSvc.UpdatePassword(uid, oldP, newP)
		if ok {
			out["uid"] = uid
			out["username"] = msg
		} else {
			out["uid"] = 0
			out["username"] = msg
		}
	case "updataEmail":
		out["uid"] = 0
		out["username"] = "邮箱不可修改，请联系管理员！"
	default:
		out["uid"] = 0
	}
	c.JSON(http.StatusOK, out)
}

// Statistic returns { uid, data } where data is AccountStatisticProcess result or error string.
func (h *APIHandler) Statistic(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	out := gin.H{}
	if uid <= 0 {
		out["uid"] = 0
		out["data"] = "用户未登录，请重新登录！"
		c.JSON(http.StatusOK, out)
		return
	}
	typ := getParam(c, "type")
	if typ == "retime" {
		// clear cache would go here; we don't cache yet
	}
	data, err := h.statSvc.AccountStatisticProcess(uid)
	if err != nil {
		out["uid"] = uid
		out["data"] = err.Error()
		c.JSON(http.StatusOK, out)
		return
	}
	out["uid"] = uid
	out["data"] = data
	c.JSON(http.StatusOK, out)
}

// Funds handles type=get, get_id, add, edit, del.
func (h *APIHandler) Funds(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	out := gin.H{}
	if uid <= 0 {
		out["uid"] = 0
		out["data"] = "用户未登录，请重新登录！"
		c.JSON(http.StatusOK, out)
		return
	}
	out["uid"] = uid
	typ := getParam(c, "type")
	dataEnc := getParam(c, "data")
	var data map[string]interface{}
	if dataEnc != "" {
		dec, _ := base64.StdEncoding.DecodeString(dataEnc)
		_ = json.Unmarshal(dec, &data)
	}
	if data == nil {
		data = make(map[string]interface{})
	}

	switch typ {
	case "get":
		v, err := h.fundsSvc.GetFundsData(uid)
		if err != nil {
			out["data"] = err.Error()
		} else {
			out["data"] = v
		}
	case "get_id":
		fid, _ := data["fundsid"].(float64)
		v, err := h.fundsSvc.GetFundsIdData(int64(fid), uid)
		if err != nil {
			out["data"] = err.Error()
		} else {
			out["data"] = v
		}
	case "add":
		name, _ := data["fundsname"].(string)
		money, _ := data["fundsmoney"].(float64)
		v, _ := h.fundsSvc.AddNewFunds(name, money, uid)
		out["data"] = v
	case "edit":
		fid, _ := data["fundsid"].(float64)
		if name, ok := data["fundsname"].(string); ok {
			v, _ := h.fundsSvc.EditFundsName(int64(fid), name, uid)
			out["data"] = v
		} else {
			out["data"] = []interface{}{true, "OK"}
		}
	case "del":
		oldID, _ := data["fundsid_old"].(float64)
		newID, _ := data["fundsid_new"].(float64)
		v, _ := h.fundsSvc.DeleteFunds(int64(oldID), uid, int64(newID))
		out["data"] = v
	default:
		out["data"] = "非法操作！"
	}
	c.JSON(http.StatusOK, out)
}

// Aclass handles type=get, getin, getout, getall, getindata, getoutdata, getalldata, add, edit, del.
func (h *APIHandler) Aclass(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	out := gin.H{}
	if uid <= 0 {
		out["uid"] = 0
		out["data"] = "用户未登录，请重新登录！"
		c.JSON(http.StatusOK, out)
		return
	}
	out["uid"] = uid
	typ := getParam(c, "type")
	dataEnc := getParam(c, "data")
	var data map[string]interface{}
	if dataEnc != "" {
		dec, _ := base64.StdEncoding.DecodeString(dataEnc)
		_ = json.Unmarshal(dec, &data)
	}

	switch typ {
	case "get":
		in, _ := h.classSvc.GetClassData(uid, 1)
		outMap := gin.H{"in": in, "out": nil, "all": nil}
		outMap["out"], _ = h.classSvc.GetClassData(uid, 2)
		outMap["all"], _ = h.classSvc.GetClassData(uid, 0)
		out["data"] = outMap
	case "getin":
		v, _ := h.classSvc.GetClassData(uid, 1)
		out["data"] = v
	case "getout":
		v, _ := h.classSvc.GetClassData(uid, 2)
		out["data"] = v
	case "getall":
		v, _ := h.classSvc.GetClassData(uid, 0)
		out["data"] = v
	case "getindata":
		v, _ := h.classSvc.GetClassAllData(uid, 1)
		out["data"] = v
	case "getoutdata":
		v, _ := h.classSvc.GetClassAllData(uid, 2)
		out["data"] = v
	case "getalldata":
		v, _ := h.classSvc.GetClassAllData(uid, 0)
		out["data"] = v
	case "add":
		name, _ := data["classname"].(string)
		ct, _ := data["classtype"].(float64)
		v, _ := h.classSvc.AddNewClass(name, int(ct), uid)
		out["data"] = v
	case "edit":
		classID, _ := data["classid"].(float64)
		name, _ := data["classname"].(string)
		ct, _ := data["classtype"].(float64)
		v, _ := h.classSvc.EditClassName(int64(classID), name, int(ct), uid)
		out["data"] = v
	case "del":
		classID, _ := data["classid"].(float64)
		v, _ := h.classSvc.DelClass(int64(classID), uid)
		out["data"] = v
	default:
		out["data"] = "非法操作！"
	}
	c.JSON(http.StatusOK, out)
}

// Account stub: returns uid 0 when not logged in; otherwise minimal get/add/edit/del/find stubs.
func (h *APIHandler) Account(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	out := gin.H{}
	if uid <= 0 {
		out["uid"] = 0
		out["data"] = "用户未登录，请重新登录！"
		c.JSON(http.StatusOK, out)
		return
	}
	out["uid"] = uid
	typ := getParam(c, "type")
	dataEnc := getParam(c, "data")
	var data map[string]interface{}
	if dataEnc != "" {
		dec, _ := base64.StdEncoding.DecodeString(dataEnc)
		_ = json.Unmarshal(dec, &data)
	}
	if data == nil {
		data = make(map[string]interface{})
	}
	switch typ {
	case "get":
		out["data"] = map[string]interface{}{"data": []interface{}{}, "page": 1, "pagemax": 1, "count": 0}
	case "get_year", "get_all_year":
		out["data"] = []interface{}{}
	case "get_id":
		out["data"] = map[string]interface{}{}
	case "add":
		out["data"] = map[string]interface{}{"ret": false, "msg": "功能开发中"}
	case "edit", "del":
		out["data"] = map[string]interface{}{"ret": false, "msg": "功能开发中"}
	case "find":
		out["data"] = map[string]interface{}{"ret": true, "msg": map[string]interface{}{"data": []interface{}{}, "page": 1, "pagemax": 1, "count": 0}}
	case "get_image", "set_image", "del_image":
		out["data"] = map[string]interface{}{"ret": false, "msg": "功能开发中"}
	default:
		out["data"] = map[string]interface{}{}
	}
	c.JSON(http.StatusOK, out)
}

// Transfer stub.
func (h *APIHandler) Transfer(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	out := gin.H{}
	if uid <= 0 {
		out["uid"] = 0
		out["data"] = "用户未登录，请重新登录！"
		c.JSON(http.StatusOK, out)
		return
	}
	out["uid"] = uid
	typ := getParam(c, "type")
	switch typ {
	case "get":
		out["data"] = []interface{}{}
	default:
		out["data"] = map[string]interface{}{"ret": false, "msg": "功能开发中"}
	}
	c.JSON(http.StatusOK, out)
}

// Find stub.
func (h *APIHandler) Find(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	out := gin.H{}
	if uid <= 0 {
		out["uid"] = 0
		out["data"] = "用户未登录，请重新登录！"
		c.JSON(http.StatusOK, out)
		return
	}
	out["uid"] = uid
	out["data"] = map[string]interface{}{"ret": true, "msg": map[string]interface{}{"data": []interface{}{}}}
	c.JSON(http.StatusOK, out)
}

// Chart stub: type=year|month, returns JSON.
func (h *APIHandler) Chart(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	if uid <= 0 {
		c.JSON(http.StatusOK, gin.H{"uid": 0})
		return
	}
	typ := getParam(c, "type")
	if typ == "month" {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	c.JSON(http.StatusOK, []interface{}{})
}

// Autocopy stub: admin only get/updata.
func (h *APIHandler) Autocopy(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	if uid <= 0 {
		c.String(http.StatusOK, "非法操作autoCopy.")
		return
	}
	if int64(h.cfg.User.AdminUID) != uid {
		c.String(http.StatusOK, "非法操作autoCopy.")
		return
	}
	typ := getParam(c, "type")
	if typ == "updata" {
		c.JSON(http.StatusOK, gin.H{"strData": "", "enable": false, "enablePullDown": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"strData": "", "enable": false, "enablePullDown": false})
}

// AdminImport imports a MySQL dump file into the current DB (admin only). Target must be SQLite.
func (h *APIHandler) AdminImport(c *gin.Context) {
	sess := sessions.Default(c)
	uid := session.GetUID(sess)
	if uid <= 0 || int64(h.cfg.User.AdminUID) != uid {
		c.JSON(http.StatusForbidden, gin.H{"ok": false, "msg": "需要管理员权限"})
		return
	}
	if h.cfg.Database.Driver != "sqlite" && h.cfg.Database.Driver != "sqlite3" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "msg": "当前仅支持目标为 SQLite 时使用导入"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "msg": "请上传 file 字段的 .sql 文件"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "msg": err.Error()})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 10*1024*1024))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "msg": err.Error()})
		return
	}
	statements := importsql.MySQLToSQLite(string(data))
	if err := importsql.RunSQLiteStatements(h.db, statements); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "msg": "导入失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "msg": "导入成功", "statements": len(statements)})
}