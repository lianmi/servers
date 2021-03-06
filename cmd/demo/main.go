package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	zaprotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var level zapcore.Level // 初始化配置文件的Level, 必须高于或等于此级别才显示或写入日志 文件

// const LogZap = "silent"

// const LogZap = "zap"
const LogZap = "error"

const Director = "./logs"
const Prefix = "demo"
const StacktraceKey = "stacktrace"
const LinkName = "latest_log"
const Zap_Level = "debug"
const LogInConsole = true     //同时显示在屏幕上
const OnlyLogInConsole = true //只显示在屏幕上

var (
	//使用 lianmicloud 数据库
	// dsn = "lianmidba:12345678@tcp(127.0.0.1:3306)/lianmicloud?charset=utf8&parseTime=True&loc=Local"
	dsn     = "root:password@tcp(127.0.0.1:3306)/lianmicloud?charset=utf8&parseTime=True&loc=Local"
	db      *gorm.DB
	GVA_LOG *zap.Logger
)

// writer log writer interface
type writer interface {
	Printf(string, ...interface{})
}

type config struct {
	SlowThreshold time.Duration
	Colorful      bool
	LogLevel      logger.LogLevel
}

var (
	Discard = New(log.New(ioutil.Discard, "", log.LstdFlags), config{})
	Default = New(log.New(os.Stdout, "\r\n", log.LstdFlags), config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Warn,
		Colorful:      true,
	})
	Recorder = traceRecorder{Interface: Default, BeginAt: time.Now()}
)

