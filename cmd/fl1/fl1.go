/*
8 bit processor

push <n> -> <1>                                                 1nnn nnnn
pop <1> -> <>                                                   0000 0000
not <1> -> <1>              bitwise not                         0000 0001
add <1>, <2> -> <1>                                             0000 0002
mul <1>, <2> -> <1>, <2>                                        0000 0003
store <1>, <2> -> <>        stores <1> at <2> % PLEN            0000 0004
dup <1> -> <1>, <2>                                             0000 0005
jump <1> -> <>              is also return                      0000 0006
jnz <1>, <2>                jump to <1> % PLEN if <2> nz        0000 0007
load <1> -> <1>                                                 0000 0008
swap <1>, <2> -> <2>, <1>                                       0000 0009
cheat <1>, <2> -> <1>            copy <1> to <1> + <2>          0000 000a
*/

package main

import (
	"fmt"
	"os"
	"time"

	"pgregory.net/rand"
)

const PLEN = 32
const SLEN = 16
const NPROGRAMS = 64
const MUTATION_RATE = 320000 * 32 / NPROGRAMS

const (
	PUSH  = 0x80
	POP   = 0x00
	NOT   = 0x01
	ADD   = 0x02
	MUL   = 0x03
	STORE = 0x04
	DUP   = 0x05
	JUMP  = 0x06
	JNZ   = 0x07
	LOAD  = 0x08
	SWAP  = 0x09
	CHEAT = 0x0a
	CALL  = 0x0b
)

const RLEN = PLEN * 2

func run(program *[RLEN]uint8) {
	var stack [SLEN]uint8
	sp := 0
	pc := 0
	iterations := 0
outer:
	for {
		if iterations > 500 || pc >= RLEN || sp > SLEN || sp < 0 {
			break
		}
		iterations++
		op := program[pc]
		pc++
		if op&PUSH == PUSH {
			if sp >= SLEN {
				break outer
			}
			stack[sp] = op & 0x7f
			sp++
			continue
		}
		switch op {
		case POP:
			if sp > 0 {
				sp--
			}
		case NOT:
			if sp > 0 {
				stack[sp-1] = ^stack[sp-1]
			}
		case ADD:
			if sp > 1 {
				stack[sp-2] += stack[sp-1]
				sp--
			}
		case MUL:
			if sp > 1 {
				t := uint16(stack[sp-1]) * uint16(stack[sp-2])
				stack[sp-1] = uint8(t >> 8)
				stack[sp-2] = uint8(t & 0xff)
			}
		case STORE:
			if sp > 1 {
				program[stack[sp-2]%RLEN] = stack[sp-1]
				sp -= 2
			}
		case DUP:
			if sp > 0 && sp < SLEN {
				stack[sp] = stack[sp-1]
				sp++
			}
		case JUMP:
			if sp > 0 {
				pc = int(stack[sp-1]) % RLEN
				sp--
			}
		case JNZ:
			if sp > 1 {
				if stack[sp-1] != 0 {
					pc = int(stack[sp-2]) % RLEN
				}
				sp -= 2
			}
		case LOAD:
			if sp > 0 {
				stack[sp-1] = program[stack[sp-1]%RLEN]
			}
		case SWAP:
			if sp > 1 {
				t := stack[sp-1]
				stack[sp-1] = stack[sp-2]
				stack[sp-2] = t
			}
		case CHEAT:
			if sp > 1 {
				program[(stack[sp-1]+stack[sp-2])%RLEN] = program[stack[sp-1]%RLEN]
			}
		case CALL:
			if sp > 0 {
				t := stack[sp-1]
				stack[sp-1] = uint8(pc)
				pc = int(t) % RLEN
				sp--
			}
		}
	}
}

