TARGET=bin/pocket-pick
SRC=$(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "*_test.go")
BUILD_FLAGS?=-v

.PHONY: clean test dep tidy

all: build
build: $(TARGET)

$(TARGET): $(SRC)
	@mkdir -p bin
	go build -o bin/ ${BUILD_FLAGS} ./cmd/...

clean:
	rm -f ${TARGET}

test:
	go test

# update modules & tidy
dep:
	rm -f go.mod go.sum
	go mod init github.com/whitekid/pocket-pick
	@$(MAKE) tidy

tidy:
	go mod tidy

swag:
	swag init -d pkg  -g app.go
