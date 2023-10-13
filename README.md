# quick-quic

Demo for QUIC connection between server, publisher clients and subscriber clients. More information can be found in [quic-go](https://github.com/quic-go/quic-go) repo.

### Running the app

- To run the server, execute this command from the root:

```
go run server/main.go
```

- To run the publisher client, execute this command from the root:
```
go run client-pub/main.go
```

- To run the subscriber client, execute this command from the root:
```
go run client-sub/main.go
```

Multiple pub and sub clients are supported

### Testing

*TBD.
