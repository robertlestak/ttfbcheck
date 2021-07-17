# ttfbcheck

Check the time to first byte of a URL.

## Install

```bash
# local
go build -o ttfbcheck *.go

# docker
docker build . -t ttfbcheck
```

## Usage

```bash
ttfbcheck
  -concurrent int
        concurrent tests to run (default 10)
  -log string
        log level. (debug,info,warn,error,fatal,color,nocolor,json) (default "info")
  -out string
        output format. (jsonl, csv) (default "jsonl")
  -total int
        total tests to run (default 10)
  -url string
        url to test
```
