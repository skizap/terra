all: buildkit

buildkit:
	@buildctl build --frontend=dockerfile.v0 --frontend-opt filename=Dockerfile --local context=. --local dockerfile=. --progress plain --exporter=local --exporter-opt output=./""

.PHONY: buildkit
