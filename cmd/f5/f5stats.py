#import binascii
import brotli
import pandas as pd
import plotly.express as px
import plotly.graph_objects as go
import sys
import math

from plotly.subplots import make_subplots

def read_long(s):
    b = s.read(8)
    if len(b) < 8:
        return 0, False
    return int.from_bytes(b, byteorder='little'), True

f = open(sys.argv[1], 'rb')

ULEN, _ = read_long(f)
SLEN, _ = read_long(f)
ILIMIT, _ = read_long(f)
MUTATION_RATE, _ = read_long(f)
RUNNERS, _ = read_long(f)

#print(f'ULEN: {ULEN} SLEN: {SLEN} ILIMIT: {ILIMIT} MUTATION_RATE: {MUTATION_RATE} RUNNERS: {RUNNERS}')

def sign_extend(a):
	if a&0x08 == 0x08:
		return a - 0x10
	return a

def entropy(data):
    entropy = 0
    for byte in range(256):
        p_x = float(data.count(byte))/len(data)
        if p_x > 0:
            entropy += - p_x * math.log(p_x, 2)
    return entropy

def changes(a, b):
    c = 0
    for i in range(len(a)):
        if a[i] != b[i]:
            c += 1
    return c

pg = 0
po = 0
x = []
y = []
y2 = []
y3 = []
y4 = []
y5 = []
prev_program = None
print("generation, op_count, delta_gen, rate, cratio, entropy, entropy-cratio, changes")
while True:
    generation, ok = read_long(f)
    if not ok:
        break
    op_count, _ = read_long(f)
    program = f.read(ULEN)
    
    rate = (op_count-po) / (generation-pg)
    compressaed = brotli.compress(program)
    cratio = len(compressaed) / len(program) * 8
    e = entropy(program)
    change = changes(program, prev_program) if prev_program is not None else None
    print(f'{generation}, {op_count}, {generation-pg}, {rate}, {cratio}, {e}, {e-cratio}, {change}')
    pg = generation
    po = op_count

    x.append(generation)
    y.append(rate)
    y2.append(cratio)
    y3.append(e)
    y4.append(e - cratio)
    y5.append(change)
    prev_program = program

    #if generation > 1000000000:
    #    break

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

print(len(x), len(y))
df = pd.DataFrame({'generation': x, 'rate': y, 'cratio': y2, 'entropy': y3, 'change': y5})
#fig = px.line(df, x='generation', y='rate')
fig = make_subplots(specs=[[{"secondary_y": True}]])
fig.add_trace(go.Scatter(x=df['generation'], y=df['rate'], mode='lines', name='rate'), secondary_y=False)
#fig.add_trace(go.Scatter(x=df['generation'], y=df['cratio'], mode='lines', name='cratio'), secondary_y=True)
#fig.add_trace(go.Scatter(x=df['generation'], y=df['entropy'], mode='lines', name='entropy'), secondary_y=True)
fig.add_trace(go.Scatter(x=df['generation'], y=df['entropy']-df['cratio'], mode='lines', name='entropy-cratio', yaxis='y2'), secondary_y=True)
fig.add_trace(go.Scatter(x=df['generation'], y=df['change'], mode='lines', name='change', yaxis='y3'), secondary_y=False)
fig.show()