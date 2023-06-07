test:
	GO_ENV=test go test ./...
test\:all:
	@sh tests/test.sh
build_web:
	cd web; yarn && yarn build
perform:
	@go run main.go -- perform -m demo -c ./gobackup_test.yml
run:
	GO_ENV=dev go run main.go -- run --config ./gobackup_test.yml
start:
	GO_ENV=dev go run main.go -- start --config ./gobackup_test.yml
build: build_web
	go build -o dist/gobackup
dev:
	cd web && yarn dev
