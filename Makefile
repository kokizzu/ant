
pkg  := .
func := .

deps:
	@go get -d -t ./...

test:
	@go test --cover --timeout 1s --race ./...

bench:
	@go test \
		--run=Bench \
		--bench=$(func) \
		--benchmem \
		--blockprofile=block.prof \
		--cpuprofile=cpu.prof \
		--memprofile=mem.prof \
		$(pkg)

prof.%:
	@go tool pprof --http :8080 ant.test $*.prof

clean:
	rm -fr *.test *.prof
