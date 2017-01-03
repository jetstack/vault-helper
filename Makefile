IMAGE := jetstackexperimental/vault-helper
IMAGE_TAG := 0.1

build:
	docker build -t $(IMAGE):$(IMAGE_TAG) .

push: build test
	docker push $(IMAGE):$(IMAGE_TAG)

test:
	bundle install
	bundle exec rake
