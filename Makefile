all: uro

uro: _build
	go build -o _build/uro cmd/uro/*.go

test-process: _build
	go build -o _build/test-process cmd/test-process/*.go

_build:
	mkdir -p _build

clean:
	rm -rf _build