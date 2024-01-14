/*
0000 xxxx	Push [xxxx], sign extended
0001 xxxx	<top> = (<top> << 4) + [xxxx]
0010 0000	Copy *(<pc> + <top - 1>) to *(<pc> + <top - 1> + <top>), pop 1
0010 0001	Inc <top>
0010 0010	Dec <top>
0010 0011	Jump to <pc> + <top - 1> if <top> != 0, pop 2
0010 0100   Duplicate <top>
0010 0101   Swap <top> and <top - 1>
0010 0110   Rotate the top <top> elements, pop 1
0010 0111   Load: replace <top> with *(<pc> + <top>)
0010 1000   Store: store <top - 1> at <pc> + <top>, pop 2
0010 1001   Add <top> and <top - 1>, pop 1
*/

package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"
)

const ULEN = 8192 * 8
const SLEN = 1024
const ILIMIT = 500
const MUTATION_RATE = 80_000 * 32 / ULEN
const RUNNERS = 8
const STRICT = true
const SHOW_LEN = 8192

const (
	PUSH       = 0x00
	SHIFT_PUSH = 0x10
	COPY       = 0x20
	INC        = 0x21
	DEC        = 0x22
	JNZ        = 0x23
	DUP        = 0x24
	SWAP       = 0x25
	ROT        = 0x26
	LOAD       = 0x27
	STORE      = 0x28
	ADD        = 0x29
	MAX_OP     = ADD
)

func pmod(a int, b int) int {
	return (a%b + b) % b
}

func sign_extend(a uint8) int8 {
	if a&0x08 == 0x08 {
		return int8(a - 0x10)
	}
	return int8(a)
}

func run(program *[ULEN]uint8, pc int) int {
	var stack [SLEN]int8
	sp := 0
	iterations := 0

OUTER:
	for {
		if iterations++; iterations > ILIMIT {
			break
		}

		op := program[pc]
		pc = (pc + 1) % ULEN
		if op&0xf0 == PUSH {
			if sp >= SLEN {
				if STRICT {
					break OUTER
				}
			} else {
				stack[sp] = sign_extend(op & 0x0f)
				sp++
			}
		} else if op&0xf0 == SHIFT_PUSH {
			if sp > 0 {
				stack[sp-1] = (stack[sp-1] << 4) + int8(op&0x0f)
			} else if STRICT {
				break OUTER
			}
		} else {
			switch op {
			case COPY:
				if sp > 1 {
					loc := pmod(pc+int(stack[sp-2]), ULEN)
					off := int(stack[sp-1])
					program[pmod(loc+off, ULEN)] = program[loc]
					sp-- // Leave the destination on the stack
					//sp -= 2
				} else if STRICT {
					break OUTER
				}
			case INC:
				if sp > 0 {
					stack[sp-1]++
				} else if STRICT {
					break OUTER
				}
			case DEC:
				if sp > 0 {
					stack[sp-1]--
				} else if STRICT {
					break OUTER
				}
			case JNZ:
				if sp > 1 {
					if stack[sp-1] != 0 {
						pc = pmod(pc+int(stack[sp-2]), ULEN)
					}
					sp -= 2
				} else if STRICT {
					break OUTER
				}
			case DUP:
				if sp > 0 {
					if sp >= SLEN {
						if STRICT {
							break OUTER
						}
					} else {
						stack[sp] = stack[sp-1]
						sp++
					}
				} else if STRICT {
					break OUTER
				}
			case SWAP:
				if sp > 1 {
					stack[sp-1], stack[sp-2] = stack[sp-2], stack[sp-1]
				} else if STRICT {
					break OUTER
				}
			case ROT:
				if sp > 0 { // sp == 3
					n := int(stack[sp-1]) // n == 2
					sp--
					if n > sp {
						if STRICT {
							break OUTER
						}
					} else {
						if n > 0 {
							t := stack[sp-1]
							for i := 0; i > n-1; i-- { // i == 0, sp == 2, n == 2
								stack[sp-i-1] = stack[sp-i-2]
							}
							stack[sp-n] = t
						}
					}
				}
			case LOAD:
				if sp > 0 {
					loc := pmod(pc+int(stack[sp-1]), ULEN)
					stack[sp-1] = int8(program[loc])
				} else if STRICT {
					break OUTER
				}
			case STORE:
				if sp > 1 {
					loc := pmod(pc+int(stack[sp-1]), ULEN)
					program[loc] = uint8(stack[sp-2])
					sp -= 2
				} else if STRICT {
					break OUTER
				}
			case ADD:
				if sp > 1 {
					stack[sp-2] += stack[sp-1]
					sp--
				} else if STRICT {
					break OUTER
				}
			}
		}
	}
	return iterations
}

