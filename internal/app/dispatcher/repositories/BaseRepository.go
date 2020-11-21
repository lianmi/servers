package repositories

import (
	// "github.com/lianmi/servers/internal/pkg/models"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"reflect"
	"strings"
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
func (b *BaseRepository) BuildCondition(where map[string]interface{}) (whereSql string,
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

// First2 根据条件获取一个实体
// func (b *BaseRepository) First2(where interface{}, out interface{}, selects ...string) error {
// 	// db := b.db.Where(condition, values)
// 	db := b.db.Where(&models.User{Username: "lsj001", Password: "654321"})
// 	// if len(selects) > 0 {
// 	// 	for _, sel := range selects {
// 	// 		db = db.Select(sel)
// 	// 	}
// 	// }
// 	return db.First(out).Error
// }

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

// BuildWhere 构建where条件
func (b *BaseRepository) BuildWhere(db *gorm.DB, where interface{}) (*gorm.DB, error) {
	var err error
	t := reflect.TypeOf(where).Kind()
	if t == reflect.Struct || t == reflect.Map {
		db = db.Where(where)
	} else if t == reflect.Slice {
		for _, item := range where.([]interface{}) {
			item := item.([]interface{})
			column := item[0]
			if reflect.TypeOf(column).Kind() == reflect.String {
				count := len(item)
				if count == 1 {
					return nil, errors.New("切片长度不能小于2")
				}
				columnstr := column.(string)
				// 拼接参数形式
				if strings.Index(columnstr, "?") > -1 {
					db = db.Where(column, item[1:]...)
				} else {
					cond := "and" //cond
					opt := "="
					_opt := " = "
					var val interface{}
					if count == 2 {
						opt = "="
						val = item[1]
					} else {
						opt = strings.ToLower(item[1].(string))
						_opt = " " + strings.ReplaceAll(opt, " ", "") + " "
						val = item[2]
					}

					if count == 4 {
						cond = strings.ToLower(strings.ReplaceAll(item[3].(string), " ", ""))
					}

					/*
					   '=', '<', '>', '<=', '>=', '<>', '!=', '<=>',
					   'like', 'like binary', 'not like', 'ilike',
					   '&', '|', '^', '<<', '>>',
					   'rlike', 'regexp', 'not regexp',
					   '~', '~*', '!~', '!~*', 'similar to',
					   'not similar to', 'not ilike', '~~*', '!~~*',
					*/

					if strings.Index(" in notin ", _opt) > -1 {
						// val 是数组类型
						column = columnstr + " " + opt + " (?)"
					} else if strings.Index(" = < > <= >= <> != <=> like likebinary notlike ilike rlike regexp notregexp", _opt) > -1 {
						column = columnstr + " " + opt + " ?"
					}

					if cond == "and" {
						db = db.Where(column, val)
					} else {
						db = db.Or(column, val)
					}
				}
			} else if t == reflect.Map /*Map*/ {
				db = db.Where(item)
			} else {
				/*
					// 解决and 与 or 混合查询，但这种写法有问题，会抛出 invalid query condition
					db = db.Where(func(db *gorm.DB) *gorm.DB {
						db, err = BuildWhere(db, item)
						if err != nil {
							panic(err)
						}
						return db
					})*/

				db, err = b.BuildWhere(db, item)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		return nil, errors.New("参数有误")
	}
	return db, nil
}

func (b *BaseRepository) BuildQueryList(db *gorm.DB, wheres interface{}, columns interface{}, orderBy interface{}, page, rows int) (*gorm.DB, error) {
	var err error
	db, err = b.BuildWhere(db, wheres)
	if err != nil {
		return nil, err
	}
	db = db.Select(columns)
	if orderBy != nil && orderBy != "" {
		db = db.Order(orderBy)
	}
	if page > 0 && rows > 0 {
		db = db.Limit(rows).Offset((page - 1) * rows)
	}
	return db, err
}
