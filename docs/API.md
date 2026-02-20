# 小歆记账 API 接口说明

与旧版 ThinkPHP 接口兼容，供 Web 前端与微信小程序调用。支持 GET 与 POST；部分接口使用 `type` + base64 编码的 JSON `data` 传参。

## 基础

- **Base URL**: 同部署域名，如 `https://example.com`
- **路径**: `/api/xxx` 或 `/Home/Api/xxx`（兼容）
- **认证**: Cookie Session（登录后由服务端写 Cookie）
- **响应**: JSON，多数为 `{ "uid": number, "data"?: any, "uname"?: string }`；未登录时 `uid` 为 0，`data` 可能为错误提示字符串。

---

## 0. 首次初始化（无需认证）

- **GET /api/init/status**  
  - 返回 `{ "initialized": true|false }`。`false` 表示尚未创建任何用户，前端应跳转至初始化页。
- **POST /api/init/setup**  
  - 仅当 `initialized === false` 时可调用。请求体 JSON：`username`, `password`, `email`。创建第一个用户（管理员）。  
  - 成功：`{ "ok": true, "msg": "管理员账号创建成功", "uid": 1 }`；失败：`{ "ok": false, "msg": "..." }`。
- **POST /api/init/import**  
  - 未初始化时：无需登录，上传 `file`（.sql 文件），将旧 MySQL 导出转换为当前库并导入。  
  - 已初始化后：需管理员登录，行为同 `POST /api/admin/import`。  
  - 成功：`{ "ok": true, "msg": "导入成功", "statements": n }`。

---

## 1. 登录

- **URL**: `/api/login` 或 `/Home/Api/login`
- **方法**: GET / POST
- **参数**: `username`, `password`, `submit`（可选）
- **响应**:
  - 成功: `{ "uid": 1, "uname": "用户名" }`
  - 失败: `{ "uid": 0, "uname": "用户名或密码错误！" }` 或锁定提示

---

## 2. 版本与配置

- **URL**: `/api/version` 或 `/Home/Api/version`
- **方法**: GET / POST
- **无需登录**
- **响应**: `{ "title", "message", "version", "wechat", "admin", "demo", "account": { "PAGE_SIZE", "MONEY_FORMAT_DECIMALS", ... } }`

---

## 3. 用户信息

- **URL**: `/api/user` 或 `/Home/Api/user`
- **参数**: `uid`, `type`, `data`（base64 编码的 JSON，部分 type 需要）
- **type**:
  - `get`: 返回 uid, username, email
  - `updataUsername`: data 含 username, email, password
  - `updataPassword`: data 含 old, new
  - `updataEmail`: 返回“邮箱不可修改”
- **响应**: `{ "uid", "username"?, "email"?: string }`

---

## 4. 统计

- **URL**: `/api/statistic` 或 `/Home/Api/statistic`
- **参数**: `type`（可选，如 `retime` 表示刷新缓存）
- **响应**: `{ "uid", "data": { "TodayDate", "TodayInMoney", "TodayOutMoney", "MonthInMoney", "MonthOutMoney", "YearInMoney", "YearOutMoney", "SumInMoney", "SumOutMoney", ... } }` 或未登录时 `data` 为字符串提示

---

## 5. 资金账户

- **URL**: `/api/funds` 或 `/Home/Api/funds`
- **参数**: `type`, `data`（base64 JSON）
- **type**: `get` | `get_id` | `add` | `edit` | `del`
- **data 示例**:
  - get_id: `{ "fundsid": 1 }`
  - add: `{ "fundsname": "现金", "fundsmoney": 0 }`
  - edit: `{ "fundsid": 1, "fundsname"?: "新名", "fundsmoney"?: 100 }`
  - del: `{ "fundsid_old": 1, "fundsid_new": 2 }`
- **响应**: `{ "uid", "data": array | object | [ok, msg] }`

---

## 6. 分类

- **URL**: `/api/aclass` 或 `/Home/Api/aclass`
- **参数**: `type`, `data`（base64 JSON，add/edit/del 等需要）
- **type**: `get` | `getin` | `getout` | `getall` | `getindata` | `getoutdata` | `getalldata` | `add` | `edit` | `del`
- **get 响应**: `{ "uid", "data": { "in": { "classid": "分类名" }, "out": {...}, "all": {...} } }`
- **add**: data 含 `classname`, `classtype`（1 收入 / 2 支出）
- **edit**: data 含 `classid`, `classname`, `classtype`
- **del**: data 含 `classid`

---

## 7. 记账

- **URL**: `/api/account` 或 `/Home/Api/account`
- **参数**: `type`, `data`（base64 JSON）
- **type**: `get` | `get_year` | `get_all_year` | `get_id` | `add` | `edit` | `del` | `find` | `get_image` | `set_image` | `del_image`
- **get**: data 含分页与时间范围（如 gettype, year, month, day, page）
- **add**: data 含 acmoney, acclassid, actime, acremark, zhifu, fid 等
- **响应**: `{ "uid", "data": { "ret"?, "msg"?, "data"?, "page"?, "pagemax"?, "count"? } }`

---

## 8. 转账

- **URL**: `/api/transfer` 或 `/Home/Api/transfer`
- **参数**: `type`, `data`（base64 JSON）
- **type**: `get` | `get_id` | `add` | `edit` | `del` | `find`
- **响应**: `{ "uid", "data": array | { "ret", "msg" } }`

---

## 9. 搜索

- **URL**: `/api/find` 或 `/Home/Api/find`
- **参数**: `type`（account | transfer | all）, `data`（base64 JSON，含 jiid, page 等）
- **响应**: `{ "uid", "data": { "ret", "msg": { "data", "page", "pagemax", "count" } } }`

---

## 10. 图表

- **URL**: `/api/chart` 或 `/Home/Api/chart`
- **参数**: `type`（year | month）, `date`（时间戳）
- **响应**: JSON 数组或对象（年度/月度统计）

---

## 11. 自动复制（管理员）

- **URL**: `/api/autocopy` 或 `/Home/Api/autocopy`
- **参数**: `type`（get | updata）, `data`, `enable`, `enablePullDown`
- **响应**: `{ "strData", "enable", "enablePullDown" }` 或纯文本“非法操作”

---

## 12. 管理员导入（MySQL 导出导入）

- **URL**: `/api/admin/import` 或 `/Home/Api/admin/import`
- **方法**: POST
- **Content-Type**: multipart/form-data，字段 `file` 为 .sql 文件
- **权限**: 仅管理员（session uid == 配置 admin_uid）
- **响应**: `{ "ok": true, "msg": "导入成功", "statements": n }` 或错误信息

---

## 微信小程序说明

- 若小程序原请求为 `https://domain/Home/Api/login`，可继续使用该路径或改为 `/api/login`。
- 请求需携带 Cookie（同域）以维持 Session；若为跨域，需服务端配置 CORS 并确保 Cookie 策略允许。
- 请求体/查询参数与上述一致；`data` 为 base64(JSON.stringify(obj))。
