EXECUTABLE := "kubeletmein"

build: clean test
	go build -ldflags "-extldflags '-static'" -o ${EXECUTABLE} ./cmd/kubeletmein

clean:
	@rm -f ${EXECUTABLE}

test:
	go test -v ./pkg/do
	go test -v ./pkg/gke

docker:
	docker build -f build/Dockerfile . -t ${EXECUTABLE}