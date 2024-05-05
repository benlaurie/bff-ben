/*
<          Decrement head 0
>          Increment head 0
{		   Decrement head 1
}		   Increment head 1
+		   Increment cell at head 0
-		   Decrement cell at head 0
.		   Copy cell at head 0 to head 1
,		   Copy cell at head 1 to head 0
[		   Beginning of loop (error if no matching ]) - go to end if *head0 == 0
]		   End of loop (error if no matching [) - go to beginning if *head0 != 0
!		   Place head 0 here
?		   Place head 1 here

a          Add 2 to head 0
b          Add 4 to head 0
c          Add 8 to head 0
d          Add 16 to head 0
e          Add 32 to head 0
f          Add 64 to head 0
g          Add 128 to head 0

z		  Subtract 2 from head 0
y		  Subtract 4 from head 0
x		  Subtract 8 from head 0
w		  Subtract 16 from head 0
v		  Subtract 32 from head 0
u		  Subtract 64 from head 0
t		  Subtract 128 from head 0

A		  Add 2 to head 1
B		  Add 4 to head 1
C		  Add 8 to head 1
D		  Add 16 to head 1
E		  Add 32 to head 1
F		  Add 64 to head 1
G		  Add 128 to head 1

Z		  Subtract 2 from head 1
Y		  Subtract 4 from head 1
X		  Subtract 8 from head 1
W		  Subtract 16 from head 1
V		  Subtract 32 from head 1
U		  Subtract 64 from head 1
T		  Subtract 128 from head 1
*/

package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

/*
func graphics(universe *[65536]uint8) {
	// Sleep forever
	select {}
}
*/

// const OPS = "<>{}+-.,[]!?abcdefgtuvwxyzABCDEFZYXWVUT"
const OPS = "<>{}+-.,[]"
const SQRT_ULEN = 256
const ULEN = SQRT_ULEN * SQRT_ULEN
const SLEN = 1024
const ILIMIT = 5_000
const MUTATION_RATE = 10_000 // Higher is less mutation
const RUNNERS = 8
const STRICT = true
const SHOW_LEN = 8192

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
	iterations := 0
	head0 := pc
	head1 := pc + 12
	/*
	   copy := uint8(0)
	   copy_set := false
	*/
OUTER:
	for {
		if iterations++; iterations > ILIMIT {
			break
		}
		head0 = pmod(head0, ULEN)
		head1 = pmod(head1, ULEN)

		op := program[pc]
		switch op {
		case '<':
			head0 -= 1
		case '>':
			head0 += 1
		case '{':
			head1 -= 1
		case '}':
			head1 += 1
		case '+':
			program[head0]++
		case '-':
			program[head0]--
		case '.':
			program[head1] = program[head0]
			/*
				copy = program[head0]
				copy_set = true
			*/
		case ',':
			program[head0] = program[head1]
			/*
				if !copy_set {
					break OUTER
				}
				program[head1] = copy
				copy_set = false
			*/
		case '[':
			npc := pmod(pc+1, ULEN)
			count := 1
			for npc != pc {
				if program[npc] == '[' {
					count++
				} else if program[npc] == ']' {
					count--
				}
				if count == 0 {
					break
				}
				npc = pmod(npc+1, ULEN)
			}
			if npc == pc {
				break OUTER
			}
			if program[head0] != 0 {
				break
			}
			pc = pmod(npc+1, ULEN)
		case ']':
			npc := pmod(pc-1, ULEN)
			count := 1
			for npc != pc {
				if program[npc] == '[' {
					count--
				} else if program[npc] == ']' {
					count++
				}
				if count == 0 {
					break
				}
				npc = pmod(npc-1, ULEN)
			}
			if npc == pc {
				break OUTER
			}
			if program[head0] == 0 {
				break
			}
			pc = pmod(npc+1, ULEN)
			/*
				case '!':
					head0 = pc
				case '?':
					head1 = pc
			*/
			/*
				case 'a':
					head0 += 2
				case 'A':
					head1 += 2
				case 'z':
					head0 -= 2
				case 'Z':
					head1 -= 2
			*/
			/*
				default:
					switch {
					case op >= 'a' && op <= 'g':
						head0 += 1 << int(op-'a')
					case op >= 't' && op <= 'z':
						head0 -= 256 >> int(op-'t')
					case op >= 'A' && op <= 'G':
						head1 += 1 << int(op-'A')
					case op >= 'T' && op <= 'Z':
						head1 -= 256 >> int(op-'T')
					default:
						iterations--
					}
			*/
			//		default:
			//			iterations--
		}
		pc = (pc + 1) % ULEN
	}
	return iterations
}

func charp(op uint8) string {
	if !strings.Contains(OPS, string(op)) {
		return " "
	}
	return string(op)
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
	//program[rand.Intn(ULEN)] = uint8(OPS[rand.Intn(12)])
	program[rand.Intn(ULEN)] = uint8(OPS[rand.Intn(len(OPS))])
	//program[rand.Intn(ULEN)] = uint8(rand.Intn(256))
	//program[rand.Intn(ULEN)] = uint8(rand.Intn(MAX_OP + 1))
	/*
		switch rand.Intn(5) {
		case 0:
			program[rand.Intn(ULEN)] = uint8(rand.Intn(256))
		case 1:
			program[rand.Intn(ULEN)] = uint8(rand.Intn(MAX_OP + 1))
		default:
			program[rand.Intn(ULEN)] = 0x20 + uint8(rand.Intn(MAX_OP+1-0x20))
		}
	*/
}

func runner(universe *[ULEN]uint8, generation *uint64, n_ops *uint64) {
	t := 0
	for {
		n := run(universe, rand.Intn(ULEN))
		*n_ops += uint64(n)
		t += n
		for t > MUTATION_RATE {
			mutate(universe)
			t -= MUTATION_RATE
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
	f := fmt.Sprintf("logs/bfsoup.log.%s.%s", strict, time.Now().Format("2006-01-02-15:04:05"))
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
		//universe[i] = 0x3f
		//universe[i] = uint8(rand.Intn(MAX_OP + 1))
		//universe[i] = uint8(rand.Intn(256))
		universe[i] = 0
		//mutate(&universe)
	}

	var generation uint64
	var n_ops uint64
	for i := 0; i < RUNNERS; i++ {
		go runner(&universe, &generation, &n_ops)
	}

	go func() {
		p_n_ops := uint64(0)
		p_generation := uint64(0)
		for {
			var u2 [ULEN]uint8
			copy(u2[:], universe[:])
			dump(log, generation, n_ops, &u2)
			if generation == p_generation {
				p_generation--
			}
			fmt.Println("\033c", generation, n_ops, generation-p_generation, n_ops-p_n_ops, (n_ops-p_n_ops)/(generation-p_generation))
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

	/*
		n := x + y*w
		if n >= ULEN {
			return color.Black
		}

		op := universe[n]

		if op > MAX_OP {
			//return color.Black
			hsl, _ := colorconv.HSLToColor(0.0, 0.0, 0.0)
			return hsl
		}

		hsl, _ := colorconv.HSLToColor(float64(op)/256.0*360.0, 1.0, 0.5)

		return hsl
	*/
	//hsl, err := colorconv.HSLToColor(float64(op)/256.0*360.0, 1.0, 0.5)
	// raster := canvas.NewRasterFromImage()
	graphics(&universe)
}
