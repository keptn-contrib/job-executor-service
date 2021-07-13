compile:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=amd64 go build -o ../bin/job-lint-linux-amd64 .
	GOOS=windows GOARCH=amd64 go build -o ../bin/job-lint-windows-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o ../bin/job-lint-darwin-amd64 .
