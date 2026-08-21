package car

import "fmt"

type Car struct {
	engineCapacity int
	meterConsole   string
	color          string
}

type CarBuilder interface {
	SetEngineCapacity(capacity int) CarBuilder
	SetMeterConsole(consoleType string) CarBuilder
	SetColor(color string) CarBuilder
	Build() Car
}

type NormalCarBuilder struct {
	c Car
}

func NewNormalCarBuilder() CarBuilder {
	return &NormalCarBuilder{}
}

func (cb *NormalCarBuilder) SetEngineCapacity(capacity int) CarBuilder {
	cb.c.engineCapacity = capacity
	return cb
}

func (cb *NormalCarBuilder) SetMeterConsole(consoleType string) CarBuilder {
	cb.c.meterConsole = consoleType
	return cb
}

func (cb *NormalCarBuilder) SetColor(color string) CarBuilder {
	cb.c.color = color
	return cb
}

func (cb *NormalCarBuilder) Build() Car {
	return cb.c
}

func (c Car) String() string {
	return fmt.Sprintf(
		"Car created with color: %s, engine capacity: %d and console type: %s",
		c.color,
		c.engineCapacity,
		c.meterConsole,
	)
}
