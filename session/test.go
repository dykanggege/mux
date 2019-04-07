package session

import "github.com/astaxie/beego"

func test()  {
	beego.Run()

}

type ttt struct {
	beego.Controller
}

func (t *ttt) Get() {
	t.GetSession("")

}