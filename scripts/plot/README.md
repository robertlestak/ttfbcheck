# TTFB plotter

## dependencies

- `jq`
- `gnuplot`
- [robertlestak/ttfbcheck](https://github.com/robertlestak/ttfbcheck)

## usage

```bash
# plot.sh [url] [request count] [concurrency]
./plot.sh "https://example.com" 50 1,10,50,100
```
