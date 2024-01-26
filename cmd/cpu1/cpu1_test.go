package main

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestPush(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 0x0f
	cpu.instructions[0].code[1] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 2)
	assert.Equal(t, cpu.sp, 1)
	assert.Equal(t, cpu.stack[0], -1)

	cpu.instructions[0].code[0] = SHIFT_PUSH | 0x0e
	r, c = cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 2)
	assert.Equal(t, cpu.sp, 1)
	assert.Equal(t, cpu.stack[0], -2)
}

func TestDup(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 0x0f
	cpu.instructions[0].code[1] = DUP
	cpu.instructions[0].code[2] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 3)
	assert.Equal(t, cpu.sp, 2)
	assert.Equal(t, cpu.stack[0], -1)
	assert.Equal(t, cpu.stack[1], -1)
}

func TestPull0(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 1
	cpu.instructions[0].code[1] = PUSH | 0
	cpu.instructions[0].code[2] = PULL
	cpu.instructions[0].code[3] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 4)
	assert.Equal(t, cpu.sp, 1)
	assert.Equal(t, cpu.stack[0], 1)
}

func TestPull0f(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 0
	cpu.instructions[0].code[1] = PULL
	cpu.instructions[0].code[2] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, !r)
	assert.Equal(t, c, 2)
}

func TestPull1(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 2
	cpu.instructions[0].code[1] = PUSH | 3
	cpu.instructions[0].code[2] = PUSH | 1
	cpu.instructions[0].code[3] = PULL
	cpu.instructions[0].code[4] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 5)
	assert.Equal(t, cpu.sp, 2)
	assert.Equal(t, cpu.stack[0], 3)
	assert.Equal(t, cpu.stack[1], 2)
}

func TestPull1f(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 0
	cpu.instructions[0].code[1] = PUSH | 1
	cpu.instructions[0].code[2] = PULL
	cpu.instructions[0].code[3] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, !r)
	assert.Equal(t, c, 3)
}

func TestPull2(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 1
	cpu.instructions[0].code[1] = PUSH | 3
	cpu.instructions[0].code[2] = PUSH | 4
	cpu.instructions[0].code[3] = PUSH | 2
	cpu.instructions[0].code[4] = PULL
	cpu.instructions[0].code[5] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 6)
	assert.Equal(t, cpu.sp, 3)
	assert.Equal(t, cpu.stack[0], 3)
	assert.Equal(t, cpu.stack[1], 4)
	assert.Equal(t, cpu.stack[2], 1)
}

func TestUnpull1(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 2
	cpu.instructions[0].code[1] = PUSH | 3
	cpu.instructions[0].code[2] = PUSH | 0xf // -1
	cpu.instructions[0].code[3] = PULL
	cpu.instructions[0].code[4] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 5)
	assert.Equal(t, cpu.sp, 2)
	assert.Equal(t, cpu.stack[0], 3)
	assert.Equal(t, cpu.stack[1], 2)
}

func TestUnpull1f(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 0
	cpu.instructions[0].code[1] = PUSH | 0xf // -1
	cpu.instructions[0].code[2] = PULL
	cpu.instructions[0].code[3] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, !r)
	assert.Equal(t, c, 3)
}

func TestUnpull2(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].code[0] = PUSH | 1
	cpu.instructions[0].code[1] = PUSH | 2
	cpu.instructions[0].code[2] = PUSH | 3
	cpu.instructions[0].code[3] = PUSH | 0xe // -2
	cpu.instructions[0].code[4] = PULL
	cpu.instructions[0].code[5] = HALT
	r, c := cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 6)
	assert.Equal(t, cpu.sp, 3)
	assert.Equal(t, cpu.stack[0], 3)
	assert.Equal(t, cpu.stack[1], 1)
	assert.Equal(t, cpu.stack[2], 2)
}

func TestCopy(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].init_copy()
	cpu.stack[0] = 1
	cpu.stack[1] = 2
	cpu.sp = 2
	cpu.memory[1] = 3

	r, c := cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 12)
	assert.Equal(t, cpu.sp, 1)
	assert.Equal(t, cpu.stack[0], 1)
	assert.Equal(t, cpu.memory[1], uint8(3))
	assert.Equal(t, cpu.memory[3], uint8(3))
}

func TestJNZ(t *testing.T) {
	var cpu cpu

	cpu.instructions[0].init_jnz()
	cpu.stack[0] = 1
	cpu.stack[1] = 23
	cpu.sp = 2
	cpu.pc = 5

	r, c := cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 6)
	assert.Equal(t, cpu.sp, 0)
	assert.Equal(t, cpu.pc, 28)

	cpu.stack[0] = 0
	cpu.stack[1] = 23
	cpu.sp = 2

	r, c = cpu.run_instr(0)
	assert.Assert(t, r)
	assert.Equal(t, c, 6)
	assert.Equal(t, cpu.sp, 0)
	assert.Equal(t, cpu.pc, 29)
}
