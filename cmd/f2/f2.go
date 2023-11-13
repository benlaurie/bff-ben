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
copy <1>, <2> -> <1>        copy <1> to <1> + <2>               0000 000a
call <1> -> <1>             call <1> % PLEN                     0000 000b
stop															0000 000c
loc <> -> <1>               push pc to stack                    0000 000d
*/

package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

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
	COPY  = 0x0a
	CALL  = 0x0b
	STOP  = 0x0c
	LOC   = 0x0d
)

const ULEN = 8192
const SLEN = 16
const MUTATION_RATE = 6_400_000 * 32 / ULEN
const SHOW = 1_000_000
const RUNNERS = 4

func run(program *[ULEN]uint8, pc int) {
	var stack [SLEN]uint16
	sp := 0
	iterations := 0
outer:
	for {
		if iterations > 500 || sp > SLEN || sp < 0 {
			break
		}
		iterations++
		pc %= ULEN
		op := program[pc]
		pc++
		if op&PUSH == PUSH {
			if sp >= SLEN {
				break outer
			}
			stack[sp] = uint16(op & 0x7f)
			sp++
			continue
		}
		switch op {
		case POP:
			sp--
		case NOT:
			if sp < 1 {
				break outer
			}
			stack[sp-1] = ^stack[sp-1]
		case ADD:
			if sp--; sp < 1 {
				break outer
			}
			stack[sp-1] += stack[sp]
		case MUL:
			if sp < 2 {
				break outer
			}
			t := uint32(stack[sp-1]) * uint32(stack[sp-2])
			stack[sp-1] = uint16(t >> 16)
			stack[sp-2] = uint16(t & 0xffff)
		case STORE:
			if sp -= 2; sp < 0 {
				break outer
			}
			program[stack[sp]%ULEN] = uint8(stack[sp+1] & 0xff)
		case DUP:
			if sp < 1 || sp >= SLEN {
				break outer
			}
			stack[sp] = stack[sp-1]
			sp++
		case JUMP:
			if sp--; sp < 0 {
				break outer
			}
			pc = int(stack[sp]) % ULEN
		case JNZ:
			if sp -= 2; sp < 0 {
				break outer
			}
			if stack[sp] != 0 {
				pc = int(stack[sp+1])
			}
		case LOAD:
			if sp < 1 {
				break outer
			}
			stack[sp-1] = uint16(program[stack[sp-1]%ULEN])
			sp++
		case SWAP:
			if sp < 2 {
				break outer
			}
			t := stack[sp-1]
			stack[sp-1] = stack[sp-2]
			stack[sp-2] = t
		case COPY:
			if sp < 2 {
				break outer
			}
			program[(stack[sp-2]+stack[sp-1])%ULEN] = program[stack[sp-2]%ULEN]
			sp--
		case CALL:
			if sp < 1 {
				break outer
			}
			t := stack[sp-1]
			stack[sp-1] = uint16(pc)
			pc = int(t) % ULEN
		case STOP:
			break outer
		case LOC:
			if sp >= SLEN {
				break outer
			}
			stack[sp] = uint16(pc)
			sp++
		}
	}
}

func showp(program [ULEN]uint8) {
	for i := 0; i < ULEN; i++ {
		fmt.Print(charp(program[i]))
		if i%128 == 127 {
			fmt.Print("\n")
		}
	}
}

func charp(instruction uint8) string {
	if instruction&PUSH == PUSH {
		return "P"
	}
	switch instruction {
	case POP:
		return "p"
	case NOT:
		return "~"
	case ADD:
		return "+"
	case MUL:
		return "*"
	case STORE:
		return "S"
	case DUP:
		return "D"
	case JUMP:
		return "<"
	case JNZ:
		return "Z"
	case LOAD:
		return "L"
	case SWAP:
		return "X"
	case COPY:
		return "C"
	case CALL:
		return ">"
	case STOP:
		return "!"
	case LOC:
		return "l"
	default:
		return " "
	}
}

func mutate(program *[ULEN]uint8) {
	switch rand.Intn(2) {
	case 0:
		program[rand.Intn(ULEN)] = uint8(rand.Intn(256))
	case 1:
		program[rand.Intn(ULEN)] = uint8(rand.Intn(13))
	}
}

func runner(universe *[ULEN]uint8) {
	n := 0
	for {
		n++

		run(universe, rand.Intn(ULEN))
		if rand.Intn(MUTATION_RATE) == 0 {
			mutate(universe)
		}
	}
}

func ngrams(universe *[ULEN]uint8, n int) map[string]int {
	m := make(map[string]int)
	for i := 0; i < ULEN-n; i++ {
		gram := ""
		for j := 0; j < n; j++ {
			gram += charp(universe[i+j])
		}
		m[gram]++
	}
	return m
}

func showngrams(universe *[ULEN]uint8, n int) {
	m := ngrams(universe, n)
	keys := make([]string, 0, len(m))

	for key := range m {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return m[keys[i]] < m[keys[j]]
	})

	top := m[keys[len(keys)-1]]
	l := 0
	for n := len(keys) - 1; n >= 0 && l < 8; n-- {
		l++
		key := keys[n]
		if m[key] <= top/10 {
			break
		}
		fmt.Printf("% 15s % 4d ", key, m[key])
	}
}

func main() {
	var universe [ULEN]uint8

	for i := 0; i < ULEN; i++ {
		universe[i] = uint8(rand.Intn(256))
	}

	for i := 0; i < RUNNERS; i++ {
		go runner(&universe)
	}

	for {
		var u2 [ULEN]uint8
		copy(u2[:], universe[:])
		fmt.Print("\033c")
		showp(u2)
		for i := 2; i < 16; i++ {
			showngrams(&u2, i)
			fmt.Print("\n")
		}

		time.Sleep(1 * time.Second)
	}
}
