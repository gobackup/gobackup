test:
	go run main.go model.go -- perform -m demo
release:
	mkdir -p ./build && rm -f ./build/*
	@go build -ldflags "-s -w" -o build/gobackup ./*.go && cd build/ && zip gobackup-darwin-amd64.zip gobackup && rm gobackup && cd ..
	shasum -a 256 build/gobackup-darwin-amd64.zip
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o build/gobackup ./*.go && cd build/ && tar zcf gobackup-linux-amd64.tar.gz gobackup && rm gobackup && cd ..