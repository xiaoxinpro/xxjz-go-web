package handler

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/service"
	"github.com/xiaoxinpro/xxjz-go-web/backend/internal/session"
)

// RequireSession validates session shell and aborts with 200 + uid:0 if invalid.
// Does not require login for routes that return uid:0 in body when not logged in.
func RequireSession(userSvc *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := sessions.Default(c)
		uid := session.GetUID(sess)
		username := session.GetUsername(sess)
		shell := session.GetShell(sess)
		if uid <= 0 || username == "" || !userSvc.UserShell(username, shell) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("uid", uid)
		c.Set("username", username)
		c.Next()
	}
}
