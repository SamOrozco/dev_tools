package main

import "github.com/labstack/gommon/color"

func main() {
	color.RedBg("red")
	color.Blue("blue")
	println(color.BlueBg("blue"))
}
