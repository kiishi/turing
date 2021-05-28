package main

import "os"

func main(){
	args := os.Args[1:]
	machine := TuringMachine{}
	machine.StartComputing(args[0])
}