.PHONY: build

build:
	sam build

run-local: build
	sam local invoke GenerateFunction
