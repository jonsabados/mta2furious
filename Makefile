.PHONY: mocks
mocks:
	@ MOCKERY_CASE=snake MOCKERY_WITH_EXPECTER=true MOCKERY_EXPORTED=true MOCKERY_INPACKAGE=true go generate ./...