EXECUTABLE := "kubeletmein"
GITVERSION := $(shell git describe --dirty --always --tags --long)
PACKAGENAME := $(shell go list -m -f '{{.Path}}')

build: clean
	go build -ldflags "-extldflags '-static' -X ${PACKAGENAME}/pkg/config.GitVersion=${GITVERSION}" -o ${EXECUTABLE} ./cmd/kubeletmein

build-quick: clean
	go build -ldflags "-extldflags '-static' -X ${PACKAGENAME}/pkg/config.GitVersion=${GITVERSION}" -o ${EXECUTABLE} ./cmd/kubeletmein

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-extldflags '-static' -X ${PACKAGENAME}/pkg/config.GitVersion=${GITVERSION}" -o ${EXECUTABLE}-linux ./cmd/kubeletmein

clean:
	@rm -f ${EXECUTABLE}

test:
	go test -v ./...

docker:
	docker build -f build/Dockerfile . -t ${EXECUTABLE}
