package gts

import (
	"crypto/rand"
	//"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Store struct {
	Sessions  Session
	SessionID string
	Response  http.ResponseWriter
}

func (c *Store) Get(key string) interface{} {

	session := c.Sessions

	if session == nil {
		return nil
	}
	return session.GetVal(c.SessionID, key)

}
func (c *Store) Set(key string, value interface{}) bool {

	session := c.Sessions
	if session == nil {
		return false
	}
	if c.SessionID == "" {

		c.SessionID = session.New(c.Response)
	}
	session.SetVal(c.SessionID, key, value)
	return true

}
func (c *Store) Del() {

	c.Sessions.Remove(c.SessionID)
}

type Session interface {
	New(w http.ResponseWriter) string
	Del(w http.ResponseWriter, r *http.Request)
	Remove(sessionID string)
	SetVal(sessionID string, key string, value interface{}) error
	GetVal(sessionID string, key string) interface{}
	SessionID(r *http.Request) (string, bool)
	Set(r *http.Request, key string, value interface{}) bool
	Get(r *http.Request, key string) (interface{}, bool)
}

/*Session会话管理*/
type CookieSession struct {
	mCookieName  string       //客户端cookie名称
	mLock        sync.RWMutex //互斥(保证线程安全)
	mMaxLifeTime int64        //垃圾回收时间
	mCookieTime  int64
	mSessions    map[string]*Provider //保存session的指针[sessionID] = session
	mSecure      bool
}

//创建会话管理器(cookieName:在浏览器中cookie的名字;maxLifeTime:最长生命周期)
func NewCookieSession(cookieName string, maxLifeTime, cookieTime int64, secure bool) *CookieSession {

	ses := &CookieSession{mCookieName: cookieName, mMaxLifeTime: maxLifeTime, mCookieTime: cookieTime, mSecure: secure, mSessions: make(map[string]*Provider)}
	//启动定时回收
	go ses.GC()

	return ses
}

//在开始页面登陆页面，开始Session
func (ses *CookieSession) New(w http.ResponseWriter) string {

	ses.mLock.Lock()
	defer ses.mLock.Unlock()

	//无论原来有没有，都重新创建一个新的session
	newSessionID := url.QueryEscape(newSessionID())

	//存指针
	ses.mSessions[newSessionID] = &Provider{mSessionID: newSessionID, mLastTimeAccessed: time.Now(), mValues: make(map[interface{}]interface{})}
	//让浏览器cookie设置过期时间
	cookie := http.Cookie{Name: ses.mCookieName, Value: newSessionID, Path: "/", HttpOnly: true, Secure: ses.mSecure, MaxAge: int(ses.mCookieTime)}
	http.SetCookie(w, &cookie)

	return newSessionID
}

//结束Session
func (ses *CookieSession) Del(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(ses.mCookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {
		ses.mLock.Lock()
		defer ses.mLock.Unlock()

		delete(ses.mSessions, cookie.Value)

		//让浏览器cookie立刻过期
		expiration := time.Now()
		cookie := http.Cookie{Name: ses.mCookieName, Path: "/", HttpOnly: true, Secure: ses.mSecure, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

//结束session
func (ses *CookieSession) Remove(sessionID string) {
	ses.mLock.Lock()
	defer ses.mLock.Unlock()

	delete(ses.mSessions, sessionID)
}

//设置session里面的值
func (ses *CookieSession) SetVal(sessionID string, key string, value interface{}) error {
	ses.mLock.Lock()
	defer ses.mLock.Unlock()

	if session, ok := ses.mSessions[sessionID]; ok {
		session.mValues[key] = value
	}
	return nil
}

//得到session里面的值
func (ses *CookieSession) GetVal(sessionID string, key string) interface{} {
	ses.mLock.RLock()
	defer ses.mLock.RUnlock()

	if session, ok := ses.mSessions[sessionID]; ok {
		if val, ok := session.mValues[key]; ok {

			return val
		}
	}

	return nil
}

//得到sessionID列表
func (ses *CookieSession) GetSessionIDList() []string {
	ses.mLock.RLock()
	defer ses.mLock.RUnlock()

	sessionIDList := make([]string, 0)

	for k, _ := range ses.mSessions {
		sessionIDList = append(sessionIDList, k)
	}

	return sessionIDList[0:len(sessionIDList)]
}

func (ses *CookieSession) SessionID(r *http.Request) (string, bool) {
	var cookie, err = r.Cookie(ses.mCookieName)
	if cookie == nil ||
		err != nil {
		return "", false
	}

	ses.mLock.Lock()
	defer ses.mLock.Unlock()

	sessionID := cookie.Value

	if session, ok := ses.mSessions[sessionID]; ok {
		session.mLastTimeAccessed = time.Now() //判断合法性的同时，更新最后的访问时间

		return sessionID, ok

	}

	return "", false
}

func (ses *CookieSession) Set(r *http.Request, key string, value interface{}) bool {

	var cookie, err = r.Cookie(ses.mCookieName)
	if cookie == nil ||
		err != nil {
		return false
	}

	ses.mLock.Lock()
	defer ses.mLock.Unlock()

	sessionID := cookie.Value

	if session, ok := ses.mSessions[sessionID]; ok {
		session.mLastTimeAccessed = time.Now() //判断合法性的同时，更新最后的访问时间
		session.mValues[key] = value

		return ok
	}

	return false

}
func (ses *CookieSession) Get(r *http.Request, key string) (interface{}, bool) {

	var cookie, err = r.Cookie(ses.mCookieName)
	if cookie == nil ||
		err != nil {
		return nil, false
	}

	ses.mLock.Lock()
	defer ses.mLock.Unlock()

	sessionID := cookie.Value

	if session, ok := ses.mSessions[sessionID]; ok {
		session.mLastTimeAccessed = time.Now() //判断合法性的同时，更新最后的访问时间

		if val, ok := session.mValues[key]; ok {
			return val, ok
		}
	} else {
		return nil, false
	}

	return nil, false

}

//更新最后访问时间
func (ses *CookieSession) GetLastAccessTime(sessionID string) time.Time {
	ses.mLock.RLock()
	defer ses.mLock.RUnlock()

	if session, ok := ses.mSessions[sessionID]; ok {
		return session.mLastTimeAccessed
	}

	return time.Now()
}

//GC回收
func (ses *CookieSession) GC() {
	ses.mLock.Lock()
	defer ses.mLock.Unlock()

	for sessionID, session := range ses.mSessions {
		//删除超过时限的session
		if session.mLastTimeAccessed.Unix()+ses.mMaxLifeTime < time.Now().Unix() {
			delete(ses.mSessions, sessionID)
		}
	}

	//定时回收
	time.AfterFunc(time.Duration(ses.mMaxLifeTime)*time.Second, func() { ses.GC() })
}

//创建唯一ID
func newSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		nano := time.Now().UnixNano() //微秒
		return strconv.FormatInt(nano, 10)
	}

	return Encode(b, BitcoinAlphabet)
	//return base64.URLEncoding.EncodeToString(b)
}

//——————————————————————————
/*会话*/
type Provider struct {
	mSessionID        string                      //唯一id
	mLastTimeAccessed time.Time                   //最后访问时间
	mValues           map[interface{}]interface{} //其它对应值(保存用户所对应的一些值，比如用户权限之类)
}
