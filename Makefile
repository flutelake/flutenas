
GO_MODULE_NAME = flutelake/fluteNAS
VERSION_FLAG=-X '$(GO_MODULE_NAME)/pkg/version.gitBranch=`git branch --show-current`' \
-X '$(GO_MODULE_NAME)/pkg/version.gitCommit=`git rev-parse HEAD`' \
-X '$(GO_MODULE_NAME)/pkg/version.buildUser=`whoami`' \
-X '$(GO_MODULE_NAME)/pkg/version.buildDate=`date +'%Y-%m-%dT%H:%M:%SZ'`'

GO_LDFLAGS :=-ldflags "-s $(VERSION_FLAG)"
GO_ENV := CGO_ENABLED=0 GOOS=linux
GO_ENV_AMD64 := GOARCH=amd64
GO_ENV_ARM64:= GOARCH=arm64
GO_ENV_SW64:= GOARCH=sw64
GO_ENV_MIPS64LE := GOARCH=mips64le
DIST_DIR_AMD64 := dist/x86_64
DIST_DIR_ARM64 := dist/aarch64
DIST_DIR_SW64 := dist/sw64
DIST_DIR_MIPS64LE := dist/mips64le

all:
	cd frontend/flute-nas/ && pnpm build
	$(GO_ENV) $(GO_ENV_AMD64) go build $(GO_LDFLAGS) -o $(DIST_DIR_AMD64)/flute-nas-server cmd/fluteNAS/main.go

#################### dev commands ####################
deploy:
	rsync -avP dist/x86_64/flute-nas-server root@10.0.1.10:/opt/flute-nas/
	ssh root@10.0.1.10 "systemctl restart flute-nas"
	
mpush: all deploy

demopush:
	rsync -avP -e "ssh -p 2030" dist/x86_64/flute-nas-server root@47.243.81.136:/opt/flute-nas/
	ssh root@47.243.81.136 -p 2030 "systemctl restart flute-nas"

#################### dev commands end ####################

frontend:
	cd frontend/flute-nas/ && pnpm build

server:
	$(GO_ENV) $(GO_ENV_AMD64) go build $(GO_LDFLAGS) -o $(DIST_DIR_AMD64)/flute-nas-server cmd/fluteNAS/main.go

start_frontend:
	cd frontend/flute-nas/ && pnpm dev
