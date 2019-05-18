package session

//实现session存储的接口
type Provider interface {
	Create(sid string) (Sessioner,error)
	Read(sid string) (Sessioner,error)
	Delete(sid string) error
	Reset()
	ReSid(string) (string,error)
	GC(maxLifeTime int64)
	Exist(sid string) bool
}

var Providers = make(map[string]Provider)

//驱动应当主动调用Register，把自己注入
func Register(name string,provider Provider)  {
	if provider == nil{
		panic("session: "+name+"是 nil，provider 不能为空")
	}
	if _,ok := Providers[name];ok{
		panic("session: "+name+"已经被使用，换个名试试？")
	}
	Providers[name] = provider
}