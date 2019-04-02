test:
	GO_ENV=test go test ./...
run:
	go run main.go -- perform -m demo
release:
	@rm -Rf dist/
	@goreleaser --skip-validate
