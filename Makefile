IMAGE := sapcc/k8s-conntrack-nanny
VERSION:= v0.1.0


build:
	go build -o bin/k8s-conntrack-nanny

docker:
	docker build -t $(IMAGE):$(VERSION) .
push: 
	docker push $(IMAGE):$(VERSION)

test:
	docker run --cap-add=NET_ADMIN -v $(HOME)/.kube/config:/kubeconfig $(IMAGE):$(VERSION) --kubeconfig /kubeconfig --context $(KUBECONTEXT)
