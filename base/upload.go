package base

import (
	"net/http"

	"github.com/daodao97/xgo/xrequest"
	"github.com/gin-gonic/gin"
)

type Uploader struct {
	Token string
	URL   string
}

var (
	uploader = &Uploader{}
)

func SetUploadToken(_uploader *Uploader) {
	uploader = _uploader
}

func GenUploadToken(c *gin.Context) {
	if uploader.URL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Upload URL not set",
		})
		return
	}

	resp, err := xrequest.New().
		SetHeader("X-API-Key", uploader.Token).
		Post(uploader.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if resp.Error() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": resp.Error().Error(),
		})
		return
	}

	body := resp.Json()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   body.Get("token").String(),
	})
}
