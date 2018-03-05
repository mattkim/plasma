compile:
	go install ./...

start: compile
	${GOPATH}/bin/plasma start

clean:
	rm -rf ~/.plasma

fresh-start: clean start