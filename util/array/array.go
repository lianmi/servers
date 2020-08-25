package array

import (
	"bytes"
	"encoding/json"

	// "crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math"

	// "math/big"
	mrand "math/rand"
	"regexp"
	"strconv"
	"time"

	//高精度科学运算
	"github.com/gomodule/redigo/redis"
	"github.com/shopspring/decimal"
)

const base_format = "2006-01-02 15:04:05"

//截取字符串 start 起点下标 end 终点下标(不包括)
func Substr2(str string, start int, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < 0 || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end])
}

//取小数点后n位
func Round2(f float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n)+"f", f)
	inst, _ := strconv.ParseFloat(floatStr, 64)
	return inst
}

func Int32ToByte(num int32) []byte {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, num); err != nil {
		return nil
	}
	return buffer.Bytes()
}

//字节转换成整形
func BytesToInt32(b []byte) int32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return tmp

}

//创建redis Pool
//db范围是0-15
func NewPool(addr, password string, db int) *redis.Pool {
	// return &redis.Pool{
	// 	MaxIdle:     10,
	// 	IdleTimeout: 240 * time.Second,
	// 	Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	// }
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if db < 0 || db > 15 {
				c.Close()
				return nil, err
			}

			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func IsEmailValid(email string) bool {
	if m, _ := regexp.MatchString(`^([\w\.\_]{2,10})@(\w{1,}).([a-z]{2,4})$`, email); !m {
		// fmt.Fprint(res, "电子邮件格式不正确")
		fmt.Println("m=", m)
		return false
	}
	return true
}

func IsEmail(email string) bool {
	if email != "" {
		if isOk, _ := regexp.MatchString("^[_a-z0-9-]+(\\.[_a-z0-9-]+)*@[a-z0-9-]+(\\.[a-z0-9-]+)*(\\.[a-z]{2,4})$", email); isOk {
			return true
		}
	}

	return false
}

func IsPhone(phoneStr string) bool {
	if phoneStr != "" {
		if isOk, _ := regexp.MatchString(`^\([\d]{3}\) [\d]{3}-[\d]{4}$`, phoneStr); isOk {
			return isOk
		}
	}

	return false
}

//取两个数字之间的随机数
func RandInt(min, max int) int {
	mrand.Seed(time.Now().Unix())
	randNum := mrand.Intn(max - min)
	randNum = randNum + min
	// fmt.Printf("rand is %v\n", randNum)
	return randNum
}

//取两个数字之间的随机float64数
func RandFloat64(min, max float64) float64 {
	mrand.Seed(time.Now().UnixNano()) //利用当前时间的UNIX时间戳初始化rand包
	f64 := mrand.Float64()*(max-min) + min
	if f64 > max {
		return max
	}
	return Round2(f64, 4)
}

func Float32ToByte(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)

	return bytes
}

func ByteToFloat32(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)

	return math.Float32frombits(bits)
}

func Float64ToByte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)

	return bytes
}

func ByteToFloat64(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)

	return math.Float64frombits(bits)
}

func ReverseBytes(l [][]byte) {
	for i := 0; i < int(len(l)/2); i++ {
		li := len(l) - i - 1
		// fmt.Println(i, "<=>", li)
		l[i], l[li] = l[li], l[i]
	}
}

func PrintPretty(i interface{}) {
	data, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

//功能： 两个float64相乘，返回float64
func Float64Mul(price, amount float64) float64 {
	d1 := decimal.New(1, 4)  //精度4
	d2 := decimal.New(1, -8) //精度8
	f11 := decimal.NewFromFloat(price).Mul(d1)
	f12 := decimal.NewFromFloat(amount).Mul(d1)

	value, _ := f11.Mul(f12).Mul(d2).Float64()
	return value
}

//根据时间戳返回当前时刻（分钟不变）的第0秒的时间戳
func ParseHourMinute(ts int64) (int64, error) {
	now := time.Unix(ts, 0)
	curr_y := now.Year()
	curr_m := now.Month()
	curr_d := now.Day()
	curr_hour := now.Hour()
	curr_minute := now.Minute()

	dateStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d:00", curr_y, curr_m, curr_d, curr_hour, curr_minute)
	loc, _ := time.LoadLocation("Local")                            //重要：获取时区
	theTime, err := time.ParseInLocation(base_format, dateStr, loc) //使用模板在对应时区转化为time.time类型
	if err != nil {
		return 0, err
	}
	return theTime.Unix(), nil
}
