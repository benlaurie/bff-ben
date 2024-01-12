import sys
import zlib

import matplotlib.pyplot as plt
import pandas
import seaborn

def read_long(s):
    t = s.read(8)
    if len(t) != 8:
        raise EOFError()
    return int.from_bytes(t, byteorder='little')

n = 0
data = []
d2 = []
for fn in sys.argv[1:]:
    f = open(fn, 'rb')
    
    ULEN = read_long(f)
    SLEN = read_long(f)
    ILIMIT = read_long(f)
    MUTATION_RATE = read_long(f)
    RUNNERS = read_long(f)

    label = f'{n} {MUTATION_RATE}'

    settled = False
    while True:
        try:
            generation = read_long(f)
        except EOFError:
            break
        program = f.read(ULEN)

        pc = zlib.compress(program, level=9)


        l = len(pc)
        print(generation, l)
        data.append((label, generation, l))

        if not settled and l < 1000 and generation > 1000:
            settled = True
            d2.append((MUTATION_RATE, generation))

    n += 1

data = pandas.DataFrame(data=data, columns=('file', 'generation', 'size'))
print(data)
#seaborn.lineplot(data=data, x='generation', y='size', hue='file')
#plt.show()

d2 = pandas.DataFrame(data=d2, columns=('mutation rate', 'generation'))
print(d2)
seaborn.relplot(data=d2, x='generation', y='mutation rate')
plt.show()