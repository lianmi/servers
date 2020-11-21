/*
这个文件是和前端相关的restful接口-客服模块，/v1/customerservice/....
*/
package controllers

import (
	"fmt"
	"net/http"

	Global "github.com/lianmi/servers/api/proto/global"
	// Order "github.com/lianmi/servers/api/proto/order"
	Auth "github.com/lianmi/servers/api/proto/auth"
	// User "github.com/lianmi/servers/api/proto/user"

	"github.com/gin-gonic/gin"
)

//获取空闲的在线客服id数组
func (pc *LianmiApisController) QueryCustomerServices(c *gin.Context) {
	var req Auth.QueryCustomerServiceReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		csList, err := pc.service.QueryCustomerServices(&req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Query CustomerServices failed")
		} else {
			resp := &Auth.QueryCustomerServiceResp{}
			for _, onlineCustomerService := range csList {
				resp.OnlineCustomerServices = append(resp.OnlineCustomerServices, &Auth.CustomerServiceInfo{
					Username:   onlineCustomerService.Username,
					JobNumber:  onlineCustomerService.JobNumber,
					Type:       Global.CustomerServiceType(onlineCustomerService.Type),
					Evaluation: onlineCustomerService.Evaluation,
					NickName:   onlineCustomerService.Evaluation,
				})
			}

			RespData(c, http.StatusOK, 200, resp)
		}

	}

}

func (pc *LianmiApisController) QueryGrades(c *gin.Context) {
	var maps string
	var req Auth.GradeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {
		pageIndex := int(req.Page)
		pageSize := int(req.Limit)
		total := new(int64)
		if pageIndex == 0 {
			pageIndex = 1
		}
		if pageSize == 0 {
			pageSize = 100
		}

		// GetPages 分页返回数据
		if req.StartAt > 0 && req.EndAt > 0 {
			maps = fmt.Sprintf("created_at >= %d and created_at <= %d", req.StartAt, req.EndAt)
		}
		pfList, err := pc.service.QueryGrades(&req, pageIndex, pageSize, total, maps)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Query Grades( failed")
		} else {
			pages := Auth.GradesPage{
				TotalPage: uint64(*total),
				// Grades: pfList,
			}
			for _, grade := range pfList {
				pages.Grades = append(pages.Grades, &Auth.GradeInfo{
					Title:                   grade.Title,
					AppUsername:             grade.AppUsername,
					CustomerServiceUsername: grade.CustomerServiceUsername,
					JobNumber:               grade.JobNumber,
					Type:                    Global.CustomerServiceType(grade.Type),
					Evaluation:              grade.Evaluation,
					NickName:                grade.NickName,
					Catalog:                 grade.Catalog,
					Desc:                    grade.Desc,
					GradeNum:                int32(grade.GradeNum),
				})
			}

			RespData(c, http.StatusOK, 200, pages)
		}

	}
}

//客服人员增加求助记录，以便发给用户评分
func (pc *LianmiApisController) AddGrade(c *gin.Context) {
	var req Auth.AddGradeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if req.CustomerServiceUsername == "" {
			RespFail(c, http.StatusBadRequest, 400, "CustomerServiceUsername参数错误")
		}

		title, err := pc.service.AddGrade(&req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Add Grade failed")
		} else {

			RespData(c, http.StatusOK, 200, &Auth.GradeTitleInfo{
				CustomerServiceUsername: req.CustomerServiceUsername,
				Title:                   title,
			})
		}
	}

}

//用户提交评分
func (pc *LianmiApisController) SubmitGrade(c *gin.Context) {
	var req Auth.SubmitGradeReq
	if c.BindJSON(&req) != nil {
		pc.logger.Error("binding JSON error ")
		RespFail(c, http.StatusBadRequest, 400, "参数错误, 缺少必填字段")
	} else {

		if req.AppUsername == "" {
			RespFail(c, http.StatusBadRequest, 400, "AppUsername参数错误")
		}

		err := pc.service.SubmitGrade(&req)

		if err != nil {
			RespFail(c, http.StatusBadRequest, 400, "Submit Grade failed")
		} else {

			RespOk(c, http.StatusOK, 200)
		}
	}

}
