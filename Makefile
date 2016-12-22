IMAGE := jetstackexperimental/vault-helper
IMAGE_TAG := 0.1

build:
	docker build -t $(IMAGE):$(IMAGE_TAG) .

push: build
	docker push $(IMAGE):$(IMAGE_TAG)
