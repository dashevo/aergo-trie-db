# Aergo Trie DB

> A gRPC API for [Aergo Trie library](https://github.com/aergoio/aergo/tree/develop/pkg/trie)

# Build gRPC server

```bash
protoc -I proto/ proto/aergo-trie-db.proto --go_out=plugins=grpc:routeguide
```

## License

[MIT](LICENSE) &copy; Dash Core Group, Inc.
