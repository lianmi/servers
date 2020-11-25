package repositories

import (
	// "github.com/lianmi/servers/internal/pkg/models"
	// "github.com/pkg/errors"
	"strings"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//BaseRepository 注入db,Logger
type BaseRepository struct {
	logger *zap.Logger
	db     *gorm.DB
}


func NewBaseRepository(logger *zap.Logger, db *gorm.DB) *BaseRepository {
	return &BaseRepository{
		logger: logger.With(zap.String("type", "BaseRepository")),
		db:     db,
	}
}

//构建查询条件
func (b *BaseRepository)BuildCondition(where map[string]interface{}) (whereSql string,
    values []interface{}, err error) {
    for key, value := range where {
        conditionKey := strings.Split(key, " ")
        if len(conditionKey) > 2 {
            return "", nil, fmt.Errorf("" +
                "map构建的条件格式不对，类似于'age >'")
        }
        if whereSql != "" {
            whereSql += " AND "
        }
        switch len(conditionKey) {
        case 1:
            whereSql += fmt.Sprint(conditionKey[0], " = ?")
            values = append(values, value)
            break
        case 2:
            field := conditionKey[0]
            switch conditionKey[1] {
            case "=":
                whereSql += fmt.Sprint(field, " = ?")
                values = append(values, value)
                break
            case ">":
                whereSql += fmt.Sprint(field, " > ?")
                values = append(values, value)
                break
            case ">=":
                whereSql += fmt.Sprint(field, " >= ?")
                values = append(values, value)
                break
            case "<":
                whereSql += fmt.Sprint(field, " < ?")
                values = append(values, value)
                break
            case "<=":
                whereSql += fmt.Sprint(field, " <= ?")
                values = append(values, value)
                break
            case "in":
                whereSql += fmt.Sprint(field, " in (?)")
                values = append(values, value)
                break
            case "like":
                whereSql += fmt.Sprint(field, " like ?")
                values = append(values, value)
                break
            case "<>":
                whereSql += fmt.Sprint(field, " != ?")
                values = append(values, value)
                break
            case "!=":
                whereSql += fmt.Sprint(field, " != ?")
                values = append(values, value)
                break
            }
            break
        }
    }
    return
}

// Create 创建实体
func (b *BaseRepository) Create(value interface{}) error {
	return b.db.Create(value).Error
}

// Save 保存实体
func (b *BaseRepository) Save(value interface{}) error {
	return b.db.Save(value).Error
}

// // Updates 更新实体
// func (b *BaseRepository) Updates(model interface{}, value interface{}) error {
// 	return b.db.Model(model).Updates(value).Error
// }

// DeleteByWhere 根据条件删除实体
func (b *BaseRepository) DeleteByWhere(model, where interface{}) (count int64, err error) {
	db := b.db.Where(where).Delete(model)
	err = db.Error
	if err != nil {
		b.logger.Error("删除实体出错", zap.Error(err))
		return
	}
	count = db.RowsAffected
	return
}

// DeleteByID 根据id删除实体
func (b *BaseRepository) DeleteByID(model interface{}, id int) error {
	return b.db.Where("id=?", id).Delete(model).Error
}

// DeleteByIDS 根据多个id删除多个实体
func (b *BaseRepository) DeleteByIDS(model interface{}, ids []int) (count int64, err error) {
	db := b.db.Where("id in (?)", ids).Delete(model)
	err = db.Error
	if err != nil {
		b.logger.Error("删除多个实体出错", zap.Error(err))
		return
	}
	count = db.RowsAffected
	return
}

// First 根据条件获取一个实体
func (b *BaseRepository) First(where interface{}, out interface{}, selects ...string) error {
	// db := b.db.Where(condition, values)
	db := b.db.Where(where)
	if len(selects) > 0 {
		for _, sel := range selects {
			db = db.Select(sel)
		}
	}
	return db.First(out).Error
}

// FirstByID 根据条件获取一个实体
func (b *BaseRepository) FirstByID(out interface{}, id int) error {
	return b.db.First(out, id).Error
}

// Find 根据条件返回数据
func (b *BaseRepository) Find(where interface{}, out interface{}, sel string, orders ...string) error {
	db := b.db.Where(where)
	if sel != "" {
		db = db.Select(sel)
	}
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	return db.Find(out).Error
}

// GetPages 分页返回数据
func (b *BaseRepository) GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *int64, where interface{}, orders ...string) error {
	db := b.db.Model(model).Where(model)
	db = db.Where(where)
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	err := db.Count(totalCount).Error
	if err != nil {
		b.logger.Error("查询总数出错", zap.Error(err))
		return err
	}
	if *totalCount == 0 {
		return nil
	}
	return db.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(out).Error
}

// PluckList 查询 model 中的一个列作为切片
func (b *BaseRepository) PluckList(model, where interface{}, out interface{}, fieldName string) error {
	return b.db.Model(model).Where(where).Pluck(fieldName, out).Error
}

//GetTransaction 获取事务
func (b *BaseRepository) GetTransaction() *gorm.DB {
	return b.db.Begin()
}
