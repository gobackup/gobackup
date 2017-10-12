test:
	GO_ENV=test go test ./...
run:
	go run main.go model.go -- perform -m demo
release:
	@rm -Rf dist/
	@goreleaser --skip-validate