func show(program [PLEN]uint8) {
	for i := 0; i < PLEN; i++ {
		if program[i]&PUSH == PUSH {
			fmt.Print("P")
		} else {
			switch program[i] {
			case POP:
				fmt.Print("p")
			case NOT:
				fmt.Print("~")
			case ADD:
				fmt.Print("+")
			case MUL:
				fmt.Print("*")
			case STORE:
				fmt.Print("S")
			case DUP:
				fmt.Print("D")
			case JUMP:
				fmt.Print("<")
			case JNZ:
				fmt.Print("Z")
			case LOAD:
				fmt.Print("L")
			case SWAP:
				fmt.Print("X")
			case CHEAT:
				fmt.Print("C")
			case CALL:
				fmt.Print(">")
			default:
				fmt.Print(" ")
			}
		}
	}
}

func showp(programs [NPROGRAMS][PLEN]uint8) {
	for i := 0; i < NPROGRAMS; i++ {
		show(programs[i])
		fmt.Print("\n")
	}
}

func mutate(program *[PLEN]uint8, programs *[NPROGRAMS][PLEN]uint8) {
	r := rand.Intn(7)
	switch r {
	case 0:
		program[rand.Intn(PLEN)] = uint8(rand.Intn(10)) + 1
	case 1:
		program[rand.Intn(PLEN)] = uint8(rand.Intn(256))
	case 2:
		n := rand.Intn(PLEN) + 1
		var tmp [PLEN]uint8
		copy(tmp[:], program[:n])
		copy(program[:], program[n:])
		copy(program[PLEN-n:], tmp[:])
	case 3:
		n := rand.Intn(PLEN)
		p := rand.Intn(NPROGRAMS)
		copy(program[n:], programs[p][n:])
	case 4:
		n := rand.Intn(PLEN)
		p := rand.Intn(NPROGRAMS)
		copy(program[:n], programs[p][:n])
	case 5:
		n := rand.Intn(PLEN)
		copy(program[n:], program[n+1:])
	case 6:
		n := rand.Intn(PLEN)
		copy(program[n+1:], program[n:])
	}
}

func count_ops(program *[PLEN]uint8, count *[256]uint) {
	for i := 0; i < PLEN; i++ {
		count[program[i]]++
	}
}

const SHOW = 10_000_000

func main() {
	f := fmt.Sprintf("log.%s", time.Now().Format("2006-01-02-15:04:05"))
	log, err := os.Create(f)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	var programs [NPROGRAMS][PLEN]uint8
	for i := 0; i < NPROGRAMS; i++ {
		for j := 0; j < PLEN; j++ {
			programs[i][j] = uint8(rand.Intn(256))
		}
	}
	showp(programs)
	m := 0
	n := 0
	generation := 0
	start := time.Now()
	for {
		m++
		generation++
		if m > MUTATION_RATE {
			mutate(&programs[rand.Intn(NPROGRAMS)], &programs)
			m = 0
		}
		p1 := rand.Intn(NPROGRAMS)
		p2 := rand.Intn(NPROGRAMS)

		var merged [PLEN * 2]uint8
		copy(merged[:PLEN], programs[p1][:PLEN])
		copy(merged[PLEN:], programs[p2][:PLEN])

		run(&merged)

		copy(programs[p1][:PLEN], merged[:PLEN])
		copy(programs[p2][:PLEN], merged[PLEN:])

		if n++; n == SHOW {
			var count [256]uint
			fmt.Print("\033c")
			fmt.Printf("%d\n", generation)
			showp(programs)
			fmt.Printf("Time: %s\n", time.Since(start))
			start = time.Now()

			for i := 0; i < NPROGRAMS; i++ {
				count_ops(&programs[i], &count)
			}

			for i := 0; i < 256; i++ {
				fmt.Printf("% 4d ", count[i])
				if i%16 == 15 {
					fmt.Print("\n")
					if i == 127 {
						fmt.Print("\n")
					}
				}
				fmt.Fprintf(log, "%d,", count[i])
			}
			log.WriteString("\n")
			n = 0
		}
	}
}
