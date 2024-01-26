package main

import (
	"image/color"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/crazy3lf/colorconv"
)

/*
Mini CPU

cpu.Stack is at least wide enough to hold an address.

0000 xxxx	Push [xxxx], sign extended
0001 xxxx	<top> = (<top> << 4) + [xxxx]

0010 0000	Dup
0010 0001	Swap
0010 0010	Pull <top>th element to the top, or push top to -<top>th if <top> is -ve
0010 0011	Push PC
0010 0100	Pop PC
0010 0101	Load
0010 0110	Store
0010 0111	Add
0010 1000	If NZ if <top> is not zero, execute next opcode and skip following, otherwise skip next and execute following
0010 1001	Drop

01xx xxxx	Read Register xxxxxx
10xx xxxx	Write Register xxxxxx

1111 1111	Halt


Copy:			: src dst
		Push PC : src dst pc
	   	Add		: src pc+dst
		Swap	: pc+dst src
		Push PC : pc+dst src pc
		Add		: pc+dst pc+src
		Load	: pc+dst *(pc+src)
		Store	:

** To match f3/f5, this would need to additonally preserve src.
Copy2:			: src offset
		Swap	: offset src
		Push PC : offset src pc
		Add		: offset pc+src
		Dup		: offset pc+src pc+src
		Push 2	: offset pc+src pc+src 2
		Pull	: pc+src pc+src offset
		Add		: pc+src offset+pc+src
		Swap	: pc+src offset+pc+src
		Load	: pc+src *(pc+src+offset)
		Store	:

Copy2ps:		: src offset
		Swap	: offset src
		Dup		: offset src src
		Push PC : offset src src pc
		Add		: offset src pc+src
		Dup		: offset src pc+src pc+src
		Push 3	: offset src pc+src pc+src 3
		Pull	: src pc+src pc+src offset
		Add		: src pc+src pc+src+offset
		Swap	: src pc+src+offset pc+src
		Load	: src pc+src+offset *(pc+src)
		Store	: src

Copy3:			: src dst offset
		Dup		: src dst offset offset
		Push PC : src dst offset offset pc
		Add		: src dst offset pc+offset
		Push 2	: src dst offset pc+offset 2
		Pull	: src offset pc+offset dst
		Add		: src offset pc+offset+dst
		Swap	: src pc+offset+dst offset
		Push 2	: src pc+offset+dst offset 2
		Pull	: pc+offset+dst offset src
		Push PC : pc+offset+dst offset src pc
		Add		: pc+offset+dst offset pc+src
		Add		: pc+offset+dst pc+src+offset
		Load	: pc+offset+dst *(pc+src+offset)
		Store	:


JNZ:			: cond offset
		Push PC : cond offset pc
		Add		: cond pc+offset
		Swap	: pc+offset cond
		If NZ	: pc+offset
		Pop PC	:
		Drop	:
		Halt    :


// Read/Write heads: 0 is read head, 1 is write head

IncRead:
		Read 0
		Push 1
		Add
		Write 0

IncWrite:
		Read 1
		Push 1
		Add
		Write 1

ReadFrom:
		Read 0
		Load

WriteTo:
		Read 1
		Store

SetRead:
		Push PC
		Add
		Write 0

SetWrite:
		Push PC
		Add
		Write 1

*/

const (
	PUSH       = 0x00
	SHIFT_PUSH = 0x10
	DUP        = 0x20
	SWAP       = 0x21
	PULL       = 0x22
	PUSH_PC    = 0x23
	POP_PC     = 0x24
	LOAD       = 0x25
	STORE      = 0x26
	ADD        = 0x27
	IFNZ       = 0x28
	DROP       = 0x29

	READ  = 0x40
	WRITE = 0x80

	HALT = 0xff
)

const STRICT = true

const PLEN = 16  // length of instructiuon programs
const IWIDTH = 8 // width of instruction fields
const ICOUNT = 1 << IWIDTH
const SLEN = 1024     // length of cpu.stack
const MCOUNT = 65_536 // length of cpu.memory
const RWIDTH = 6      // width of register fields
const RCOUNT = 1 << RWIDTH
const ILIMIT = 1_000_000
const MUTATION_RATE = 100_000_000

type instruction struct {
	code   [PLEN]uint8
	uses   uint
	errors uint
}

type cpu struct {
	instructions [ICOUNT]instruction
	memory       [MCOUNT]uint8
	stack        [SLEN]int
	registers    [RCOUNT]int
	pc           int
	sp           int
}

func sign_extend(a uint8) int {
	if a&0x08 == 0x08 {
		return int(a) - 0x10
	}
	return int(a)
}

type programmer struct {
	i *instruction
	c int
}

