package session

import (
	"crypto/rand"
	"encoding/base64"
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
	//EnableSidInURLQuery     bool   //使用url重写，要是再不允许那我也没办法了,default:true
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
	EnableSetCookie bool //true，使用cookie，false 优先使用header头部重写，其次使用URL重写，default:true
	EnableSidInHTTPHeader   bool   //如果cookie不可用，则使用header重写，default:true
	//EnableSidInURLQuery     bool   //使用url重写，要是再不允许那我也没办法了,default:true
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

func (m *Manager) Session(w http.ResponseWriter,r *http.Request,confs ...*ManagerRunConfig) (Sessioner,error) {
	//TODO:实现session主逻辑
	cf := m.mergeConf(confs)

	sid, ok := m.getSid(r)
	if ok {

	}else{
		session, err := m.provider.Create(sid)
		if cf.EnableSetCookie{
			http.SetCookie(w,&http.Cookie{
				Name:cf.CookieName,
				Value:sid,
				Path:cf.Path,
				Domain:cf.Domain,
				MaxAge:cf.MaxLiftTime,
				Secure:cf.Secure,
				HttpOnly:cf.HTTPOnly,
			})
		}else if cf.EnableSidInHTTPHeader {
			if cf.SessionNameInHTTPHeader == ""{
				cf.SessionNameInHTTPHeader = "wtf"
			}
			w.Header().Set(cf.SessionNameInHTTPHeader,sid)
		}else{
		}
		return session,err
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
		//cf.EnableSidInURLQuery = conf.EnableSidInURLQuery
		cf.MaxLiftTime = conf.MaxLiftTime
		cf.HTTPOnly = conf.HTTPOnly
		cf.Secure = conf.Secure
		cf.Domain = conf.Domain
		cf.Path = conf.Path
		cf.SessionIDLength = conf.SessionIDLength
	}

	return &cf
}

func (m *Manager) getSid(r *http.Request,cf *ManagerConf) (string,bool) {
	cookie, err := r.Cookie(m.conf.CookieName)
	if err != nil || cookie.Value == ""{
		sid := m.sessionID(cf.SessionIDLength)
		for sid == "" || m.provider.Exist(sid) {
			sid = m.sessionID(cf.SessionIDLength)
		}


	}else{
		//sid := cookie.Value
		//return m.provider.Read(sid)
	}
	return "", false
}

func (m *Manager)sessionID(l uint8) string {
	if l == 0{
		l = 64
	}
	id := make([]byte,l)
	if _, err := io.ReadFull(rand.Reader,id); err != nil{
		return ""
	}
	return url.QueryEscape(base64.URLEncoding.EncodeToString(id))
}

