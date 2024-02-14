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
	"image"
	"image/color"
	"math/rand"
	"os"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"github.com/crazy3lf/colorconv"
)

const SQRT_ULEN = 256
const ULEN = SQRT_ULEN * SQRT_ULEN
const SLEN = 1024
const ILIMIT = 1_000
const MUTATION_RATE = 500_000 // Higher is less mutation
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
				if sp > 0 {
					n := int(stack[sp-1])
					sp--
					if n > sp {
						if STRICT {
							break OUTER
						}
					} else {
						if n > 0 {
							t := stack[sp-1]
							for i := 0; i > n-1; i-- {
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
	//program[rand.Intn(ULEN)] = uint8(rand.Intn(256))
	program[rand.Intn(ULEN)] = uint8(rand.Intn(MAX_OP + 1))
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
		//universe[i] = uint8(rand.Intn(256))
		//mutate(&universe)
	}

	var generation uint64
	var n_ops uint64
	for i := 0; i < RUNNERS; i++ {
		go runner(&universe, &generation, &n_ops)
	}

	myApp := app.New()
	w := myApp.NewWindow("Raster")

	raster := canvas.NewRasterWithPixels(
		func(x, y, w, h int) color.Color {
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

			x = x * SQRT_ULEN / w
			y = y * SQRT_ULEN / h
			n := x + y*SQRT_ULEN
			if n >= ULEN {
				hsl, _ := colorconv.HSLToColor(0.0, 0.0, 0.0)
				return hsl
			}

			op := universe[n]

			if op > MAX_OP {
				hsl, _ := colorconv.HSLToColor(0.0, 0.0, 0.0)
				return hsl
			}

			hsl, err := colorconv.HSLToColor(float64(op)/float64(MAX_OP+1)*360.0, 0.9, 0.5)
			//hsl, err := colorconv.HSLToColor(float64(op)/256.0*360.0, 1.0, 0.5)

			if err != nil {
				panic(err)
			}

			return hsl

		})
	// raster := canvas.NewRasterFromImage()
	w.SetContent(raster)
	w.Resize(fyne.NewSize(SQRT_ULEN*2, SQRT_ULEN*2))

	i_w := myApp.NewWindow("Instructions")
	i_raster := canvas.NewRaster(
		func(w, h int) image.Image {
			var ops [256]uint64
			max := uint64(0)
			for i := 0; i < ULEN; i++ {
				ops[universe[i]]++
				if ops[universe[i]] > max {
					max = ops[universe[i]]
				}
			}
			image := image.NewRGBA(image.Rect(0, 0, w, h))
			for y := 0; y < h; y++ {
				op := y * 256 / h
				if op > 255 {
					op = 255
				}
				hue := float64(op) / float64(MAX_OP+1) * 360.0
				l := float64(ops[op]) / float64(max)
				s := 1.0
				if op > MAX_OP {
					hue = 0.0
					s = 0.0
				}

				hsl, err := colorconv.HSLToColor(hue, s, l)
				if err != nil {
					panic(err)
				}

				for x := 0; x < w; x++ {
					image.Set(x, y, hsl)
				}
			}
			return image
		})
	i_w.SetContent(i_raster)
	i_w.Resize(fyne.NewSize(128, 512))

	go func() {
		for {
			raster.Refresh()
			i_raster.Refresh()
			time.Sleep(20 * time.Millisecond)
		}
	}()

	go func() {
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
	}()

	i_w.Show()
	w.ShowAndRun()

}
