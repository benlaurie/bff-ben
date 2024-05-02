package main

import (
	"strconv"

	"github.com/Krawabbel/go-8080/intel8080"
)

const REGION_MASK = uint16(0xfff0)

type RAM [0x10000]byte

func (r RAM) Read(addr uint16) byte {
	return r[addr]
}

func (r RAM) Write(addr uint16, data byte) {
	r[addr] = data
}

type CPUWithRAM struct {
	cpu *intel8080.Intel8080
	ram *RAM
}

func (c *CPUWithRAM) Step() {
	c.cpu.Step()
}

func (c *CPUWithRAM) Read(addr uint16) byte {
	return c.ram.Read(c.adjustAddr(addr))
}

func (c *CPUWithRAM) Write(addr uint16, data byte) {
	c.ram.Write(c.adjustAddr(addr), data)
}

func (c *CPUWithRAM) PC() uint16 {
	s := c.cpu.DebugState()
	pc, _ := strconv.ParseUint(s[6:10], 16, 16)
	return uint16(pc)
}

func (c *CPUWithRAM) adjustAddr(addr uint16) uint16 {
	offset := addr & ^REGION_MASK
	pc := c.PC() & REGION_MASK
	return pc | offset
}

func main() {
	ram := RAM{}
	cpu_with_ram := CPUWithRAM{ram: &ram}
	cpu := intel8080.NewIntel8080(&cpu_with_ram, 0)
	cpu_with_ram.cpu = cpu
	cpu.Step()
	cpu.Step()
}
