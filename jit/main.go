package main

import (
	"github.com/go-vgo/robotgo"
	"time"
)

func main() {
	for {
		x, y := robotgo.GetMousePos()
		robotgo.MoveMouseSmooth(x-100, y, 1.0, 25.0)
		robotgo.MoveMouseSmooth(x+100, y-100, 1.0, 25.0)
		robotgo.MoveMouseSmooth(x+100, y, 1.0, 25.0)
		robotgo.MoveMouseSmooth(x, y, 1.0, 25.0)
		<-time.After(time.Millisecond * 1000)
	}
}
