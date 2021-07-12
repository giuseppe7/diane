app = diane

.PHONY: build

build:
	@echo
	@echo "⋮⋮ Building..."
	go build -ldflags "-X main.version=`cat build_number``date -u +.%Y%m%d%H%M%S`"

test: build
	@echo
	@echo "⋮⋮ Testing..."
	go test -count=1 ./... -coverprofile cover.out

review: test
	@echo
	@echo "⋮⋮ Reviewing tests..."
	go tool cover -html cover.out

container: test
	@echo
	@echo "⋮⋮ Creating container..."
	docker build -f ./build/package/Dockerfile -t ${app} .

local: container
	@echo
	@echo "⋮⋮ Creating local environment..."
	docker compose -f ./deployments/docker-compose.yaml --project-name ${app} up -d --force-recreate
	docker ps | grep ${app}

clean-local:
	docker compose -f ./deployments/docker-compose.yaml --project-name ${app} down

all: build test
	@echo