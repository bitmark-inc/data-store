.PHONY: pds cds
dist =

default: build


pds:
	go build -o bin/pds commands/pds/main.go

cds:
	go build -o bin/cds commands/cds/main.go

run-pds: pds
	./bin/pds -c  commands/pds/config.yaml

run-cds: cds
	./bin/cds -c  commands/cds/config.yaml

bin: pds cds

build-pds:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t data-store:pds-$(dist) . -f Dockerfile-pds
	docker tag data-store:pds-$(dist)  083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/data-store:pds-$(dist)

build-cds:
ifndef dist
	$(error dist is undefined)
endif
	docker build --build-arg dist=$(dist) -t data-store:cds-$(dist) . -f Dockerfile-cds
	docker tag data-store:cds-$(dist)  083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/data-store:cds-$(dist)

build: build-pds build-cds

push:
ifndef dist
	$(error dist is undefined)
endif
	aws ecr get-login-password | docker login --username AWS --password-stdin 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/data-store:pds-$(dist)
	docker push 083397868157.dkr.ecr.ap-northeast-1.amazonaws.com/data-store:cds-$(dist)

clean:
	rm -r bin
