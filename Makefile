USE_LLM_PLANNER ?= false
RUN_FLAGS = --diff testdata/sample.patch --lang go --out report.json --mdout report.md

ifeq ($(USE_LLM_PLANNER),true)
RUN_FLAGS += --use-llm-planner
endif

run:
	go run ./cmd/cli $(RUN_FLAGS)

run-llm:
	$(MAKE) run USE_LLM_PLANNER=true

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...