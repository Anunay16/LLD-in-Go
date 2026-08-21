package main

import (
	"builder_pattern/car"
	"fmt"
)

func main() {
	builder := car.NewNormalCarBuilder()

	myCar := builder.
		SetColor("Red").
		SetEngineCapacity(2000).
		SetMeterConsole("Digital").
		Build()

	fmt.Println(myCar)
}
