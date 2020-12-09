package main

import (
	"github.com/go-vgo/robotgo"
	"github.com/spf13/cobra"
	"time"
)

var (
	DurationMinutes int // search dirs only flag name
	rootCmd         = &cobra.Command{
		Use:   "jit",
		Short: "mouse hitter",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			RunJit(DurationMinutes)
		},
	}
)

func main() {
	rootCmd.PersistentFlags().IntVarP(&DurationMinutes, "duration-minutes", "d", 30, "jitter mouse")
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func RunJit(durationMinutes int) {
	go func() {
		for {
			x, y := robotgo.GetMousePos()
			robotgo.MoveMouseSmooth(x-100, y, 1.0, 25.0)
			robotgo.MoveMouseSmooth(x+100, y-100, 1.0, 25.0)
			robotgo.MoveMouseSmooth(x+100, y, 1.0, 25.0)
			robotgo.MoveMouseSmooth(x, y, 1.0, 25.0)
			<-time.After(time.Millisecond * 1000)
		}
	}()
	<-time.After(time.Minute * time.Duration(durationMinutes))
}
