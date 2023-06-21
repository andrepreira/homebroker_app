.PHONY: deps tests 

deps:
	cd stockexchange_simulator/go && go mod tidy

tests:
	cd stockexchange_simulator/go && go test -v ./...
