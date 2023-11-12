package main

import (
	"fmt"
	"math/rand"
)

/*
const data_size = 30000

	func execute_bf(program []uint8) {
		data := make([]int16, data_size)
		var data_ptr uint16 = 0
		reader := bufio.NewReader(os.Stdin)
		for pc := 0; pc < len(program); pc++ {
			switch program[pc] {
			case '>':
				data_ptr++
			case '<':
				data_ptr--
			case '+':
				data[data_ptr]++
			case '-':
				data[data_ptr]--
			case '.':
				fmt.Printf("%c", data[data_ptr])
			case ',':
				read_val, _ := reader.ReadByte()
				data[data_ptr] = int16(read_val)
			case '[':
				if data[data_ptr] == 0 {
					for pc < len(program) && program[pc] != ']' {
						pc++
					}
				}
			case ']':
				if data[data_ptr] > 0 {
					for pc >= 0 && program[pc] != '[' {
						pc--
					}
					if pc < 0 {
						break
					}
				}
			default:
				panic("Unknown operator")

			}
		}
	}
*/
const nprograms = 100
const plen = 32

func execute_rbf(program *[plen * 2]uint8) {
	data_ptr := 0
	var tmp uint8 = 0
	pl := len(program)
	iterations := 0
	for pc := 0; pc < pl && iterations < 1000; pc++ {
		switch program[pc] {
		case '>':
			data_ptr++
			if data_ptr >= pl {
				data_ptr = 0
			}
		case '<':
			data_ptr--
			if data_ptr < 0 {
				data_ptr = pl - 1
			}
		case '}':
			d := (data_ptr + plen) % pl
			program[d] = program[data_ptr]
			/*
				case '*':
					if data_ptr < 0 || data_ptr >= pl {
						break
					}
					data_ptr = int(program[data_ptr]) % pl
			*/
		case '+':
			if data_ptr < 0 || data_ptr >= pl {
				break
			}
			program[data_ptr]++
		case '-':
			if data_ptr < 0 || data_ptr >= pl {
				break
			}
			program[data_ptr]--
		case '.':
			if data_ptr < 0 || data_ptr >= pl {
				break
			}
			tmp = program[data_ptr]
		case ',':
			if data_ptr < 0 || data_ptr >= pl {
				break
			}
			program[data_ptr] = tmp
		case '[':
			if data_ptr < 0 || data_ptr >= pl {
				break
			}
			if program[data_ptr] == 0 {
				c := 0
				save := pc
				pc++
				for pc < pl && (program[pc] != ']' || c > 0) {
					if program[pc] == '[' {
						c++
					} else if program[pc] == ']' {
						c--
					}
					pc++
				}
				if pc >= pl {
					pc = save
					goto again
				}
			}
		case ']':
			if data_ptr < 0 || data_ptr >= pl {
				break
			}
			if program[data_ptr] > 0 {
				c := 0
				save := pc
				pc--
				for pc >= 0 && (program[pc] != '[' || c > 0) {
					if program[pc] == ']' {
						c++
					} else if program[pc] == '[' {
						c--
					}
					pc--
				}
				if pc < 0 {
					//goto done
					pc = save
					goto again
				}
			}
		}
	again:
		iterations++
	}
	// done:
}

/*
func main() {
	execute_bf(([]uint8)(">+++++++++[<++++++++>-]<.>+++++++[<++++>-]<+.+++++++..+++.[-]>++++++++[<++++>-]"))
	execute_rbf(([]uint8)(">+++++++++[<++++++++>-]<.>+++++++[<++++>-]<+.+++++++..+++.[-]>++++++++[<++++>-]"))
}
*/

func showp(programs [nprograms][plen]uint8) {
	for i := 0; i < nprograms; i++ {
		for j := 0; j < plen; j++ {
			c := programs[i][j]
			if c == '>' || c == '<' || c == '+' || c == '}' || c == '-' || c == '.' || c == ',' || c == '[' || c == ']' || c == '*' {
				fmt.Printf("%c", c)
			} else {
				fmt.Printf("\x1b[2m%d\x1b[0m", c/26)
				//fmt.Printf(" ")
			}
		}
		fmt.Printf("\n")
	}
}

const mutation_rate = 32000
const code_chars = "><+-.,[]*}"

func mutate(program *[plen]uint8, programs *[nprograms][plen]uint8) {
	r := rand.Intn(5)
	switch r {
	case 0:
		program[rand.Intn(plen)] = code_chars[rand.Intn(len(code_chars))]
	case 1:
		program[rand.Intn(plen)] = uint8(rand.Intn(256))
	case 2:
		n := rand.Intn(plen) + 1
		var tmp [plen]uint8
		copy(tmp[:], program[:n])
		copy(program[:], program[n:])
		copy(program[plen-n:], tmp[:])
	case 3:
		n := rand.Intn(plen)
		p := rand.Intn(nprograms)
		copy(program[n:], programs[p][n:])
	case 4:
		n := rand.Intn(plen)
		p := rand.Intn(nprograms)
		copy(program[:n], programs[p][:n])
	}
}

func fix(program *[plen]uint8) {
	for i := 0; i < plen; i++ {
		if program[i] == '[' {
			c := 0
			j := i + 1
			for ; j < plen; j++ {
				if program[j] == '[' {
					c++
				} else if program[j] == ']' {
					if c == 0 {
						break
					}
					c--
				}
			}
			if j >= plen {
				if i == plen-1 || rand.Intn(2) == 0 {
					program[i] = uint8(rand.Intn(256))
					i--
				} else {
					t := rand.Intn(plen - i - 1)
					program[i+t+1] = ']'
					i = t + i
				}
			}
		} else if program[i] == ']' {
			c := 0
			j := i - 1
			for ; j >= 0; j-- {
				if program[j] == ']' {
					c++
				} else if program[j] == '[' {
					if c == 0 {
						break
					}
					c--
				}
			}
			if j < 0 {
				if i == 0 || rand.Intn(2) == 0 {
					program[i] = uint8(rand.Intn(256))
					i--
				} else {
					t := rand.Intn(i)
					program[t] = '['
					i = t - 1
				}
			}
		}
	}
}

func main() {
	//rand.Seed(1)
	var programs [nprograms][plen]uint8
	for i := 0; i < nprograms; i++ {
		for j := 0; j < plen; j++ {
			programs[i][j] = uint8(rand.Intn(256))
		}
	}
	showp(programs)
	/*
		for {
			for i := 0; i < nprograms; i++ {
				execute_rbf(&programs[i])
			}
			showp(programs)
			fmt.Printf("********************\n")
		}
	*/
	n := 0
	m := 0
	for {
		m++
		if m > mutation_rate {
			mutate(&programs[rand.Intn(nprograms)], &programs)
			m = 0
		}
		p1 := rand.Intn(nprograms)
		p2 := rand.Intn(nprograms)

		var merged [plen * 2]uint8
		copy(merged[:plen], programs[p1][:plen])
		copy(merged[plen:], programs[p2][:plen])

		execute_rbf(&merged)

		copy(programs[p1][:plen], merged[:plen])
		copy(programs[p2][:plen], merged[plen:])

		//fix(&programs[p1])
		//fix(&programs[p2])

		if n++; n > 1000000 {
			fmt.Printf("\033c")
			showp(programs)
			//	fmt.Printf("********************\n")
			n = 0
		}
	}
}
