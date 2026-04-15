package main

import (
	"fmt"

	"github.com/adit-prawira/nuclear-sim/reactor"
	engine "github.com/adit-prawira/nuclear-sim/sim"
)

func main(){
	fmt.Println("RBMK-1000 Nuclear Reactor Simulator")
	fmt.Println("Chernobyl Nuclear Power Plant — Unit 4")
	fmt.Println("========================================")

	r := reactor.NominalRBMK()
	e := engine.NewRBMKEnginer(r)
	e.Run()
}
