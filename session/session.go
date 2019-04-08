package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

//session接口
type Sessioner interface {
	Get(key string) interface{}
	Set(key string,val interface{})
	Del(key string)
	ID() string
	Reset() //删除session中所有元素
}

//实现session存储的接口
type Provider interface {
	Create(sid string) (Sessioner,error)
	Read(sid string) (Sessioner,error)
	Delete(sid string) error
	Reset()
	GC(maxLifeTime int64)
	Exist(string) bool
}

var providers = make(map[string]Provider)

//驱动应当主动调用Register，把自己注入
func Register(name string,provider Provider)  {
	if provider == nil{
		panic("session: "+name+"是 nil，provider 不能为空")
	}
	if _,ok := providers[name];ok{
		panic("session: "+name+"已经被使用，换个名试试？")
	}
	providers[name] = provider
}

type ManagerConf struct {
	ProviderName string //default:memory
	CookieName string //default:randstring
	EnableSetCookie bool //true，使用cookie，false 优先使用header头部重写，其次使用URL重写，default:true
	EnableSidInHTTPHeader   bool   //如果cookie不可用，则使用header重写，default:true
	SessionNameInHTTPHeader string //default:"wtf"
	EnableSidInURLQuery     bool   //使用url重写，要是再不允许那我也没办法了,只能在form的token里找找了,default:true
	MaxLiftTime int //cookie在浏览器存活时间，default:0，即浏览器关闭清理
	GCTime int64	//default:
	HTTPOnly bool //default:true，js不能获取到cookie，防止session劫持
	Secure bool //只在https下使用cookie？default:false
	//ProviderConfig          string `json:"providerConfig"`
	Domain                  string //default:""
	Path string
	SessionIDLength         uint8  //default:64
}

//动态调整session的配置
type ManagerRunConfig struct {
	EnableSetCookie bool //是否使用cookie，若 false 优先使用header头部重写，其次使用URL重写，再其次会向query或form中查询token，否则就不能创建session，default:true
	EnableSidInHTTPHeader   bool   //如果cookie不可用，则使用header重写，default:true
	EnableSidInURLQuery     bool   //使用url重写，要是再不允许那我也没办法了,只能在form的token里找找了,default:true
	//优先级 cookie > header > url > query/form token
	HTTPOnly bool //default:true，js不能获取到cookie，防止session劫持
	Secure bool //只在https下使用cookie？default:false
	MaxLiftTime int //cookie在浏览器存活时间，default:0，即浏览器关闭清理
	SessionIDLength         uint8  //default:64
	SessionNameInHTTPHeader string //default:"wtf"
	//ProviderConfig          string `json:"providerConfig"`
	Path string
	Domain                  string //default:""
}

//全局session管理器,Manager并不是并发安全的，应当在provider层实现并发安全
type Manager struct {
	provider Provider
	conf ManagerConf
}

func NewManager(conf ManagerConf) (*Manager,error) {
	provider,ok := providers[conf.ProviderName]
	if !ok{
		return nil,fmt.Errorf("session: unknown provider %q 查查是不是没导包",conf.ProviderName)
	}
	return &Manager{provider:provider,conf:conf},nil
}

//查询到session就返回，否则返回一个新的session
func (m *Manager) Session(w http.ResponseWriter,r *http.Request,confs ...*ManagerRunConfig) (Sessioner,error) {
	cf := m.mergeConf(confs)
	//如果sid存在则返回，否则创建一个新的返回
	sid, ok := m.getSid(w,r,cf)
	if !ok {
		return nil,errors.New("找不到session，你多半把所有存放sid的方式都禁用了")
	}
	if exist := m.provider.Exist(sid);exist{
		return m.provider.Read(sid)
	}else{
		return m.provider.Create(sid)
	}
}



func (m *Manager) ReSessionID(w http.ResponseWriter,r *http.Request)  {

}

func (m *Manager)mergeConf(confs []*ManagerRunConfig) *ManagerConf {
	conf := confs[0]
	cf := m.conf
	if len(confs) > 0 {
		cf.EnableSetCookie = conf.EnableSetCookie
		cf.EnableSidInHTTPHeader = conf.EnableSidInHTTPHeader
		cf.SessionNameInHTTPHeader = conf.SessionNameInHTTPHeader
		cf.EnableSidInURLQuery = conf.EnableSidInURLQuery
		cf.MaxLiftTime = conf.MaxLiftTime
		cf.HTTPOnly = conf.HTTPOnly
		cf.Secure = conf.Secure
		cf.Domain = conf.Domain
		cf.Path = conf.Path
		cf.SessionIDLength = conf.SessionIDLength
	}
	return &cf
}

//cookie>header>url>query/form中依次查找，写入也是同样的顺序
func (m *Manager) getSid(w http.ResponseWriter,r *http.Request,cf *ManagerConf) (string,bool) {
	cookie, err := r.Cookie(cf.CookieName)
	if err == nil && cookie.Value != ""{
		sid,_ := url.QueryUnescape(cookie.Value)
		return sid,true
	}
	if sid := r.Header.Get(cf.SessionNameInHTTPHeader); sid != ""{
		return sid,true
	}
	if sid := r.URL.Query().Get(cf.CookieName); sid != ""{
		return sid,true
	}
	if sid := r.Form.Get(cf.CookieName);sid != ""{
		return sid,true
	}

	//查找不到，setsid
	sid := m.createSID(cf.SessionIDLength)
	for sid == "" || m.provider.Exist(sid){
		sid = m.createSID(cf.SessionIDLength)
	}
	if cf.EnableSetCookie {
		http.SetCookie(w,&http.Cookie{
			Name:cf.CookieName,
			Value:sid,
			Path:cf.Path,
			Domain:cf.Domain,
			MaxAge:cf.MaxLiftTime,
			Secure:cf.Secure,
			HttpOnly:cf.HTTPOnly,
		})
		return sid,true
	}
	if cf.EnableSidInHTTPHeader {
		w.Header().Add(cf.SessionNameInHTTPHeader,sid)
		return sid,true
	}
	if cf.EnableSidInURLQuery {
		u := r.RequestURI
		if len(u) == len(r.URL.EscapedPath()){
			u = fmt.Sprintf("%s?%s=%s",u,cf.CookieName,sid)
		}else{
			u = fmt.Sprintf("%s&%s=%s",u,cf.CookieName,sid)
		}
		http.Redirect(w,r,u,302)
		return sid,true
	}
	return "",false
}

func (m *Manager) createSID(l uint8) string {
	if l == 0{
		l = 64
	}
	id := make([]byte,l)
	if _, err := io.ReadFull(rand.Reader,id); err != nil{
		return ""
	}
	return url.QueryEscape(base64.URLEncoding.EncodeToString(id))
}

