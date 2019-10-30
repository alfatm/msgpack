module github.com/alfatm/msgpack

require (
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/vmihailenco/tagparser v0.1.0
	gitlab.msoft.io/hub/zerror v1.1.1
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80 // indirect
	google.golang.org/appengine v1.6.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
)

replace github.com/vmihailenco/msgpack/v4 github.com/alfatm/msgpack

go 1.13
