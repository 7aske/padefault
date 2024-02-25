PREFIX = /usr/local
CMD = padefault

default: cmd/$(CMD)/main.go
	go build -o bin/$(CMD) cmd/$(CMD)/main.go
	chmod +x bin/$(CMD)

install: default
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	cp bin/$(CMD) $(PREFIX)/bin/$(CMD)

clean:
	rm -f bin/$(CMD)