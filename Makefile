NAME = dockermon
RM = rm -f

all: $(NAME)

$(NAME):
	go build

fmt:
	go fmt ./...

clean:
	$(RM) $(NAME)
	$(RM) ./internal/config/testdata

gen:
	go run ./cmd/gen/configVersion.go
	go fmt ./internal/config/version.go

vet:
	go vet ./...

test:
	go test ./...

fuzz:
	go test -fuzz=FuzzParseConfig -fuzztime=5m ./internal/config

re: clean all

.PHONY: all clean re fmt