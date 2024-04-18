import binascii
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

#print(f'ULEN: {ULEN} SLEN: {SLEN} ILIMIT: {ILIMIT} MUTATION_RATE: {MUTATION_RATE} RUNNERS: {RUNNERS}')

def find_longest_repoeated_substring(s):
    n = len(s)
    z = [0] * n
    l = r = 0
    longest_substring = bytes()
    for i in range(1, n):
        if i <= r:
            z[i] = min(r-i+1, z[i-l])
        while i+z[i] < n and s[z[i]] == s[i+z[i]]:
            z[i] += 1
        if i+z[i]-1 > r:
            l, r = i, i+z[i]-1
    max_z = max(z)
    if max_z > 1:
        longest_substring = s[l:l+max_z]
    print(binascii.hexlify(longest_substring), end=' ')
    return longest_substring
    if len(longest_substring) > 1:
        find_longest_repoeated_substring(longest_substring)

def test_find_longest_repoeated_substring():
    assert find_longest_repoeated_substring(b'abc') == b''
    assert find_longest_repoeated_substring(b'abcaa') == b'a'
    assert find_longest_repoeated_substring(b'abcab') == b'ab'
    assert find_longest_repoeated_substring(b'abcabc') == b'abc'
    assert find_longest_repoeated_substring(b'abcabcd') == b''

test_find_longest_repoeated_substring()
sys.exit(0)

def sign_extend(a):
	if a&0x08 == 0x08:
		return a - 0x10
	return a

pg = 0
po = 0
while True:
    generation = read_long(f)
    op_count = read_long(f)
    program = f.read(ULEN)
    if generation == 0:
         continue
    find_longest_repoeated_substring(program)
    print()
    continue

    print(f'{generation}, {op_count}, {generation-pg}, {(op_count-po) / (generation-pg)}')
    pg = generation
    po = op_count

    continue

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
