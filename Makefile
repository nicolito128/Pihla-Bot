build:
	go build -o bin/ main.go

build-exe:
	go build -o bin/main.exe main.go

dev:
	go build -o bin/ main.go
	./bin/main -debug $(ARGUMENTS)
