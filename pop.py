import sys
import seaborn
import pandas
import matplotlib.pyplot as plt

pop = pandas.read_csv(sys.argv[1], header=None)
#pop.rename(columns={0: 'pop', 1: 'not', 2: 'fitness', 3: 'age', 4: 'size', 5: 'depth', 6: 'nodes', 7: 'hits', 8: 'time', 9: 'evals', 10: 'nodes', 11: 'hits'}, inplace=True)
print("Read data", len(pop), pop.columns)
for n in range(12):
    seaborn.lineplot(x=range(len(pop)), y=pop[n])
#pnop = pop[12:128].sum()
#seaborn.lineplot(x=range(len(pnop)), y=pnop)
ppush = pop[128:256].sum()
seaborn.lineplot(x=range(len(ppush)), y=ppush)

plt.legend(labels=['pop', 'not', 'add', 'mul', 'store', 'dup', 'jump', 'jnz', 'load', 'swap', 'copy', 'call', 'push'])

plt.show()