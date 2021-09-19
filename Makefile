app = diane

.PHONY: build

build:
	@echo
	@echo "⠿ Building..."
	go build -ldflags "-X main.version=`cat build_number``date -u +.%Y%m%d%H%M%S`"

test: build
	@echo
	@echo "⠿ Testing..."
	go test -count=1 ./... -coverprofile cover.out

review: cover.out
	@echo
	@echo "⠿ Reviewing tests..."
	go tool cover -html cover.out

container: test
	@echo
	@echo "⠿ Creating container..."
	docker build -f ./build/package/Dockerfile -t ${app} .

local: container
	@echo
	@echo "⠿ Creating local environment..."
	docker compose -f ./deployments/docker-compose.yaml --project-name ${app} up -d 
	@echo 
	docker ps | grep ${app}
	@echo 
	@echo "Local Grafana URL:"
	@docker ps | grep diane | grep grafana | perl -pe 's/.* (0.0.0.0:.*?)->3000.*/http:\/\/\1/'

clean-local:
	@echo
	@echo "⠿ Deleting local environment..."
	docker compose -f ./deployments/docker-compose.yaml --project-name ${app} down

all: build test
	@echo