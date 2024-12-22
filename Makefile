NAME = dockermon
RM = rm -f

all: $(NAME)

$(NAME):
	go build ./cmd/dockermon

fmt:
	go fmt ./...

clean:
	$(RM) $(NAME)

re: clean all

.PHONY: all clean re fmt