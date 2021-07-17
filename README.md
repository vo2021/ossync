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
  -interval int
        the interval between sync (default 10)
  -output string
        the local folder path to sync to
  -profile string
        the OCI profile name (default "DEFAULT")
```


![](dev.png)
