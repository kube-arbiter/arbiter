observiability indicant proto package.

### protoc installation
Enter the protobuf release page and select the compressed package file suitable for your operating system.

For example, we can use the one from 'https://github.com/protocolbuffers/protobuf/releases/download/v3.20.3/protoc-3.20.3-osx-x86_64.zip'

Unzip protoc-3.20.3-osx-x86_64.zip and enter protoc-3.20.3-osx-x86_64
```
$ cd protoc-3.20.3-osx-x86_64/bin
```
Move the started protoc binary file to any path added to the environment variable, such as $GOPATH/bin. It is not recommended putting it directly with the next path of the system.

Verify the installation result
```
$ mv protoc $GOPATH/bin
$ protoc --version
libprotoc 3.20.3
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
protoc --go_out=. --go-grpc_out=. ./observiability.proto
# generate go code of executor interface using proto file
protoc --go_out=. --go-grpc_out=. ./executor.proto
```
