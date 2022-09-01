package gts

import (
	"github.com/gomodule/redigo/redis"
	"github.com/vmihailenco/msgpack"

	"net/http"
	"net/url"

	"time"
)

/*Session会话管理*/
type RedisSession struct {
	mCookieName string //客户端cookie名称
	//mLock        sync.RWMutex //互斥(保证线程安全)
	mMaxLifeTime int64 //垃圾回收时间
	mCookieTime  int64
	mSecure      bool
	//mSessions    map[string]*Provider //保存session的指针[sessionID] = session
}

var rp *RedisPool

type RedisPool struct {
	Pool *redis.Pool
}

func newRedisPool(server, password string, database ...int) (*RedisPool, error) {

	db := 0
	if database != nil {

		db = database[0]
	}

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
			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
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
func NewRedisSession(cookieName string, maxLifeTime, cookieTime int64,secure bool, RedisHost, RedisPwd string, database ...int) *RedisSession {

	var err error

	rp, err = newRedisPool(RedisHost, RedisPwd, database...)
	if err != nil {
		panic(err)
	}

	ses := &RedisSession{mCookieName: cookieName, mMaxLifeTime: maxLifeTime, mCookieTime: cookieTime, mSecure: secure}

	return ses
}

//在开始页面登陆页面，开始Session
func (ses *RedisSession) New(w http.ResponseWriter) string {

	//无论原来有没有，都重新创建一个新的session
	newSessionID := url.QueryEscape(newSessionID())

	mValues := map[string]interface{}{newSessionID: newSessionID}

	b, err := msgpack.Marshal(mValues)
	if err == nil {

		SetEx(ses.mCookieName+newSessionID, b, int32(ses.mMaxLifeTime))
	}

	//让浏览器cookie设置过期时间
	cookie := http.Cookie{Name: ses.mCookieName, Value: newSessionID, Path: "/", HttpOnly: true,Secure: ses.mSecure, MaxAge: int(ses.mCookieTime)}
	http.SetCookie(w, &cookie)

	return newSessionID
}

//结束Session
func (ses *RedisSession) Del(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(ses.mCookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {

		DelEx(ses.mCookieName + cookie.Value)

		//让浏览器cookie立刻过期
		expiration := time.Now()
		cookie := http.Cookie{Name: ses.mCookieName, Path: "/", HttpOnly: true,Secure: ses.mSecure, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

//结束session
func (ses *RedisSession) Remove(sessionID string) {

	DelEx(ses.mCookieName + sessionID)
}

//设置session里面的值
func (ses *RedisSession) SetVal(sessionID string, key string, value interface{}) error {

	b, err := GetEx(ses.mCookieName + sessionID)
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

	return SetEx(ses.mCookieName+sessionID, m, int32(ses.mMaxLifeTime))

}

//得到session里面的值
func (ses *RedisSession) GetVal(sessionID string, key string) interface{} {

	b, err := GetEx(ses.mCookieName + sessionID)
	if err != nil {

		return nil
	}

	var out map[string]interface{}
	if err := msgpack.Unmarshal(b, &out); err != nil {

		return nil
	}

	SetEx(ses.mCookieName+sessionID, b, int32(ses.mMaxLifeTime))

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

	b, err := GetEx(ses.mCookieName + sessionID)
	if err != nil {

		return "", false
	}

	SetEx(ses.mCookieName+sessionID, b, int32(ses.mMaxLifeTime))

	return sessionID, true
}

func (ses *RedisSession) Set(r *http.Request, key string, value interface{}) bool {

	var cookie, err = r.Cookie(ses.mCookieName)
	if cookie == nil ||
		err != nil {
		return false
	}

	sessionID := cookie.Value

	b, err := GetEx(ses.mCookieName + sessionID)
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

	err = SetEx(ses.mCookieName+sessionID, m, int32(ses.mMaxLifeTime))
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

	b, err := GetEx(ses.mCookieName + sessionID)
	if err != nil {

		return nil, false
	}

	var out map[string]interface{}
	if err := msgpack.Unmarshal(b, &out); err != nil {

		return nil, false
	}

	SetEx(ses.mCookieName+sessionID, b, int32(ses.mMaxLifeTime))

	if out[key] == nil {

		return nil, false
	}
	return out[key], true

}