func new_programmer(i *instruction) programmer {
	return programmer{i, 0}
}

func (p *programmer) append(i uint8) {
	p.i.code[p.c] = i
	p.c++
}

func (p *programmer) push(n int) {
	if n > 7 || n < -8 {
		// FIXME: do multiple pushes
		panic("push out of range")
	}
	p.append(PUSH | uint8(n&0x0f))
}

func (p *programmer) shift_push(n int) {
	if n > 7 || n < -8 {
		panic("shift_push out of range")
	}
	p.append(SHIFT_PUSH | uint8(n&0x0f))
}

func (p *programmer) read_register(n int) {
	if n > 63 || n < 0 {
		panic("register out of range")
	}
	p.append(READ | uint8(n&0x3f))
}

func (p *programmer) write_register(n int) {
	if n > 63 || n < 0 {
		panic("register out of range")
	}
	p.append(WRITE | uint8(n&0x3f))
}

func (i *instruction) init_copy() {
	p := new_programmer(i)
	p.append(SWAP)
	p.append(DUP)
	p.append(PUSH_PC)
	p.append(ADD)
	p.append(DUP)
	p.push(3)
	p.append(PULL)
	p.append(ADD)
	p.append(SWAP)
	p.append(LOAD)
	p.append(STORE)
	p.append(HALT)
}

func (i *instruction) init_jnz() {
	p := new_programmer(i)
	p.append(PUSH_PC)
	p.append(ADD)
	p.append(SWAP)
	p.append(IFNZ)
	p.append(POP_PC)
	p.append(DROP)
	p.append(HALT)
}

func (i *instruction) init_inc_read() {
	p := new_programmer(i)
	p.read_register(0)
	p.push(1)
	p.append(ADD)
	p.write_register(0)
	p.append(HALT)
}

func (i *instruction) init_inc_write() {
	p := new_programmer(i)
	p.read_register(1)
	p.push(1)
	p.append(ADD)
	p.write_register(1)
	p.append(HALT)
}

func (i *instruction) init_read_from() {
	p := new_programmer(i)
	p.read_register(0)
	p.append(LOAD)
	p.append(HALT)
}

func (i *instruction) init_write_to() {
	p := new_programmer(i)
	p.read_register(1)
	p.append(STORE)
	p.append(HALT)
}

func (i *instruction) init_set_read() {
	p := new_programmer(i)
	p.append(PUSH_PC)
	p.append(ADD)
	p.write_register(0)
	p.append(HALT)
}

func (i *instruction) init_set_write() {
	p := new_programmer(i)
	p.append(PUSH_PC)
	p.append(ADD)
	p.write_register(1)
	p.append(HALT)
}

func (i *instruction) init_push(n int) {
	p := new_programmer(i)
	p.push(n)
	p.append(HALT)
}

func (i *instruction) init_shift_push(n int) {
	p := new_programmer(i)
	p.shift_push(n)
	p.append(HALT)
}

func (i *instruction) init_dup() {
	p := new_programmer(i)
	p.append(DUP)
	p.append(HALT)
}

func (i *instruction) init_swap() {
	p := new_programmer(i)
	p.append(SWAP)
	p.append(HALT)
}

func (i *instruction) init_drop() {
	p := new_programmer(i)
	p.append(DROP)
	p.append(HALT)
}

func (i *instruction) init_random() {
	for n := 0; n < PLEN; n++ {
		i.code[n] = uint8(rand.Intn(ICOUNT))
	}
}

func mod(a, b int) int {
	r := a % b
	if r < 0 {
		r += b
	}
	return r
}

func addr(a int) int {
	return mod(a, MCOUNT)
}

