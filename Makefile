PYTHON ?= python3

.PHONY: test-cases-sync test-cases-stats
test-cases-sync:
	$(PYTHON) test/test-sync-cases.py

test-cases-stats:
	$(PYTHON) test/test-case-stats.py
