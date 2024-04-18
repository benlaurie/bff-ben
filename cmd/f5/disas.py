import sys

for c in sys.argv[1]:
    print(c, end=' ')
    if c >= 'A' and c <= 'P':
        t = ord(c) - ord("A")
        if t > 7:
            t = t - 16
        print(f'PUSH {t}')
    elif c >= 'a' and c <= 'p':
        t = ord(c) - ord("a")
        if t > 7:
            t = t - 16
        print(f'SHIFT {t}')
    elif c == '=':
        print('COPY')
    elif c == '>':
        print('INC')
    elif c == '<':
        print('DEC')
    elif c == '^':
        print('JNZ')
    else:
        print('NOP')
