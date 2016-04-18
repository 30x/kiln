test-build-and-package: test-source build-and-package

build-and-package: compile-linux build-image

test-source:
	go test -v $$(glide novendor)

compile-linux:	
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o build/shipyard .
	
build-image:
	docker build -t tnine/shipyard .

push-to-local:
	docker tag -f tnine/shipyard localhost:5000/tnine/shipyard
	docker push localhost:5000/tnine/shipyard

push-to-hub:
	docker tag -f tnine/shipyard tnine/shipyard:v1
	docker push tnine/shipyard:v1

deploy-to-kube:
	kubectl run shipyard --image=localhost:5000/tnine/shipyard:latest
