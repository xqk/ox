package protoc

//ProtocHelpTemplate ...
const ProtocHelpTemplate = `
ox [commands|flags]

The commands & flags are:
  protoc        ox protoc tools
  -g,--grpc     whether to generate GRPC code
  -s,--server   whether to generate grpc server code
  -c,--client   generate grpc server code
  -f,--file     path of proto file
  -o,--out      path of code generation
  -p,--prefix   prefix(current project name)
Examples:
   # Generate GRPC code from the Proto file 
   # -f: Proto file address -o: Code generation path -g: Whether to generate GRPC code
   ox protoc -f ./pb/hello/hello.proto -o ./pb/hello -g
   # According to the proto file, generate the server implementation
   # -f: Proto file address -o: Code generation path -p:prefix(Current project name) -g: Whether to generate Server code
   ox protoc -f ./pb/hello/hello.proto -o ./internal/app/grpc -p ox-demo -s
  
`
