package gts

import (
	"crypto/rand"
	"github.com/garyburd/redigo/redis"
	"github.com/vmihailenco/msgpack"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

/*Session会话管理*/
type RedisSession struct {
	mCookieName string //客户端cookie名称
	//mLock        sync.RWMutex //互斥(保证线程安全)
	mMaxLifeTime int64 //垃圾回收时间
	mCookieTime  int64
	//mSessions    map[string]*Provider //保存session的指针[sessionID] = session
}

var rp *RedisPool

type RedisPool struct {
	Pool *redis.Pool
}

func newRedisPool(server, password string) (*RedisPool, error) {

	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}

			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return &RedisPool{pool}, nil
}

func SetEx(key string, value []byte, time int32) error {

	c := rp.Pool.Get()
	defer c.Close()

	_, err := c.Do("SET", key, value, "EX", time)
	return err
}
func DelEx(key string) error {

	c := rp.Pool.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	return err
}
func GetEx(key string) ([]byte, error) {

	c := rp.Pool.Get()
	defer c.Close()
	return redis.Bytes(c.Do("GET", key))
}

//创建会话管理器(cookieName:在浏览器中cookie的名字;maxLifeTime:最长生命周期)
func NewRedisSession(cookieName string, maxLifeTime, cookieTime int64, RedisHost, RedisPwd string) *RedisSession {

	var err error

	rp, err = newRedisPool(RedisHost, RedisPwd)
	if err != nil {
		panic(err)
	}

	ses := &RedisSession{mCookieName: cookieName, mMaxLifeTime: maxLifeTime, mCookieTime: cookieTime}

	return ses
}

//在开始页面登陆页面，开始Session
func (ses *RedisSession) New(w http.ResponseWriter) string {

	//无论原来有没有，都重新创建一个新的session
	newSessionID := url.QueryEscape(ses.NewSessionID())

	mValues := map[string]interface{}{newSessionID: newSessionID}

	b, err := msgpack.Marshal(mValues)
	if err == nil {

		SetEx(newSessionID, b, int32(ses.mMaxLifeTime))
	}

	//让浏览器cookie设置过期时间
	cookie := http.Cookie{Name: ses.mCookieName, Value: newSessionID, Path: "/", HttpOnly: true, MaxAge: int(ses.mCookieTime)}
	http.SetCookie(w, &cookie)

	return newSessionID
}

//结束Session
func (ses *RedisSession) Del(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(ses.mCookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {

		DelEx(cookie.Value)

		//让浏览器cookie立刻过期
		expiration := time.Now()
		cookie := http.Cookie{Name: ses.mCookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

//结束session
func (ses *RedisSession) Remove(sessionID string) {

	DelEx(sessionID)
}

//设置session里面的值
func (ses *RedisSession) SetVal(sessionID string, key string, value interface{}) error {

	b, err := GetEx(sessionID)
	if err != nil {

		return err
	}

	var out map[string]interface{}

	if err := msgpack.Unmarshal(b, &out); err != nil {

		return err
	}

	out[key] = value

	m, err := msgpack.Marshal(out)
	if err != nil {
		return err
	}

	return SetEx(sessionID, m, int32(ses.mMaxLifeTime))

}

//得到session里面的值
func (ses *RedisSession) GetVal(sessionID string, key string) interface{} {

	b, err := GetEx(sessionID)
	if err != nil {

		return nil
	}

	var out map[string]interface{}
	if err := msgpack.Unmarshal(b, &out); err != nil {

		return nil
	}

	SetEx(sessionID, b, int32(ses.mMaxLifeTime))

	return out[key]
}

//得到sessionID列表
func (ses *RedisSession) GetSessionIDList() []string {

	return nil
}

func (ses *RedisSession) SessionID(r *http.Request) (string, bool) {
	var cookie, err = r.Cookie(ses.mCookieName)
	if cookie == nil ||
		err != nil {
		return "", false
	}

	sessionID := cookie.Value

	b, err := GetEx(sessionID)
	if err != nil {

		return "", false
	}

	SetEx(sessionID, b, int32(ses.mMaxLifeTime))

	return sessionID, true
}

func (ses *RedisSession) Set(r *http.Request, key string, value interface{}) bool {

	var cookie, err = r.Cookie(ses.mCookieName)
	if cookie == nil ||
		err != nil {
		return false
	}

	sessionID := cookie.Value

	b, err := GetEx(sessionID)
	if err != nil {

		return false
	}

	var out map[string]interface{}

	if err := msgpack.Unmarshal(b, &out); err != nil {

		return false
	}

	out[key] = value

	m, err := msgpack.Marshal(out)
	if err != nil {
		return false
	}

	err = SetEx(sessionID, m, int32(ses.mMaxLifeTime))
	if err != nil {

		return false
	}
	return true

}
func (ses *RedisSession) Get(r *http.Request, key string) (interface{}, bool) {

	var cookie, err = r.Cookie(ses.mCookieName)
	if cookie == nil ||
		err != nil {
		return nil, false
	}

	sessionID := cookie.Value

	b, err := GetEx(sessionID)
	if err != nil {

		return nil, false
	}

	var out map[string]interface{}
	if err := msgpack.Unmarshal(b, &out); err != nil {

		return nil, false
	}

	SetEx(sessionID, b, int32(ses.mMaxLifeTime))

	if out[key] == nil {

		return nil, false
	}
	return out[key], true

}

//更新最后访问时间
func (ses *RedisSession) GetLastAccessTime(sessionID string) time.Time {

	b, err := GetEx(sessionID)
	if err != nil {

		return time.Now()
	}
	SetEx(sessionID, b, int32(ses.mMaxLifeTime))

	return time.Now()
}

//创建唯一ID
func (ses *RedisSession) NewSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		nano := time.Now().UnixNano() //微秒
		return strconv.FormatInt(nano, 10)
	}

	return Encode(b, BitcoinAlphabet)
	//return base64.URLEncoding.EncodeToString(b)
}