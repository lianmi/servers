/*
此文件处理GitHub，QQ/微信等第三方授权的回调


*/
package controllers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//see: https://studygolang.com/articles/20408
//https://docs.github.com/en/free-pro-team@latest/developers/apps/creating-an-oauth-app
func (pc *LianmiApisController) GitHubOAuth(c *gin.Context) {
	code := c.Query("code")
	errMsg := c.Query("error")
	errDescription := c.Query("error_description")
	pc.logger.Debug("GitHubOAuth start ...",
		zap.String("code", code),
		zap.String("error", errMsg),
		zap.String("description", errDescription),
	)

	//TODO 根据获得的code，以及注册号的client_id, client_secret，用POST方法提交到GitHub获取令牌
	//https://api.lianmi.cloud/login-github?error=redirect_uri_mismatch&error_description=The+redirect_uri+MUST+match+the+registered+callback+URL+for+this+application.&error_uri=https%3A%2F%2Fdocs.github.com%2Fapps%2Fmanaging-oauth-apps%2Ftroubleshooting-authorization-request-errors%2F%23redirect-uri-mismatch&state=app

}
