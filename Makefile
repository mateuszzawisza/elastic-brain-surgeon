test:
	 go test ./... -cover -short

test-full:
	 go test ./... -cover

build:
	gox -output "./bin/elastic-brain-surgeon_{{.OS}}_{{.Arch}}" -os "linux darwin" -arch "amd64 386"
