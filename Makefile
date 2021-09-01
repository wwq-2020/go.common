all: genenv generrcode
genenv: cleanenv
	@protoc  --go_out=env env/env.proto
	@mv env/github.com/wwq-2020/go.common/env/env.pb.go env
	@rm -rf env/github.com
generrcode: cleanerrcode
	@protoc  --go_out=errcode errcode/errcode.proto
	@mv errcode/github.com/wwq-2020/go.common/errcode/errcode.pb.go errcode
	@rm -rf errcode/github.com
cleanenv:
	@rm -rf env/env.pb.go
cleanerrcode:
	@rm -rf errcode/errcode.pb.go