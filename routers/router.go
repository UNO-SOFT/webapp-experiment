package routers

import (
	"github.com/UNO-SOFT/webapp-experiment/WE/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
