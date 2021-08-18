# ossync - syncing from OCI Object Storage

## Compile

```go
go build -o ossync.exe
```

## Usage

```bash
$./ossync.exe --help
Usage of ./ossync.exe:
  -bucket string
        the OCI bucket which is synced to local (default "bucket-name")
  -debug
        debug mode
  -interval int
        the interval between sync (default 10)
  -output string
        the local folder path to sync to
  -prefix string
        the prefix of a folder or file in the OCI bucket which is synced to local
  -profile string
        the OCI profile name (default "DEFAULT")
```

## TODO

- sync sub folder