func charp(op uint8) string {
	if op&0xf0 == PUSH {
		return "P"
	} else if op&0xf0 == SHIFT_PUSH {
		return "S"
	} else {
		switch op {
		case COPY:
			return "C"
		case INC:
			return ">"
		case DEC:
			return "<"
		case JNZ:
			return "J"
		case DUP:
			return "="
		case SWAP:
			return "X"
		case ROT:
			return "R"
		case LOAD:
			return "^"
		case STORE:
			return "v"
		case ADD:
			return "+"
		}
	}
	return " "
}

var show_off = 0

func showp(program *[ULEN]uint8) {
	//show_off += SHOW_LEN
	if show_off >= ULEN {
		show_off -= ULEN
	}
	for i := 0; i < SHOW_LEN; i++ {
		fmt.Print(charp(program[(i+show_off)%ULEN]))
		if i%128 == 127 {
			fmt.Print("\n")
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
		fmt.Printf("% 15s% 5d ", key, m[key])
	}
}

func mutate(program *[ULEN]uint8) {
	switch rand.Intn(5) {
	case 0:
		program[rand.Intn(ULEN)] = uint8(rand.Intn(256))
	case 1:
		program[rand.Intn(ULEN)] = uint8(rand.Intn(MAX_OP + 1))
	default:
		program[rand.Intn(ULEN)] = 0x20 + uint8(rand.Intn(MAX_OP+1-0x20))
	}
}

func runner(universe *[ULEN]uint8, generation *uint64, n_ops *uint64) {
	n := 0
	for {
		n++

		*n_ops += uint64(run(universe, rand.Intn(ULEN)))
		if rand.Intn(MUTATION_RATE) == 0 {
			mutate(universe)
		}
		*generation++
	}
}

func dump(f *os.File, generation uint64, n_ops uint64, universe *[ULEN]uint8) {
	binary.Write(f, binary.LittleEndian, generation)
	binary.Write(f, binary.LittleEndian, n_ops)
	for i := 0; i < ULEN; i++ {
		binary.Write(f, binary.LittleEndian, universe[i])
	}
}

func main() {
	strict := "lenient"
	if STRICT {
		strict = "strict"
	}
	f := fmt.Sprintf("logs/f5.log.%s.%s", strict, time.Now().Format("2006-01-02-15:04:05"))
	log, err := os.Create(f)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	binary.Write(log, binary.LittleEndian, uint64(ULEN))
	binary.Write(log, binary.LittleEndian, uint64(SLEN))
	binary.Write(log, binary.LittleEndian, uint64(ILIMIT))
	binary.Write(log, binary.LittleEndian, uint64(MUTATION_RATE))
	binary.Write(log, binary.LittleEndian, uint64(RUNNERS))

	var universe [ULEN]uint8

	for i := 0; i < ULEN; i++ {
		universe[i] = 0x3f
	}

	var generation uint64
	var n_ops uint64
	for i := 0; i < RUNNERS; i++ {
		go runner(&universe, &generation, &n_ops)
	}

	p_n_ops := uint64(0)
	p_generation := uint64(0)
	for {
		var u2 [ULEN]uint8
		copy(u2[:], universe[:])
		dump(log, generation, n_ops, &u2)
		fmt.Println("\033c", generation, n_ops, generation-p_generation, n_ops-p_n_ops, (n_ops-p_n_ops)/(generation-p_generation))
		p_n_ops = n_ops
		p_generation = generation
		showp(&u2)
		for i := 2; i < 16; i++ {
			showngrams(&u2, i)
			fmt.Print("\n")
		}
		time.Sleep(1 * time.Second)
	}
}
