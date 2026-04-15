run:
	go run ./cmd/cli --diff testdata/sample.patch --lang go --out report.json --mdout report.md

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...