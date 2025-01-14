build:
	go build -o bin/ main.go

build-exe:
	go build -o bin/main.exe main.go

dev:
	make build
	./bin/main -debug $(ARGUMENTS)

run:
	make build
	./bin/main $(ARGUMENTS)