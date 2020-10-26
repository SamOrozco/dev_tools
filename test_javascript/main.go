package main

import "github.com/robertkrimen/otto"

var vm = otto.New()

func main() {

	_, err := vm.Run(`
var status = 200;
var def = status == 200;
`)

	if err != nil {
		panic(err)
	}

	val, err := vm.Get("def")
	if err != nil {
		panic(err)
	}
	println(val.String())

}
