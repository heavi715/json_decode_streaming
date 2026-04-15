PYTHON ?= python3
NODE_BIN ?= /Users/heavi/.nvm/versions/node/v22.14.0/bin

.PHONY: \
	test-cases-sync \
	test-cases-stats \
	package-python \
	package-javascript \
	package-php \
	package-go \
	package-all \
	release-sync-version \
	release-check \
	release-build \
	release-setup-secrets \
	release-publish-python \
	release-publish-javascript \
	release-publish-php \
	release-publish-go \
	release-publish-all

test-cases-sync:
	$(PYTHON) test/test-sync-cases.py

test-cases-stats:
	$(PYTHON) test/test-case-stats.py

package-python:
	cd python && uv build

package-javascript:
	export PATH="$(NODE_BIN):$$PATH" && cd javascript && npm pack

package-php:
	cd php && composer validate --strict && composer archive --format=zip --dir=.

package-go:
	go list ./... >/dev/null

package-all: package-python package-javascript package-php package-go

release-sync-version:
	bash scripts/sync-version.sh

release-check:
	bash scripts/release-check.sh

release-build:
	bash scripts/release-build.sh

release-setup-secrets:
	bash scripts/setup-release-secrets.sh

release-publish-python:
	bash scripts/release-publish.sh python

release-publish-javascript:
	export PATH="$(NODE_BIN):$$PATH" && bash scripts/release-publish.sh javascript

release-publish-php:
	bash scripts/release-publish.sh php

release-publish-go:
	bash scripts/release-publish.sh go

release-publish-all:
	export PATH="$(NODE_BIN):$$PATH" && bash scripts/release-publish.sh all
