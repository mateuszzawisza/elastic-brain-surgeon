test:
	 go test ./... -cover -short

test-full:
	 go test ./... -cover

build:
	@GOOS=linux go build -o bin/brain
