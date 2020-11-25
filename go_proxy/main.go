package main

import (
	"bufio"
	"dev_tools/files"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

type MatcherProxy struct {
	Matcher Matcher
	Proxy   *Proxy
}

func main() {
	if len(os.Args) < 2 {
		panic("must pass in configuration file")
	}

	configFile := os.Args[1]
	data, err := files.ReadBytesFromFile(configFile)
	if err != nil {
		panic(err)
	}

	proxyConf := &ProxyConfig{}
	err = yaml.Unmarshal(data, proxyConf)
	if err != nil {
		panic(err)
	}
	proxyConfig := buildProxyConfigFromConfig(proxyConf)
	listener, err := net.Listen("tcp", ":9005")
	if err != nil {
		panic(err)
	}

	for {
		con, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleIncomingRequest(con, proxyConfig)
	}
}

func handleIncomingRequest(con net.Conn, mp *MatcherProxy) {
	defer con.Close()
	writer := bufio.NewWriter(con)
	data, err := ioutil.ReadAll(con)
	if err != nil {
		panic(err)
	}

	path := getPathFromData(string(data))
	if mp.Matcher.Path(path) {
		println("matches")

		// establish connection
		writeCon, err := net.Dial("tcp", fmt.Sprintf("%s:%d", mp.Proxy.Host, mp.Proxy.Port))
		defer writeCon.Close()
		if err != nil {
			panic(err)
		}

		// write data
		if _, err := writeCon.Write(data); err != nil {
			panic(err)
		}
		// wire response
		buffer := make([]byte, 2048)
		_, err = writeCon.Read(buffer)
		if err != nil {
			panic(err)
		}

		if _, err = writer.Write(buffer); err != nil {
			panic(err)
		}
	}
}

func buildProxyConfigFromConfig(conf *ProxyConfig) *MatcherProxy {
	if conf.Match == nil || conf.Proxy == nil {
		panic("invalid configuration must configure matcher and proxy")
	}

	if conf.Match.Type == "path" {
		return &MatcherProxy{
			Matcher: NewSimplePathMatcher(conf.Match.PathMatch.MatchValue, conf.Match.PathMatch.MatchType),
			Proxy:   conf.Proxy,
		}
	}

	// todo incomplete
	return nil
}

/**
makes a lot of assumptions but works
*/
func getPathFromData(dataString string) string {
	return strings.Split(strings.Split(dataString, "\n")[0], " ")[1]
}
