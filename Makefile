EXECUTABLE := "kubeletmein"

build: clean test
	go build -ldflags "-extldflags '-static'" -o ${EXECUTABLE} ./cmd/kubeletmein

clean:
	@rm -f ${EXECUTABLE}

test:
	go test -v ./pkg/bootstrap