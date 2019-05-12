# Session
## 内置session
mux.Default默认提供了session，你可以在配置文件中使用 httpSession = true 打开，并使用 httpSessionName = cookiename 配置cookie名称

如果没有特殊需求，你应当调用最外层的API

    //如果传入的key是 name.name 形式，则第一个name为session name，第二个为key name
    //否则使用默认的session name，muxDefaultSessionName
    func (c *Context) Session(key string) interface{} {
    	return c.mux.Session.Get(key)
    }

    func (c *Context) SessionSet(key string,val interface{})  {
    	c.mux.Session.Set(key,val)
    }



默认的session存储在memory中，如果你想提供其他的存储方式，应该导入对应的驱动，并在配置文件中设置 httpSessionDriver = drivername

如果想要使用其他的session库，那你应该实现Sessioner接口，并注入到mux实例中

如果你想使用其他的session driver，你应当实现Provider接口，并调用Register方法注入drivername，如同sql库中的驱动一样


## 更底层的方法
Context中的session调用都依赖于 mux.Session.Session方法