func (i *instruction) inner_run(cpu *cpu) (bool, int) {
	count := 0
	skip := 0
	next_pc := cpu.pc + 1
OUTER:
	for n := 0; n < PLEN; n++ {
		if skip > 0 {
			skip--
			if skip == 0 {
				continue
			}
		}
		instruction := i.code[n]
		count++
		if instruction&0xf0 == PUSH {
			if cpu.sp >= SLEN {
				if STRICT {
					return false, count
				}
			} else {
				cpu.stack[cpu.sp] = sign_extend(instruction & 0x0f)
				cpu.sp++
			}
		} else if instruction&0xf0 == SHIFT_PUSH {
			if cpu.sp > 0 {
				cpu.stack[cpu.sp-1] = (cpu.stack[cpu.sp-1] << 4) + int(instruction&0x0f)
			} else if STRICT {
				return false, count
			}
		} else if instruction&0xc0 == READ {
			if cpu.sp >= SLEN {
				if STRICT {
					return false, count
				}
			} else {
				cpu.stack[cpu.sp] = cpu.registers[instruction&0x3f]
				cpu.sp++
			}
		} else if instruction&0xc0 == WRITE {
			if cpu.sp <= 0 {
				if STRICT {
					return false, count
				}
			} else {
				cpu.registers[instruction&0x3f] = cpu.stack[cpu.sp-1]
				cpu.sp--
			}
		} else {
			switch i.code[n] {
			case DUP:
				if cpu.sp > 0 && cpu.sp < SLEN {
					cpu.stack[cpu.sp] = cpu.stack[cpu.sp-1]
					cpu.sp++
				} else if STRICT {
					return false, count
				}
			case SWAP:
				if cpu.sp > 1 {
					cpu.stack[cpu.sp-1], cpu.stack[cpu.sp-2] = cpu.stack[cpu.sp-2], cpu.stack[cpu.sp-1]
				} else if STRICT {
					return false, count
				}
			case PULL:
				// FIXME: Doesn't respect !STRICT
				if cpu.sp > 0 {
					n := cpu.stack[cpu.sp-1]
					cpu.sp--
					if n < 0 {
						n = -n
						if n >= cpu.sp || n < 0 { // Yes, n can be < 0 still
							return false, count
						}
						t := cpu.stack[cpu.sp-1]
						for i := 0; i < n; i++ {
							cpu.stack[cpu.sp-1-i] = cpu.stack[cpu.sp-2-i]
						}
						cpu.stack[cpu.sp-n-1] = t
					} else {
						if n >= cpu.sp {
							return false, count
						}
						t := cpu.stack[cpu.sp-n-1]
						for i := 0; i < n; i++ {
							cpu.stack[cpu.sp-n-1+i] = cpu.stack[cpu.sp-n+i]
						}
						cpu.stack[cpu.sp-1] = t
					}
				} else if STRICT {
					return false, count
				}
			case PUSH_PC:
				if cpu.sp >= SLEN {
					if STRICT {
						return false, count
					}
				} else {
					cpu.stack[cpu.sp] = cpu.pc
					cpu.sp++
				}
			case POP_PC:
				if cpu.sp > 0 {
					next_pc = cpu.stack[cpu.sp-1]
					cpu.sp--
				} else if STRICT {
					return false, count
				}
			case LOAD:
				if cpu.sp > 0 {
					cpu.stack[cpu.sp-1] = int(cpu.memory[addr(cpu.stack[cpu.sp-1])])
				} else if STRICT {
					return false, count
				}
			case STORE:
				if cpu.sp > 1 {
					cpu.memory[addr(cpu.stack[cpu.sp-1])] = uint8(cpu.stack[cpu.sp-2])
					cpu.sp -= 2
				} else if STRICT {
					return false, count
				}
			case ADD:
				if cpu.sp > 1 {
					cpu.stack[cpu.sp-2] += cpu.stack[cpu.sp-1]
					cpu.sp--
				} else if STRICT {
					return false, count
				}
			case HALT:
				break OUTER
			case IFNZ:
				if cpu.sp > 0 {
					cpu.sp--
					if cpu.stack[cpu.sp] != 0 {
						skip = 2
					} else {
						n++ // skip next instruction
					}
				} else if STRICT {
					return false, count
				}
			case DROP:
				if cpu.sp > 0 {
					cpu.sp--
				} else if STRICT {
					return false, count
				}
			default:
				return false, count
			}
		}
	}
	cpu.pc = next_pc
	return true, count
}

func (i *instruction) run(cpu *cpu) (bool, int) {
	i.uses++
	success, count := i.inner_run(cpu)
	if !success {
		i.errors++
	}
	return success, count
}

func (cpu *cpu) run_instr(instr uint8) (bool, int) {
	success, count := cpu.instructions[instr].run(cpu)

	return success, count
}

func (cpu *cpu) init_memory() {
	for i := 0; i < MCOUNT; i++ {
		cpu.memory[i] = uint8(rand.Intn(ICOUNT))
	}
}

func (cpu *cpu) run() int {
	icount := 0
	for {
		//println(cpu.pc)
		r, c := cpu.run_instr(cpu.memory[addr(cpu.pc)])
		if !r {
			break
		}
		if c == 0 {
			panic("zero count")
		}
		icount += c
		if icount > ILIMIT {
			break
		}
	}
	return icount
}

func (cpu *cpu) mutate() {
	cpu.memory[rand.Intn(MCOUNT)] = uint8(rand.Intn(ICOUNT))
}

func (cpu *cpu) runner() {
	icount := 0
	for {
		cpu.pc = rand.Intn(MCOUNT)
		cpu.sp = 0
		icount += cpu.run()
		if icount > MUTATION_RATE {
			cpu.mutate()
		}

	}
}

