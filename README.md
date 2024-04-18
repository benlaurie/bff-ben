# BFF

Actually, mostly these are Forth experiments.

Most of them need large terminal windows.

## bf

My version of Blaise's original idea. I don't have read and write heads (not for any great reason other than I forgot that's how Blaise's worked) - instead there's a copy operator. I wouldn't say this was a great plan but you do get replicators with far fewer tapes.

```shell
$ go run links.org/bf/cmd/bf
```

## f1

Like all the rest of the experiments, this is based on Forth and uses a circular universe in which all programs run, rather than tapes that get combined randomly. This does not have copy, nor read/write heads, but it has load and store. IIRC, this version did not lead to replicators.

```shell
$ go run links.org/bf/cmd/f1
```

## f2

Adds a copy operator to f1. I think this did make replicators. Adds parallelism

```shell
$ GOMAXPROCS=32 go run links.org/bf/cmd/f2
```

## f3

Reduces the language to just instructions I judged were "good for replicators". This one does produce interesting behaviour.

This also introduced the notion of "strict" execution: in strict mode, programs halt if they try something illegal, such as adding 2 numbers when there's only 1 on the stack. Lenient just skips such oporations.

IMO, strict mode is the more interesting mode.

```shell
$ GOMAXPROCS=32 go run links.org/bf/cmd/f3
```

`f3.py` will produce a disassembly.

## f4

This was an attempt to introduce the notion of finite resources. Incomplete experiment.

```shell
$ GOMAXPROCS=32 go run links.org/bf/cmd/f4
```

## f5

Increasing language complexity, also bringing back load and store, Again this produces interesting behaviour. If you remove the copy operator (easily done by just commenting out the code for it) replicators have not arisen in my experiments.

```shell
$ GOMAXPROCS=32 go run --tags="graphics" links.org/bf/cmd/f5
```

`f5.py` will produce a CSV of iteration statistics.

## f6

Adds read/write heads, removes copy. Once more, interesting behaviour is observed.

```shell
$ GOMAXPROCS=32 go run links.org/bf/cmd/f5
```

Logs are compatible with f5 so you can use `f5.py`.