func New(writer writer, config config) logger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = logger.Green + "%s\n" + logger.Reset + logger.Green + "[info] " + logger.Reset
		warnStr = logger.BlueBold + "%s\n" + logger.Reset + logger.Magenta + "[warn] " + logger.Reset
		errStr = logger.Magenta + "%s\n" + logger.Reset + logger.Red + "[error] " + logger.Reset
		traceStr = logger.Green + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
		traceWarnStr = logger.Green + "%s " + logger.Yellow + "%s\n" + logger.Reset + logger.RedBold + "[%.3fms] " + logger.Yellow + "[rows:%v]" + logger.Magenta + " %s" + logger.Reset
		traceErrStr = logger.RedBold + "%s " + logger.MagentaBold + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
	}

	return &customLogger{
		writer:       writer,
		config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type customLogger struct {
	writer
	config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (c *customLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *c
	newLogger.LogLevel = level
	return &newLogger
}

// Info print info
func (c *customLogger) Info(ctx context.Context, message string, data ...interface{}) {
	if c.LogLevel >= logger.Info {
		c.Printf(c.infoStr+message, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages
func (c *customLogger) Warn(ctx context.Context, message string, data ...interface{}) {
	if c.LogLevel >= logger.Warn {
		c.Printf(c.warnStr+message, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (c *customLogger) Error(ctx context.Context, message string, data ...interface{}) {
	if c.LogLevel >= logger.Error {
		c.Printf(c.errStr+message, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
func (c *customLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if c.LogLevel > 0 {
		elapsed := time.Since(begin)
		switch {
		case err != nil && c.LogLevel >= logger.Error:
			sql, rows := fc()
			if rows == -1 {
				c.Printf(c.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				c.Printf(c.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		case elapsed > c.SlowThreshold && c.SlowThreshold != 0 && c.LogLevel >= logger.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", c.SlowThreshold)
			if rows == -1 {
				c.Printf(c.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				c.Printf(c.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		case c.LogLevel >= logger.Info:
			sql, rows := fc()
			if rows == -1 {
				c.Printf(c.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				c.Printf(c.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		}
	}
}

func (c *customLogger) Printf(message string, data ...interface{}) {
	if LogZap != "" {
		switch len(data) {
		case 0:
			GVA_LOG.Info(message)
		case 1:
			GVA_LOG.Info("gorm", zap.Any("src", data[0]))
		case 2:
			GVA_LOG.Info("gorm", zap.Any("src", data[0]), zap.Any("duration", data[1]))
		case 3:
			GVA_LOG.Info("gorm", zap.Any("src", data[0]), zap.Any("duration", data[1]), zap.Any("rows", data[2]))
		case 4:
			GVA_LOG.Info("gorm", zap.Any("src", data[0]), zap.Any("duration", data[1]), zap.Any("rows", data[2]), zap.Any("sql", data[3]))
		}
		return
	}
	switch len(data) {
	case 0:
		c.writer.Printf(message, "")
	case 1:
		c.writer.Printf(message, data[0])
	case 2:
		c.writer.Printf(message, data[0], data[1])
	case 3:
		c.writer.Printf(message, data[0], data[1], data[2])
	case 4:
		c.writer.Printf(message, data[0], data[1], data[2], data[3])
	case 5:
		c.writer.Printf(message, data[0], data[1], data[2], data[3], data[4])
	}
}

type traceRecorder struct {
	logger.Interface
	BeginAt      time.Time
	SQL          string
	RowsAffected int64
	Err          error
}

func (t traceRecorder) New() *traceRecorder {
	return &traceRecorder{Interface: t.Interface, BeginAt: time.Now()}
}

func (t *traceRecorder) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	t.BeginAt = begin
	t.SQL, t.RowsAffected = fc()
	t.Err = err
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetWriteSyncer() (zapcore.WriteSyncer, error) {
	fileWriter, err := zaprotatelogs.New(
		path.Join(Director, "%Y-%m-%d.log"),
		zaprotatelogs.WithLinkName(LinkName),
		zaprotatelogs.WithMaxAge(7*24*time.Hour),
		zaprotatelogs.WithRotationTime(24*time.Hour),
	)
	if OnlyLogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), err
	}
	if LogInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	}
	return zapcore.AddSync(fileWriter), err
}

func Zap() (logger *zap.Logger) {
	if ok, _ := PathExists(Director); !ok { // 判断是否有Director文件夹
		fmt.Printf("create %v directory\n", Director)
		_ = os.Mkdir(Director, os.ModePerm)
	}

	switch Zap_Level { // 初始化配置文件的Level
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "dpanic":
		level = zap.DPanicLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}

	if level == zap.DebugLevel || level == zap.ErrorLevel {
		logger = zap.New(getEncoderCore(), zap.AddStacktrace(level))
	} else {
		logger = zap.New(getEncoderCore())
	}
	// if global.GVA_CONFIG.Zap.ShowLine {
	logger = logger.WithOptions(zap.AddCaller())
	// }
	return logger
}

// getEncoderConfig 获取zapcore.EncoderConfig
func getEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey: "message",
		LevelKey:   "level",
		TimeKey:    "time",
		NameKey:    "logger",
		CallerKey:  "caller",
		// StacktraceKey:  StacktraceKey,
		LineEnding: zapcore.DefaultLineEnding,
		// EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	config.EncodeLevel = zapcore.LowercaseLevelEncoder
	return config
}

// getEncoder 获取zapcore.Encoder
func getEncoder() zapcore.Encoder {
	// if global.GVA_CONFIG.Zap.Format == "json" {
	// return zapcore.NewJSONEncoder(getEncoderConfig())
	// 	}
	return zapcore.NewConsoleEncoder(getEncoderConfig())
}

// getEncoderCore 获取Encoder的zapcore.Core
func getEncoderCore() (core zapcore.Core) {
	writer, err := GetWriteSyncer() // 使用file-rotatelogs进行日志分割
	if err != nil {
		fmt.Printf("Get Write Syncer Failed err:%v", err.Error())
		return
	}
	return zapcore.NewCore(getEncoder(), writer, level)
}

// 自定义日志输出时间格式
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(Prefix + " 2006/01/02 - 15:04:05.000"))
}

func gormConfig(mod bool) *gorm.Config {
	var config = &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}
	switch LogZap {
	case "silent", "Silent":
		config.Logger = Default.LogMode(logger.Silent)
	case "error", "Error":
		config.Logger = Default.LogMode(logger.Error)
	case "warn", "Warn":
		config.Logger = Default.LogMode(logger.Warn)
	case "info", "Info":
		config.Logger = Default.LogMode(logger.Info)
	case "zap", "Zap":
		config.Logger = Default.LogMode(logger.Info)
	default:
		if mod {
			config.Logger = Default.LogMode(logger.Info)
			break
		}
		config.Logger = Default.LogMode(logger.Silent)
	}
	return config
}

func init() {
	var err error
	GVA_LOG = Zap() // 初始化zap日志库

	db, err = gorm.Open(mysql.Open(dsn), gormConfig(false))
	if err != nil {
		log.Fatalln(err)
	}
	db = db.Debug()

}

func PrintPretty(i interface{}) {
	data, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

func main() {
	page := 0
	pageSize := 20
	// var count int64
	var users []models.User
	userModel := new(models.User)

	db.Model(&userModel).Find(&users, "user_type=?", 0)

	GVA_LOG.Info(" *********  查询 users *******  ", zap.Int("count", len(users)))
	// PrintPretty(users)

	// db.Model(&userModel).Scopes(IsNormalUser, Paginate(page, pageSize)).Find(&users).Order("updated_at DESC")
	//注意！Order必须在Find之前
	db.Model(&userModel).Scopes(IsNormalUser, Paginate(page, pageSize), BetweenCreateAt(0, 1606514952437)).Order("created_at DESC").Find(&users)
	// db = db
	// db.Model(&userModel).Scopes(IsBusinessUser, Paginate(page, pageSize)).Find(&users)
	// db.Model(&userModel).Scopes(IsPreBusinessUser, LegalPerson([]string{"杜老板"}), Paginate(page, pageSize)).Find(&users)

	// log.Println("分页显示users列表, count: ", len(users))
	GVA_LOG.Info("分页显示users列表", zap.Int("count", len(users)))

	for idx, user := range users {
		GVA_LOG.Debug(fmt.Sprintf("idx=%d, create_at %d, username=%s, mobile=%s\n", idx, user.CreatedAt, user.Username, user.Mobile))
	}
	_ = page
	_ = pageSize

}

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

//按createAt的时间段
func BetweenCreateAt(startAt, endAt uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at>= ? and created_at<= ? ", startAt, endAt)
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

// SELECT * FROM `users` WHERE user_type = 1  AND (create_at>= 1605328128169 and create_at<= 1603789653918 ) AND `users`.`deleted_at` IS NULL LIMIT 20
