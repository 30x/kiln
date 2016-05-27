IMAGE_VERSION=0.0.2-dev

test-build-and-package: test-source build-and-package

build-and-package: compile-linux build-image

test-source:
	go test -v $$(glide novendor)

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
