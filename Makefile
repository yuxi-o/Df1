
CC=cc

all: df1d df1c

df1d:
	${CC} src/*.c -o df1d  --std=gnu99
	strip $@

df1c:
	${CC} client/*.c -o df1c --std=gnu99
	strip $@

install: df1d df1c
	cp df1d df1c /usr/bin

uninstall:
	rm -f /usr/bin/df1d
	rm -f /usr/bin/df1c

clean:
	-rm -f df1d df1c
