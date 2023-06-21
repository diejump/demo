package main

import (
	"gin-demo/api"
	dao2 "gin-demo/dao"
)

func main() {
	dao2.InitDB()
	api.InitRouter()
}