func main() {
	var cpu cpu

	for i := 0; i < ICOUNT; i++ {
		//cpu.instructions[i].init_random()
		cpu.instructions[i].code[0] = HALT
	}

	//cpu.instructions[0].init_copy()
	cpu.instructions[1].init_jnz()
	cpu.instructions[2].init_inc_read()
	cpu.instructions[3].init_inc_write()
	cpu.instructions[4].init_read_from()
	cpu.instructions[5].init_write_to()
	cpu.instructions[6].init_set_read()
	cpu.instructions[7].init_set_write()
	cpu.instructions[8].init_dup()
	cpu.instructions[9].init_swap()
	cpu.instructions[10].init_drop()

	for i := -8; i < 7; i++ {
		cpu.instructions[i+8+10].init_push(i)
	}

	for i := -8; i < 7; i++ {
		cpu.instructions[i+24].init_shift_push(i)
	}

	cpu.init_memory()

	cpu.sp = 0
	cpu.pc = 0

	go cpu.runner()

	myApp := app.New()
	w := myApp.NewWindow("Universe")
	i_w := myApp.NewWindow("Instructions")
	i2_w := myApp.NewWindow("Instructions 2")

	raster := canvas.NewRasterWithPixels(
		func(x, y, w, h int) color.Color {
			n := x + y*w
			if n >= MCOUNT {
				return color.Black
			}

			op := cpu.memory[n]

			hsl, _ := colorconv.HSLToColor(float64(op)/256.0*360.0, 1.0, 0.5)

			return hsl
		})
	// raster := canvas.NewRasterFromImage()
	w.SetContent(raster)
	w.Resize(fyne.NewSize(256, MCOUNT/256))
	w.Show()

	var i_instructions [ICOUNT]instruction
	var i_prev_instructions [ICOUNT]instruction
	i_max_uses := 1.0
	i_max_errors := 1.0

	i_raster := canvas.NewRasterWithPixels(
		func(x, y, w, h int) color.Color {
			n := y / 5
			c := uint(0)
			max := 0.0
			if x < 10 {
				c = i_instructions[n].uses - i_prev_instructions[n].uses
				max = i_max_uses
			} else if x < 20 {
				c = i_instructions[n].errors - i_prev_instructions[n].errors
				max = i_max_errors
			} else {
				c = uint(n)
				max = ICOUNT
			}

			t := float64(c) / max

			// a) 1.0 is illegal, 0.0 is the same point ... b) it is possible, it seems, for max to be out of sync wtih i_* (presumably if 2 refreshes get merged into 1)
			// FIXME: there must be some way to fix b...
			if t >= 1.0 {
				t = 0.0 // equivalent to 1.0
			}

			hsl, err := colorconv.HSLToColor(t*360.0, 1.0, 0.5)

			if err != nil {
				panic(err)
			}

			return hsl
		})
	i_w.SetContent(i_raster)
	i_w.Resize(fyne.NewSize(20, ICOUNT*5))
	i_w.Show()

	i2_grid := container.NewGridWithColumns(2)
	var i2_text [ICOUNT]*widget.Label
	for n := 0; n < ICOUNT; n++ {
		t := strconv.Itoa(n) + ": " + strconv.Itoa(int(cpu.instructions[n].uses)) + " " + strconv.Itoa(int(cpu.instructions[n].errors))
		i2_text[n] = widget.NewLabel(t)
		i2_grid.Add(i2_text[n])
	}
	i2_w.SetContent(i2_grid)
	i2_w.Resize(fyne.NewSize(200, 500))
	i2_w.Show()

	go func() {
		for {
			time.Sleep(time.Second / 10)

			i_prev_instructions = i_instructions
			i_instructions = cpu.instructions
			mu := uint(0)
			me := uint(0)
			for n := 0; n < ICOUNT; n++ {
				t := i_instructions[n].uses - i_prev_instructions[n].uses
				if t > mu {
					mu = t
				}
				t = i_instructions[n].errors - i_prev_instructions[n].errors
				if t > me {
					me = t
				}

				tt := strconv.Itoa(n) + ": " + strconv.Itoa(int(i_instructions[n].uses-i_prev_instructions[n].uses)) + " " + strconv.Itoa(int(i_instructions[n].errors-i_prev_instructions[n].errors))
				i2_text[n].SetText(tt)
			}
			i2_grid.Refresh()

			i_max_uses = float64(mu)
			i_max_errors = float64(me)
			i_raster.Refresh()

			raster.Refresh()
			/*
				for n := 0; n < ICOUNT; n++ {
					print(n, " ", cpu.instructions[n].uses, " ")
				}
				println()
			*/
		}
	}()

	myApp.Run()
}
