VERSION=`git rev-parse --short HEAD`
flags=-ldflags="-s -w -X main.version=${VERSION}"

all: build

vet:
	go vet .

build:
	go clean; rm -rf pkg register*; go build ${flags}

build_all: build build_osx build_linux build_power8 build_arm64

build_osx:
	go clean; rm -rf pkg register_osx; GOOS=darwin go build ${flags}
	mv register register_osx

build_linux:
	go clean; rm -rf pkg register_linux; GOOS=linux go build ${flags}
	mv register register_linux

build_power8:
	go clean; rm -rf pkg register_power8; GOARCH=ppc64le GOOS=linux go build ${flags}
	mv register register_power8

build_arm64:
	go clean; rm -rf pkg register_arm64; GOARCH=arm64 GOOS=linux go build ${flags}
	mv register register_arm64

install:
	go install

clean:
	go clean; rm -rf pkg

test : test1

test1:
	go test -v .
bench:
	go test -bench=.
