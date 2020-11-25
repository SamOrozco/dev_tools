package main

import (
	"bufio"
	"dev_tools/files"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
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

	println(fmt.Sprintf("listening on port %d", port))
	for {
		con, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleIncomingRequest(con, matcherProxies...)
	}
}

/**
handles incoming tcp request
*/
func handleIncomingRequest(con net.Conn, mp ...*MatcherProxy) {
	defer con.Close()

	data, err := readAllData(con)
	if err != nil {
		panic(err)
	}

	path := getPathFromData(string(data))

	wg := sync.WaitGroup{}
	for i := range mp {
		currentMatcherCheck := mp[i]
		if currentMatcherCheck.Matcher.Path(path) {
			wg.Add(1)
			go handleProxyMatch(currentMatcherCheck, data, con, &wg)
		}
	}
	wg.Wait()
}

/**
executes proxy when a request is matched
*/
func handleProxyMatch(
	currentMatcherCheck *MatcherProxy,
	requestData []byte,
	originalConnection net.Conn,
	wg *sync.WaitGroup,
) {
	// establish connection
	writeCon, err := net.Dial("tcp", fmt.Sprintf("%s:%d", currentMatcherCheck.Proxy.Host, currentMatcherCheck.Proxy.Port))
	if err != nil {
		panic(err)
	}

	// write data
	if _, err := writeCon.Write(requestData); err != nil {
		panic(err)
	}

	transferWriter := io.TeeReader(writeCon, originalConnection)
	// read all into response hopefully
	ioutil.ReadAll(transferWriter)

	wg.Done()
	writeCon.Close()
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

func readAllData(reader io.Reader) ([]byte, error) {
	if reader == nil {
		return []byte{}, nil
	}

	bufferLength := 2048
	completeReadValue := make([]byte, 0)
	buffer := make([]byte, bufferLength)
	readData := 1
	var err error
	rdr := bufio.NewReader(reader)

	for readData > 0 {
		readData, err = rdr.Read(buffer)
		completeReadValue = append(completeReadValue, buffer...)
		if err != nil {
			return completeReadValue, err
		}

		// if we didn't read the whole buffer we know there shouldn't be any more data, I think!
		if readData < bufferLength {
			return completeReadValue, nil
		}
	}

	return completeReadValue, nil
}
