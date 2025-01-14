build:
	go build -o bin/main .

build-exe:
	go build -o bin/main.exe .

dev:
	make build
	./bin/main -debug $(ARGUMENTS)

run:
	make build
	./bin/main $(ARGUMENTS)