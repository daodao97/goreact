package invite

import (
	"math/rand"
	"net/http"

	"github.com/daodao97/goreact/base/login"
	"github.com/daodao97/goreact/dao"
	"github.com/daodao97/xgo/xdb"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type InviteCode struct {
	Code string `json:"code" binding:"required"`
}

func GetUserInviteCode(c *gin.Context) {
	user, err := login.GetUserInfo(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user info"})
		return
	}
	inviteCode := getUserInviteCode(user.GetInt64("id"))

	if inviteCode == "" {
		inviteCode = generateInviteCode(user.GetInt64("id"))
		dao.UserModel.Update(
			xdb.Record{"invite_code": inviteCode},
			xdb.WhereEq("id", user.GetInt64("id")),
		)
	}

	c.JSON(http.StatusOK, gin.H{"data": inviteCode})
}

func SetUserInviteCode(c *gin.Context) {
	user, err := login.GetUserInfo(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user info"})
		return
	}

	var userSetCode InviteCode
	if err := c.ShouldBindJSON(&userSetCode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	exist, _ := login.UserModel.Count(xdb.WhereEq("invite_code", userSetCode.Code))
	if exist > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invite code already used"})
		return
	}

	login.UserModel.Update(
		xdb.Record{"invite_code": userSetCode.Code},
		xdb.WhereEq("id", user.GetInt64("id")),
	)
	c.JSON(http.StatusOK, gin.H{"data": userSetCode.Code})
}

func InvitedList(c *gin.Context) {
	user, err := login.GetUserInfo(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user info"})
		return
	}

	page := cast.ToInt(c.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(c.DefaultQuery("page_size", "10"))
	if pageSize > 100 {
		pageSize = 100
	}

	total, invitedList, _ := login.UserModel.Page(page, pageSize, xdb.WhereEq("ref_uid", user.GetInt64("id")))

	var list []xdb.Record

	for _, v := range invitedList {
		name := v.GetString("user_name")
		if name == "" {
			name = v.GetString("email")
		}
		// 对 name 三分之二的字符进行打码
		runeName := []rune(name)
		nameLen := len(runeName)
		if nameLen > 0 {
			maskLen := nameLen * 2 / 3
			if maskLen > 0 {
				start := (nameLen - maskLen) / 2
				end := start + maskLen
				for i := start; i < end; i++ {
					runeName[i] = '*'
				}
			}
			name = string(runeName)
		}
		list = append(list, xdb.Record{
			"username":   name,
			"created_at": v.GetTime("created_at").Format("2006-01-02 15:04:05"),
			"avatar_url": v.GetString("avatar_url"),
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total, "page": page, "page_size": pageSize})
}

func getUserInviteCode(uid int64) string {
	inviteCode, _ := login.UserModel.First(xdb.WhereEq("id", uid))
	return inviteCode.GetString("invite_code")
}

func generateInviteCode(uid int64) string {
	// 生成一个10位以内的随机邀请码
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLen := 8 + uid%3 // 8-10位
	b := make([]byte, codeLen)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
