package repositories

import (
	"github.com/lianmi/servers/internal/pkg/models"
)

// 查询出大于 systemMsgAt 的系统公告
func (s *MysqlLianmiRepository) GetSystemMsgs(systemMsgAt uint64) (systemMsgs []*models.SystemMsg, err error) {
	wheres := make([]interface{}, 0)
	wheres = append(wheres, []interface{}{"created_at", ">=", systemMsgAt})

	pageIndex := int(0)
	pageSize := int(99) //只搜前99条

	columns := []string{"*"}
	orderBy := "created_at desc"

	db2 := s.db
	db2, err = s.base.BuildQueryList(db2, wheres, columns, orderBy, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	err = db2.Find(&systemMsgs).Error

	return systemMsgs, nil

}
