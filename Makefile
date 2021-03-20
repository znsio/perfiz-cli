clean:
	rm -rf artifacts

build:
	go build -o bin/main main.go

run:
	go run main.go

compile: clean
	echo "Running Cross Compiler"
	goxc