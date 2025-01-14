# How to run

Need go 1.23

## For linux

Run

```bash
CGO_ENABLED=1 go build -ldflags '-s -w' . # the first build is long since it compiles some C libraries
./optimview
```

## For wsl

```bash
devbox run build
devbox run run
```
