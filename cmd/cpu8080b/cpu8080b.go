package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/GinjaNinja32/go-i8080"
)

const REGION_MASK = uint16(0xfff0)
const TOP_BIT = ^REGION_MASK & ((^REGION_MASK) >> 1)

const ULEN = 0x10000
const ILIMIT = 1_000            // clocks
const MUTATION_RATE = 4_000_000 // Higher is less mutation
const RUNNERS = 16

type RAM struct {
	ram *[ULEN]byte
}

type CPUWithRAM struct {
	i8080.CPU
	ram  RAM
	base uint16
}

func (r RAM) Read(addr uint16) uint8 {
	return r.ram[addr]
}

func (r RAM) Write(addr uint16, data uint8) {
	r.ram[addr] = data
}

/*
func (c *CPUWithRAM) Step() uint64 {
	return c.cpu.Step()
}
*/

func (c *CPUWithRAM) Read(addr uint16) uint8 {
	return c.ram.Read(c.adjustAddr(addr))
}

func (c *CPUWithRAM) Write(addr uint16, data uint8) {
	c.ram.Write(c.adjustAddr(addr), data)
}

/*
	func (c *CPUWithRAM) PC() uint16 {
		return c.cpu.PC
	}
*/
func (c *CPUWithRAM) adjustAddr(addr uint16) uint16 {
	return c.base + addr
}

/*
	func (c *CPUWithRAM) NewCPU(pc uint16) *i8080.CPU {
		cpu := i8080.Basic()
		return cpu
	}
*/
func (c *CPUWithRAM) run(pc uint16) uint64 {
	c.InitBasic(c)
	c.base = pc // this will cause a PC of 0 to be at pc
	t := uint64(0)
	for {
		n := c.Step()
		t += n
		if c.Halted || t > ILIMIT {
			return t
		}
		if n == 0 {
			panic("n == 0")
		}
	}
}

func runner(cpu *CPUWithRAM, generation *uint64, n_ops *uint64) {
	t := uint64(0)
	for {
		n := cpu.run(uint16(rand.Intn(ULEN)))
		*n_ops += uint64(n)
		t += n
		for t > MUTATION_RATE {
			//cpu.ram[rand.Intn(ULEN)] ^= uint8(1) << rand.Intn(8)
			cpu.ram.ram[rand.Intn(ULEN)] = uint8(rand.Intn(256))
			t -= MUTATION_RATE
		}
		*generation++
	}
}

var show_off = 0

const SHOW_LEN = 4096

func showp(ram *[ULEN]byte) {
	//show_off += SHOW_LEN
	if show_off >= ULEN {
		show_off -= ULEN
	}
	for i := 0; i < SHOW_LEN; i++ {
		b := ram[(i+show_off)%ULEN]
		/*
			if b >= 0x20 && b < 0x7f {
				fmt.Printf("%c", b)
			} else {
				fmt.Printf(".")
			}
		*/
		fmt.Printf("%02x", b)
		if i%64 == 63 {
			fmt.Print("\n")
		} else {
			fmt.Print(" ")
		}

	}
}

func main() {
	f := fmt.Sprintf("logs/cpu8080.log.%s.%s", "", time.Now().Format("2006-01-02-15:04:05"))
	log, err := os.Create(f)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	binary.Write(log, binary.LittleEndian, uint64(ULEN))
	binary.Write(log, binary.LittleEndian, uint64(0)) // SLEN
	binary.Write(log, binary.LittleEndian, uint64(ILIMIT))
	binary.Write(log, binary.LittleEndian, uint64(MUTATION_RATE))
	binary.Write(log, binary.LittleEndian, uint64(RUNNERS))

	ram := RAM{}
	var dram [ULEN]byte
	ram.ram = &dram

	for i := 0; i < ULEN; i++ {
		//universe[i] = 0x3f
		ram.ram[i] = uint8(rand.Intn(256))
		//mutate(&universe)
	}

	var generation uint64
	var n_ops uint64
	for i := 0; i < RUNNERS; i++ {
		cpu_with_ram := CPUWithRAM{ram: ram}
		go runner(&cpu_with_ram, &generation, &n_ops)
	}

	go func() {
		p_n_ops := uint64(0)
		p_generation := uint64(0)
		t := time.Now()
		for {
			t2 := time.Now()
			var u2 [ULEN]byte
			copy(u2[:], ram.ram[:])
			//dump(log, generation, n_ops, &u2)
			if generation != p_generation {
				fmt.Println("\033c", generation, n_ops, generation-p_generation, n_ops-p_n_ops, float64(n_ops-p_n_ops)/t2.Sub(t).Seconds(), (n_ops-p_n_ops)/(generation-p_generation))
			} else {
				fmt.Println("\033c", generation, n_ops, generation-p_generation, n_ops-p_n_ops)
			}
			t = t2
			p_n_ops = n_ops
			p_generation = generation
			showp(&u2)
			/*
				for i := 2; i < 16; i++ {
					showngrams(&u2, i)
					fmt.Print("\n")
				}
			*/
			time.Sleep(1 * time.Second)
		}
	}()

	select {}
}
