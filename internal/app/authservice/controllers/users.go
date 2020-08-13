package controllers

import (
	"time"
	// "encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lianmi/servers/internal/app/authservice/services"
	"github.com/lianmi/servers/internal/common/codes"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
	// jwt "github.com/appleboy/gin-jwt/v2"
)

type UsersController struct {
	logger  *zap.Logger
	service services.UsersService
}

func NewUsersController(logger *zap.Logger, s services.UsersService) *UsersController {
	return &UsersController{
		logger:  logger,
		service: s,
	}
}

func (pc *UsersController) GetUser(c *gin.Context) {
	ID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	p, err := pc.service.GetUser(ID)
	if err != nil {
		pc.logger.Error("get product by id error", zap.Error(err))
		c.String(http.StatusInternalServerError, "%+v", err)
		return
	}

	c.JSON(http.StatusOK, p)
}

func (pc *UsersController) Register(c *gin.Context) {
	var user models.User
	code := codes.InvalidParams

	// binding JSON,本质是将request中的Body中的数据按照JSON格式解析到user变量中
	if c.BindJSON(&user) != nil {
		pc.logger.Error("binding JSON error ")

	} else {

		// roles := jwt.ExtractClaims(c)
		// createdBy := roles["userName"].(string)
		user.CreatedAt = time.Now() //注意，必须要显式

		user.State = 1
		user.Avatar = "https://zbj-bucket1.oss-cn-shenzhen.aliyuncs.com/avatar.JPG"
		if !pc.service.ExistUserByName(user.Username) {
			if err := pc.service.Register(&user); err == nil {
				code = codes.SUCCESS
			} else {
				pc.logger.Error("Register user error", zap.Error(err))
				code = codes.ERROR
			}
		} else {
			code = codes.ErrExistUser
		}

		// resp := models.Response{Code: 1, Message: "Ok", Data: ""}
		// pc.logger.Debug("Response Data", zap.String("Json", resp.ToJson()))
		// c.JSON(http.StatusOK, &resp)
	}
	RespOk(c, http.StatusOK, code)
}

func (pc *UsersController) GenerateSmsCode(c *gin.Context) {
	code := codes.SUCCESS
	RespOk(c, http.StatusOK, code)
}

func (pc *UsersController) ChanPassword(c *gin.Context) {
	return
}

func (pc *UsersController) GetUserRoles(username string) []*models.Role {

	return pc.service.GetUserRoles(username)
}

func (pc *UsersController) CheckUser(username, password string) bool {
	return pc.service.CheckUser(username, password)
}

func (pc *UsersController) SaveUserToken(username string, token string, expire time.Time) bool {
	return pc.service.SaveUserToken(username, token, expire)
}
