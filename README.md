# Rolling hash file diff

## Commands

1. Create an signature from `testdata/000.old` file with chunk size 512 bytes and store results in `testdata/000.signature`

```
go run cmd/main.go signature testdata/000.old testdata/000.signature --chunksize 512
```

2. Calculate delta between signature `testdata/000.signature` and `testdata/000.new` file,
store results in `testdata/000.delta`

```
go run cmd/main.go delta testdata/000.signature testdata/000.new testdata/000.delta
```
