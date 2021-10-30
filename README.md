# gomock
---

`gomock` is a source code generator written for and in go, with the express purpose of generating malware simulants for deployment in containerized environments.

## Usage

Build gomock:

```bash
go build -o gomock main.go
```

Generate the test program:

```bash
./gomock myPackageName ./testProgram.json
```

## What's next

`gomock` is currently in very early stages of development, and only the `file/stat` operation is currently fully implemented.

Future supported operations include:
 - Sending HTTPS requests
 - Inspecting OS processes
 - Dynamic module loading
 - File I/O operations
 - Embedding string resources
 - etc.