#Format is MAJOR . MINOR . PATCH

IMAGE_VERSION=0.1.20


test-build-and-package: test-source build-and-package

build-and-push-to-hub: build-and-package push-to-hub

build-and-package: compile-linux build-image

dev-build-push: compile-linux build-image push-to-dev

# builds and pushes the image used for local dev clusters
build-push-local: compile-linux build-image push-local-image

test-source:
	go test -v $$(glide novendor)

#Creates a test coverage file called coverage.txt.  This then opens a browser window to view the coverage
test-coverage:
	eval $( aws ecr get-login --region us-east-1)
	go test ./pkg/kiln -covermode=atomic -coverprofile=coverage.tmp
	go test ./pkg/kiln -covermode=atomic -coverprofile=coverage.tmp
	echo 'mode: atomic' > coverage.out
	tail -n +2 coverage.tmp >> coverage.out
	go tool cover -html=coverage.out -o=coverage.html

compile-linux:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o build/kiln .

build-image:
	docker build -t thirtyx/kiln .

push-to-dev:
	docker tag thirtyx/kiln thirtyx/kiln:dev
	docker push thirtyx/kiln:dev

push-local-image:
	docker tag thirtyx/kiln thirtyx/kiln:local
	docker push thirtyx/kiln:local

push-to-local:
	docker tag thirtyx/kiln localhost:5000/thirtyx/kiln
	docker push localhost:5000/thirtyx/kiln

push-to-hub:
	docker tag thirtyx/kiln thirtyx/kiln:$(IMAGE_VERSION)
	docker push thirtyx/kiln:$(IMAGE_VERSION)

run-dev:
	env AUTH_API_HOST="https://api.e2e.apigee.net/" REGISTRY_API_SERVER="http://api.shipyard.dev:31222" DEPLOY_STATE=DEV NO_REAP="true" PORT=5280 SHUTDOWN_TIMEOUT="0" DOCKER_PROVIDER=private DOCKER_REGISTRY_URL="localhost:5000" ORG_LABEL=org APP_NAME_LABEL=app go run main.go