package datas

import (
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gookit/color"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/model"
	"gorm.io/gorm"
	"os"
)

func InitMysqlData(db *gorm.DB) {
	InitSysApi(db)
	InitSysUser(db)
	InitExaCustomer(db)
	InitCasbinModel(db)
	InitSysAuthority(db)
	InitSysBaseMenus(db)
	InitAuthorityMenu(db)
	InitSysDictionary(db)
	InitSysAuthorityMenus(db)
	InitSysDataAuthorityId(db)
	InitSysDictionaryDetail(db)
	InitExaFileUploadAndDownload(db)
	InitWkProcess(db)
}

func InitMysqlTables(db *gorm.DB) {
	var err error
	if !db.Migrator().HasTable("casbin_rule") {
		err = db.Migrator().CreateTable(&gormadapter.CasbinRule{})
		color.Info.Println("[Mysql]-->表casbin_rule创建成功")
	} else {
		color.Info.Println("[Mysql]-->表casbin_rule 已存在，不需要创建 ")
	}
	err = db.AutoMigrate(
		model.SysApi{},
		model.SysUser{},
		model.ExaFile{},
		model.ExaCustomer{},
		model.SysBaseMenu{},
		model.SysAuthority{},
		model.JwtBlacklist{},
		model.ExaFileChunk{},
		model.SysDictionary{},
		model.ExaSimpleUploader{},
		model.SysOperationRecord{},
		model.SysDictionaryDetail{},
		model.SysBaseMenuParameter{},
		model.ExaFileUploadAndDownload{},
		model.WorkflowProcess{},
		model.WorkflowNode{},
		model.WorkflowEdge{},
		model.WorkflowStartPoint{},
		model.WorkflowEndPoint{},
	)
	if err != nil {
		color.Warn.Printf("[Mysql]-->初始化数据表失败,err: %v\n", err)
		os.Exit(0)
	}
	color.Info.Println("[Mysql]-->初始化数据表成功")
}
