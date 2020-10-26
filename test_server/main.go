package main

import "github.com/labstack/echo"

type User struct {
	UserName string `json:"user_name"`
}

func main() {
	e := echo.New()
	e.GET("/test", func(context echo.Context) error {
		return context.JSON(200, User{
			UserName: "sam",
		})
	})
	panic(e.Start(":9005"))
}
