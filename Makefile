build-to-docker: main.go
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o build/shipyard .
	docker build -t tnine/shipyard .

push-to-local:
	docker tag -f tnine/shipyard localhost:5000/tnine/shipyard
	docker push localhost:5000/tnine/shipyard

push-to-hub:
	docker tag -f tnine/shipyard tnine/shipyard:v0
	docker push tnine/shipyard:v0

deploy-to-kube:
	kubectl run shipyard --image=localhost:5000/tnine/shipyard:latest
