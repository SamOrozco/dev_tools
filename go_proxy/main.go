package main

import (
	"dev_tools/files"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
)

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

	listener, err := net.Listen("tcp", ":9005")
	if err != nil {
		panic(err)
	}

	for {
		con, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleIncomingRequest(con)
	}
}

func handleIncomingRequest(con net.Conn) {
	data ,err := ioutil.ReadAll(con)
	if err != nil {
		panic(err)
	}
	println(string(data))
	con.Close()
}
