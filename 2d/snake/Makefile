.DEFAULT_GOAL := build

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
GODOT?=$(shell which godot)
CWD=$(shell pwd)

OUTPUT_PATH=project/lib
TEST_MAIN=main.go

ifeq ($(GOOS),windows)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-2dsnake-windows-$(GOARCH).dll
else ifeq ($(GOOS),darwin)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-2dsnake-macos-$(GOARCH).framework
else ifeq ($(GOOS),linux)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-2dsnake-linux-$(GOARCH).so
else
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-2dsnake-$(GOOS)-$(GOARCH).so
endif

.PHONY: goenv build clean run editor

goenv:
	go env

build: goenv
	CGO_ENABLED=1 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_CFLAGS='-Og -g3 -g -DX86=1 -fPIC' \
	CGO_LDFLAGS='-Og -g3 -g' \
	go build -gcflags=all="-N -l" -tags tools -buildmode=c-shared -trimpath -o "$(TEST_BINARY_PATH)" $(TEST_MAIN)

clean:
	rm -f project/lib/libgodotgo-*

ci_gen_project_files:
	CI=1 \
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=sbrk=1,gctrace=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	$(GODOT) --headless --verbose --path project/ --editor --quit

run:
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=sbrk=1,gctrace=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	$(GODOT) --debug --path project/

editor:
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=sbrk=1,gctrace=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	$(GODOT) --verbose --debug --path project/ --editor
