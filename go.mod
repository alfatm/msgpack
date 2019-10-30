module github.com/alfatm/msgpack

require (
	github.com/kr/pretty v0.1.0 // indirect
	github.com/vmihailenco/msgpack/v4 v4.0.0-00010101000000-000000000000
	github.com/vmihailenco/tagparser v0.1.0
	gitlab.msoft.io/hub/zerror v1.1.1
	google.golang.org/appengine v1.6.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
)

replace github.com/vmihailenco/msgpack/v4 => ./

go 1.13
