TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=local.com
NAMESPACE=test
NAME=postgrestable
BINARY=terraform-provider-${NAME}
VERSION=0.1
OS_ARCH=darwin_amd64

default: install

init:
	go mod init terraform-provider-postgrestable
	go mod vendor

build:
	go build -o ${BINARY}


install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

local_install:install
	rm -f examples/.terraform.lock.hcl
	cd examples && terraform init && terraform apply --auto-approve

test: 
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

format:
	go fmt ./...

.PHONY: docker
docker:
	docker-compose -f ops/docker-compose.yml up -d --renew-anon-volumes

.PHONY: lint
lint:
	go fmt ./...