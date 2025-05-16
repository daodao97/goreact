package base

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Privacy(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html:Privacy.js", map[string]string{"name": "privacy"})
}

func TermsOfService(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html:Privacy.js", map[string]string{"name": "terms"})
}
