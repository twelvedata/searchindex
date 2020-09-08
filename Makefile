PATH_THIS:=$(realpath $(dir $(lastword ${MAKEFILE_LIST})))
DIR:=$(PATH_THIS)


help:
	@echo "    test"
	@echo "        Run tests"


.PHONY: test
test:
	@cd $(DIR) \
	&& go test ./...
