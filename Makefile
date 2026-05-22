GO:=go

.PHONY: run

all: run

run:
	${GO} build -v -o run ./cmd/run

clean:
	rm run
