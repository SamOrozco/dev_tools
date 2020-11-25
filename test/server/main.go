package main

import (
	"github.com/labstack/echo"
	"io/ioutil"
)

type User struct {
	UserName string `json:"user_name"`
}

func main() {
	e := echo.New()
	e.GET("/test", func(context echo.Context) error {
		println("sam")
		return context.JSON(200, User{
			UserName: "sam",
		})
	})
	e.POST("/test", func(context echo.Context) error {
		data, err := ioutil.ReadAll(context.Request().Body)
		if err != nil {
			return err
		}
		println(string(data))
		return context.JSON(200, User{
			UserName: "sam",
		})
	})
	panic(e.Start(":9006"))
}
