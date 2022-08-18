IMAGE := keppel.eu-de-1.cloud.sap/ccloud/k8s-conntrack-nanny
VERSION:= v0.2.4


build:
	go build -o bin/k8s-conntrack-nanny

docker:
	docker build -t $(IMAGE):$(VERSION) .
push:
	docker push $(IMAGE):$(VERSION)

docker-push-mac:
	docker buildx build  --platform linux/amd64 . -t ${IMAGE}:${VERSION} --push

test:
	docker run --cap-add=NET_ADMIN -v $(HOME)/.kube/config:/kubeconfig $(IMAGE):$(VERSION) --kubeconfig /kubeconfig --context $(KUBECONTEXT)
