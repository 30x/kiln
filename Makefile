#Format is MAJOR . MINOR . PATCH

IMAGE_VERSION=0.1.18


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
	docker tag -f thirtyx/kiln thirtyx/kiln:dev
	docker push thirtyx/kiln:dev

push-local-image:
	docker tag -f thirtyx/kiln thirtyx/kiln:local
	docker push thirtyx/kiln:local

push-to-local:
	docker tag -f thirtyx/kiln localhost:5000/thirtyx/kiln
	docker push localhost:5000/thirtyx/kiln

push-to-hub:
	docker tag -f thirtyx/kiln thirtyx/kiln:$(IMAGE_VERSION)
	docker push thirtyx/kiln:$(IMAGE_VERSION)

deploy-to-kube:
	kubectl run kiln --image=localhost:5000/thirtyx/kiln:latest

deploy-dev:
	kubectl create -f kubernetes/dev-deployment.yaml --namespace=shipyard
