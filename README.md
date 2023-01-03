# Rolling hash file diff

## Commands

1. Create an signature from `testdata/000.old` file with chunk size 512 bytes and store results in `testdata/000.signature`

The signature file stores chunk size and rolling checksums and strong checksums for every chunk.

```
go run cmd/main.go signature testdata/000.old testdata/000.signature --chunksize 512
```

2. Calculate delta between signature `testdata/000.signature` and `testdata/000.new` file,
store results in `testdata/000.delta`

The delta file stores operations required to make from the old file a new one.

```
go run cmd/main.go delta testdata/000.signature testdata/000.new testdata/000.delta
```
