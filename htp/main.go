package main

import (
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Options struct {
	Method string
}

var (
	HttpMethod string // search dirs only flag name
	rootCmd    = &cobra.Command{
		Use:   "htp",
		Short: "cli http tool",
		Long:  `curl wanna be`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				panic("must supply a url")
			}
			handleRequest(args[0], &Options{
				Method: HttpMethod,
			})
		},
	}
)

func main() {
	rootCmd.PersistentFlags().StringVarP(&HttpMethod, "method", "m", "GET", "Http method")
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func handleRequest(urlString string, opt *Options) {
	uri, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}

	request := &http.Request{
		Method: strings.ToUpper(opt.Method),
		URL:    uri,
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}
	ioutil.ReadAll(io.TeeReader(resp.Body, os.Stdout))
}
