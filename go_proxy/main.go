package main

import (
	"bufio"
	"dev_tools/files"
	"fmt"
	"gopkg.in/yaml.v2"
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
	handleConfigFile(configFile)
}

/**
program entry
*/
func handleConfigFile(fileLocation string) {
	data, err := files.ReadBytesFromFile(fileLocation)
	if err != nil {
		panic(err)
	}

	proxyConf := &ProxyConfig{}
	err = yaml.Unmarshal(data, proxyConf)
	if err != nil {
		panic(err)
	}

	port := 9005
	if proxyConf.Port != 0 {
		port = proxyConf.Port
	}

	// builds path matches
	if proxyConf.Proxies == nil {
		panic("must supply at least one proxy")
	}

	// collect matcher proxies
	matcherProxies := make([]*MatcherProxy, len(proxyConf.Proxies))
	for i := range proxyConf.Proxies {
		prox := proxyConf.Proxies[i]
		matcherProxies[i] = buildProxyConfigFromConfig(prox.Match, prox.Proxy)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	for {
		con, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleIncomingRequest(con, matcherProxies...)
	}
}

func handleIncomingRequest(con net.Conn, mp ...*MatcherProxy) {
	rdr := bufio.NewReader(con)
	data := make([]byte, 2048) // todo improve this

	if _, err := rdr.Read(data); err != nil {
		panic(err)
	}

	path := getPathFromData(string(data))

	for i := range mp {
		currentMatcherCheck := mp[i]
		if currentMatcherCheck.Matcher.Path(path) {

			// establish connection
			writeCon, err := net.Dial("tcp", fmt.Sprintf("%s:%d", currentMatcherCheck.Proxy.Host, currentMatcherCheck.Proxy.Port))
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

			if _, err = con.Write(buffer); err != nil {
				panic(err)
			}
			if err := writeCon.Close(); err != nil {
				panic(err)
			}
			if err := con.Close(); err != nil {
				panic(err)
			}
		}
	}
}

func handleMatcherProxy(mp *MatcherProxy) {

}

/**
builds proxy config from matcher and proxy
*/
func buildProxyConfigFromConfig(match *Match, proxy *Proxy) *MatcherProxy {
	if match == nil || proxy == nil {
		panic("invalid configuration must configure matcher and proxy")
	}

	if match.Type == "path" {
		return &MatcherProxy{
			Matcher: NewSimplePathMatcher(match.PathMatch.MatchValue, match.PathMatch.MatchType),
			Proxy:   proxy,
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
