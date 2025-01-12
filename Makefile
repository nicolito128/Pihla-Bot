dev:
	go build -o bin/ main.go
	./bin/main -debug -name $(name) -pass $(pass) -rooms $(rooms)