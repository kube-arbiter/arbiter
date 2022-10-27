observiability indicant proto package.

### protoc installation

Enter the protobuf release page and select the compressed package file suitable for your operating system.

For example, we can use the one from 'https://github.com/protocolbuffers/protobuf/releases/download/v21.7/protoc-21.7-osx-x86\_64.zip'

unzip protoc-21.7-osx-x86\_64.zip and enter protoc-21.7-osx-x86\_64
```
$ cd protoc-3.20.3-osx-x86_64/bin
```
Move the started protoc binary file to any path added to the environment variable, such as $GOPATH/bin. It is not recommended putting it directly with the next path of the system.

Verify the installation result
```
$ mv protoc $GOPATH/bin
$ protoc --version
libprotoc 3.21.7
```

### protoc-gen-* installation

Download and install protoc-gen-go
```
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### generate go proto files

```shell
# generate go code of OBI interface using proto file
protoc -I=. --go_out=. --go-grpc_out=. pkg/proto/observiability.proto
# generate go code of executor interface using proto file
protoc -I=. -I=./vendor --go_out=pkg/proto --go-grpc_out=pkg/proto pkg/proto/executor.proto
```

### problem

If you encounter the following problems, please try to modify the `option go_package "runtime"` of `vendor/k8s.io/apimachinery/pkg/runtime/generated.proto` to `option go_package "k8s.io/apimachinery/pkg/runtime"`.



```shell
protoc-gen-go: invalid Go import path "runtime" for "k8s.io/apimachinery/pkg/runtime/generated.proto"

The import path must contain at least one period ('.') or forward slash ('/') character.

See https://developers.google.com/protocol-buffers/docs/reference/go-generated#package for more information.

--go_out: protoc-gen-go: Plugin failed with status code 1.
```
