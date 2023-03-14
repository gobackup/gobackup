test:
	GO_ENV=test go test ./...
run:
	@go run main.go -- perform -m demo -c ./gobackup_test.yml
start:
	cd web; yarn && yarn build
	GIN_MODE=debug go run main.go -- run --config ./gobackup_test.yml
release:
	@rm -Rf dist/
	@goreleaser --skip-validate