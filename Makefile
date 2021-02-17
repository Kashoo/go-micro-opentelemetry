
.PHONY: help
## Display this help text
help:  # Based on https://gist.github.com/rcmachado/af3db315e31383502660
	$(info Available Targets)
	@awk '/^[a-zA-Z\-\_0-9]+:/ {                    \
		nb = sub( /^## /, "", helpMsg );              \
		if(nb == 0) {                                 \
		helpMsg = $$0;                              \
		nb = sub( /^[^:]*:.* ## /, "", helpMsg );   \
		}                                             \
		if (nb)                                       \
		printf "\033[1;31m%-" width "s\033[0m %s\n", $$1, helpMsg;   \
	}                                               \
	{ helpMsg = $$0 }'                              \
	$(MAKEFILE_LIST) | column -ts:


## Format & go-lint
lint: format check


###############################################################################
# Lint
###############################################################################

format: ## Format go code with goimports
	@go get golang.org/x/tools/cmd/goimports
	@goimports -l -w .

format-check: ## Check if the code is formatted
	@go get golang.org/x/tools/cmd/goimports
	@for i in $$(goimports -l .); do echo "[ERROR] Code is not formated run 'make format'" && exit 1; done

check: format-check ## Linting and static analysis
	@if grep -r --include='*.go' -E "fmt.Print|spew.Dump" *; then \
		echo "code contains fmt.Print* or spew.Dump function"; \
		exit 1; \
	fi

	@if test ! -e ./bin/golangci-lint; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh; \
	fi
	@./bin/golangci-lint run --timeout 180s -E gosec -E stylecheck -E golint -E goimports
