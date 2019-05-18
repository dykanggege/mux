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

var DefaultManager = defaultManager()

var DefaultManagerConf = &ManagerConf{
	ProviderName :"memory",
	CookieName:"wtf",
	EnableSetCookie :true,
	MaxLiftTime:0,
	GCTime:60*60,
	HTTPOnly:true,
	SessionIDLength:64,
}

func defaultManager() Manager {
	manage, err := NewManage(DefaultManagerConf)
	if err != nil {
		panic(err)
	}
	return manage
}

type ManagerConf struct {
	ProviderName string //default:memory
	CookieName string //default:randstring
	EnableSetCookie bool //true，使用cookie，false 优先使用header头部重写，其次使用URL重写，default:true
	EnableSidInHTTPHeader   bool   //如果cookie不可用，则使用header重写，default:true
	SessionNameInHTTPHeader string //default:"wtf"
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

type Manager interface {
	Session(w http.ResponseWriter,r *http.Request,confs ...*ManagerRunConfig) (Sessioner,error)
	ReSessionID(w http.ResponseWriter,r *http.Request,confs ...*ManagerRunConfig) (string,error)
	SessionID() (string,error)
	SessionByID(sid string) (Sessioner,error)
}

//全局session管理器,Manager并不是并发安全的，应当在provider层实现并发安全
type Manage struct {
	provider Provider
	conf *ManagerConf
}

func (m *Manage) SessionID() (string,error) {
	sid := m.createSID(m.conf.SessionIDLength)
	_, err := m.provider.Create(sid)
	return sid,err
}

func (m *Manage) SessionByID(sid string) (Sessioner,error) {
	return m.provider.Read(sid)
}

func NewManage(conf *ManagerConf) (*Manage,error) {
	provider,ok := Providers[conf.ProviderName]
	if !ok{
		return nil,fmt.Errorf("session: unknown provider %q 查查是不是没导包",conf.ProviderName)
	}
	return &Manage{provider: provider,conf:conf},nil
}

//查询到session就返回，否则返回一个新的session
func (m *Manage) Session(w http.ResponseWriter,r *http.Request,confs ...*ManagerRunConfig) (Sessioner,error) {
	var cf *ManagerConf
	if len(confs) > 0 {
		cf = m.mergeConf(confs)
	}

	sid, ok := m.getSid(w,r,cf)
	if ok {
		return m.provider.Read(sid)
	}

	sid, err := m.setSid(w, r, cf)
	if err != nil{
		return nil,err
	}
	return m.provider.Create(sid)
}

//给session更换id
func (m *Manage) ReSessionID(w http.ResponseWriter,r *http.Request,confs ...*ManagerRunConfig) (string,error)  {
	var cf *ManagerConf
	if len(confs) > 0 {
		cf = m.mergeConf(confs)
	}

	sid, ok := m.getSid(w, r, cf)
	if ok && m.provider.Exist(sid){
		reSid, err := m.provider.ReSid(sid)
		if err != nil { return "",err}
		return reSid,m.resetSid(w,r,reSid,cf)
	}
	return "",errors.New("session并不存在")
}

func (m *Manage) mergeConf(confs []*ManagerRunConfig) *ManagerConf {
	if len(confs) < 1{
		return m.conf
	}
	conf := confs[0]
	cf := m.conf
	cf.EnableSetCookie = conf.EnableSetCookie
	cf.EnableSidInHTTPHeader = conf.EnableSidInHTTPHeader
	cf.SessionNameInHTTPHeader = conf.SessionNameInHTTPHeader
	cf.MaxLiftTime = conf.MaxLiftTime
	cf.HTTPOnly = conf.HTTPOnly
	cf.Secure = conf.Secure
	cf.Domain = conf.Domain
	cf.Path = conf.Path
	cf.SessionIDLength = conf.SessionIDLength
	return cf
}

//cookie>header>url>query/form中依次查找，写入也是同样的顺序
func (m *Manage) getSid(w http.ResponseWriter,r *http.Request,cf *ManagerConf) (string,bool) {
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
	return "",false
}

//创建一个session，并将id写入到cookie中
func (m *Manage) setSid(w http.ResponseWriter,r *http.Request,cf *ManagerConf) (string,error) {
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
		return sid,nil
	}
	return "", errors.New("无法设置session")
}

//更换cookie的sid
func (m *Manage) resetSid(w http.ResponseWriter,r *http.Request,sid string,cf *ManagerConf) error {
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
		return nil
	}
	if cf.EnableSidInHTTPHeader {
		w.Header().Del(cf.SessionNameInHTTPHeader)
		w.Header().Add(cf.SessionNameInHTTPHeader,sid)
		return nil
	}
	return errors.New("无法设置session")
}

//随机生成一个sid
func (m *Manage) createSID(l uint8) string {
	if l == 0{
		l = 64
	}
	id := make([]byte,l)
	if _, err := io.ReadFull(rand.Reader,id); err != nil{
		return ""
	}
	return url.QueryEscape(base64.URLEncoding.EncodeToString(id))
}
