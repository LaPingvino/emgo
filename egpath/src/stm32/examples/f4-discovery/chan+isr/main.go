// This example shows how to use channels to divide interrupt handler into two
// parts: fast part - that runs in interrupt context and soft part - that runs
// in user context. Fast part only enqueues events/data and signals to the
// source if it isn't ready to receive next portion. Slow part dequeues
// events/data and handles them. This scheme can be used to implement
// interrupt driven I/O library.
package main

import (
	"arch/cortexm/exce"
	"delay"
	"rtos"

	"stm32/f4/exti"
	"stm32/f4/gpio"
	"stm32/f4/irqs"
	"stm32/f4/periph"
	"stm32/f4/setup"
)

var LED = gpio.D

const (
	Green = 12 + iota
	Orange
	Red
	Blue
)

func init() {
	setup.Performance168(8)

	periph.APB2ClockEnable(periph.SysCfg)
	periph.APB2Reset(periph.SysCfg)
	periph.AHB1ClockEnable(periph.GPIOA | periph.GPIOD)
	periph.AHB1Reset(periph.GPIOA | periph.GPIOD)

	LED.SetMode(Green, gpio.Out)
	LED.SetMode(Orange, gpio.Out)
	LED.SetMode(Red, gpio.Out)
	LED.SetMode(Blue, gpio.Out)

	// Setup external interrupt source: user button.
	gpio.A.SetMode(0, gpio.In)
	exti.L0.Connect(gpio.A)
	exti.L0.RiseTrigEnable()
	exti.L0.IntEnable()

	rtos.IRQ(irqs.Ext0).Enable()

	periph.APB2ClockDisable(periph.SysCfg)
}

var (
	c   = make(chan int, 3)
	led = Green
)

func buttonISR() {
	exti.L0.ClearPending()
	select {
	case c <- led:
		if led++; led > Red {
			led = Green
		}
	default:
		// Signal that c is full.
		LED.SetPin(Blue)
		delay.Loop(1e5)
		LED.ClearPin(Blue)
	}
}

//c:const
//c:__attribute__((section(".InterruptVectors")))
var IRQs = [...]func(){
	irqs.Ext0 - exce.IRQ0: buttonISR,
}

func toggle(led int) {
	LED.SetPin(led)
	delay.Millisec(500)
	LED.ClearPin(led)
	delay.Millisec(500)
}

func main() {
	for {
		toggle(<-c)
	}
}
