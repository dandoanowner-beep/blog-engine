.PHONY: check test vet coverage lint

## check: Run all quality gates (vet + handler linter + coverage). Required before every PR.
check: vet lint coverage
	@echo ""
	@echo "All checks passed."

## test: Run the full test suite.
test:
	go test ./... -count=1

## vet: Run go vet (catches interface compliance drift via var _ checks).
vet:
	go vet ./...

## lint: Run the handler test assertion linter.
##       Fails if any handler test only checks HTTP status code without asserting body fields.
lint:
	@bash scripts/check-handler-tests.sh

## coverage: Run the business-logic coverage gate (>=80% required).
##           Excludes SQL/cloud repository adapters — those require integration tests.
coverage:
	@bash scripts/check-coverage.sh
