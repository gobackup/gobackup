test:
	GO_ENV=test go test ./...
web:
	cd web; yarn && yarn build
perform:
	@go run main.go -- perform -m demo -c ./gobackup_test.yml
run: web
	GIN_MODE=debug go run main.go -- run --config ./gobackup_test.yml
start: web
	GIN_MODE=debug go start main.go -- run --config ./gobackup_test.yml
build: web
	go build -o dist/gobackup
dev:
	cd web && yarn dev