/*
此文件处理GitHub，QQ/微信等第三方授权的回调


*/
package controllers

import (
	// "net/http"
	// "strconv"
	// "time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//see: https://studygolang.com/articles/20408
//https://docs.github.com/en/free-pro-team@latest/developers/apps/creating-an-oauth-app
func (pc *LianmiApisController) GitHubOAuth(c *gin.Context) {
	code := c.Query("code")
	pc.logger.Debug("GitHubOAuth start ...", zap.String("code", code))

	//TODO 根据获得的code，以及注册号的client_id, client_secret，用POST方法提交到GitHub获取令牌

}
