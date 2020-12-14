package main

import (
	"github.com/gen2brain/beeep"
	"github.com/spf13/cobra"
	"time"
)

var (
	DurationMinutes int
	rootCmd         = &cobra.Command{
		Use:   "remind",
		Short: "send you a reminder in a specified amount of time",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			var reminder string
			if len(args) < 1 {
				reminder = "you wanted a reminder"
			} else {
				reminder = args[0]
			}
			<-time.After(time.Minute * time.Duration(DurationMinutes))
			sendReminder(reminder)
		},
	}
)

func main() {
	rootCmd.PersistentFlags().IntVarP(&DurationMinutes, "duration-minutes", "d", 15, "how long until I remind you")
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func sendReminder(reminder string) {
	if err := beeep.Alert("Reminder", reminder, ""); err != nil {
		panic(err)
	}
}
