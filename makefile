release: 
	mkdir tmp
	mkdir tmp/windows
	mkdir tmp/linux
	mkdir tmp/macos
	go build .
	mv kcli tmp/macos
	GOOS=linux GOARCH=amd64 go build .
	mv kcli tmp/linux
	GOOS=windows GOARCH=amd64 go build .
	mv kcli.exe tmp/windows
	mv tmp kcli
	zip -r kcli.zip kcli
	rm -rf kcli
