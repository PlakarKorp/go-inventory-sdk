GO = go
EXT=

all: build

build:
	${GO} build -v -o test-inventory${EXT} ./cmd/test

clean:
	rm -f test-inventory
