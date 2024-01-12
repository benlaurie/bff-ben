import sys

def read_long(s):
    b = s.read(8)
    assert len(b) == 8
    return int.from_bytes(b, byteorder='little')

f = open(sys.argv[1], 'rb')

ULEN = read_long(f)
SLEN = read_long(f)
ILIMIT = read_long(f)
MUTATION_RATE = read_long(f)
RUNNERS = read_long(f)

def sign_extend(a):
	if a&0x08 == 0x08:
		return a - 0x10
	return a

while True:
    generation = read_long(f)
    program = f.read(ULEN)
    print(generation)
    if generation < int(sys.argv[2]):
        continue

    PUSH       = 0x00
    SHIFT_PUSH = 0x10
    COPY       = 0x20
    INC        = 0x21
    DEC        = 0x22
    JNZ        = 0x23
    MAX_OP     = JNZ

    for op in program:
        if op&0xf0 == PUSH:
            print("PUSH", sign_extend(op&0x0f))
        elif op&0xf0 == SHIFT_PUSH:
            print("SHIFT_PUSH", op&0x0f)
        elif op == COPY:
            print("COPY")
        elif op == INC:
            print("INC")
        elif op == DEC:
            print("DEC")
        elif op == JNZ:
            print("JNZ")
        else:
            print("NOP")
