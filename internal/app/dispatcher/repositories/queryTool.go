package repositories

import (
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//用户类型 1-普通
func IsNormalUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ? ", 1)
}

//用户类型 2-商户  处于预审核状态
func IsPreBusinessUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ?  and  state = ?", 2, 0)
}

//用户类型 2-商户  处于已审核状态
func IsBusinessUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ?  and  state = ?", 2, 1)
}

//按createAt的时间段
func BetweenCreatedAt(startAt, endAt uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at>= ? and created_at<= ? ", startAt, endAt)
	}
}

//按createAt的时间段
func BetweenUpdatedAt(startAt, endAt uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("updated_at>= ? and updated_at<= ? ", startAt, endAt)
	}
}

//实名
func TrueName(trueNames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("true_name IN (?)", trueNames)
	}
}

//法人
func LegalPerson(legalPersons []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("legal_person IN (?)", legalPersons)
	}
}

//店铺名称
func Branchesname(branchesnames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("branchesname IN (?)", branchesnames)
	}
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 20
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

//按createAt的时间段
func BetweenCreateAt(startAt, endAt uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at>= ? and created_at<= ? ", startAt, endAt)
	}
}

//分页显示users列表
func (s *MysqlLianmiRepository) usersPageDemo(page, pageSize int) {
	// page := 1
	// pageSize := 20
	// var count int64
	var users []models.User
	userModel := new(models.User)

	s.db.Model(&userModel).Find(&users, "user_type=?", 0)

	// db.Model(&userModel).Scopes(IsNormalUser, Paginate(page, pageSize)).Find(&users)
	// db.Model(&userModel).Scopes(IsBusinessUser, Paginate(page, pageSize)).Find(&users)
	//注意！Order必须在Find之前
	s.db.Model(&userModel).Scopes(IsNormalUser, Paginate(page, pageSize), BetweenCreateAt(1606408683991, 1606514952437)).Order("created_at DESC").Find(&users)

	// s.db.Model(&userModel).Scopes(IsPreBusinessUser, LegalPerson([]string{"杜老板"}), Paginate(page, pageSize)).Find(&users)

	// count =
	s.logger.Debug("分页显示users列表, count: ", zap.Int("len", len(users)))

	for _, user := range users {
		// log.Printf("idx=%d, username=%s, mobile=%d\n", idx, user.Username, user.Mobile)
		s.logger.Debug("分页显示users列表 ",
			zap.String("username", user.Username),
			zap.String("Mobile", user.Mobile),
		)
	}

}

