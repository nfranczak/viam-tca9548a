viam-tca9548a: *.go */*.go
	# the executable
	go build -o $@ -ldflags "-s -w" -tags osusergo,netgo
	file $@

module.tar.gz: viam-tca9548a
	# the bundled module
	rm -f $@
	tar czf $@ $^
