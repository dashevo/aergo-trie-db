# Universe Tree DB

> A gRPC service wrapped around the [Aergo State Trie](https://github.com/aergoio/aergo/tree/develop/pkg/trie)

This service also includes metadata about the trees stored in BadgerDB.

## Table of Contents

- [Build](#build)
- [Usage](#usage)
- [Maintainer](#maintainer)
- [License](#license)

## Build

```sh
git clone https://github.com/dashevo/universe-tree-db.git
cd universe-tree-db

make
```

### Prerequisites

- Go 1.12+

### Compile Protobuf Schema

The default `make` target compiles this before building the server and example client binaries. You can also compile protocol buffers for Go using `protoc` target:

```sh
make protoc
```

These generated files are considered built artifacts and should not be checked in to source code repository.

## Usage

Create data dir (first time only) and start the server:

```sh
mkdir $PWD/data

UNIDB_DIR=$PWD/data ./bin/server
```

Test w/the example client:

```sh
# list trees
./bin/client list

# create a tree called 'x'
./bin/client create x

# update tree w/hash of string 'hi'
./bin/client update x hi

# get value of string hash from tree
./bin/client get x hi
```

## Maintainer

[@nmarley](https://github.com/nmarley)

## License

[ISC](LICENSE) &copy; Dash Core Group, Inc.
