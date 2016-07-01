#Format is MAJOR . MINOR . PATCH

IMAGE_VERSION=0.1.12


test-build-and-package: test-source build-and-package

build-and-push-to-hub: build-and-package push-to-hub

build-and-package: compile-linux build-image


test-source:
	go test -v $$(glide novendor)

#Creates a test coverage file called coverage.txt.  This then opens a browser window to view the coverage
test-coverage:
	go test ./pkg/shipyard -covermode=atomic -coverprofile=coverage.tmp
	go test ./pkg/shipyard -covermode=atomic -coverprofile=coverage.tmp
	echo 'mode: atomic' > coverage.out
	tail -n +2 coverage.tmp >> coverage.out
	go tool cover -html=coverage.out -o=coverage.html

compile-linux:	
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o build/shipyard .
	
build-image:
	docker build -t thirtyx/shipyard .

push-to-local:
	docker tag -f thirtyx/shipyard localhost:5000/thirtyx/shipyard
	docker push localhost:5000/thirtyx/shipyard

push-to-hub:
	docker tag -f thirtyx/shipyard thirtyx/shipyard:$(IMAGE_VERSION)
	docker push thirtyx/shipyard:$(IMAGE_VERSION)

deploy-to-kube:
	kubectl run shipyard --image=localhost:5000/thirtyx/shipyard:latest
