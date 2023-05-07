# Concurrent and resumable downloader

Mine wget, implemented by Go.

```
go install github.com/guobinqiu/downloader@latest

downloader --resourceUrl=https://storage.googleapis.com/golang/go1.6.3.darwin-amd64.pkg --saveDir=abc --workers=5 --resume=true
```

Params:

name|required|desc
---|---|---
resourceUrl|Y|Download url for a file
saveDir|Y|Local directory for downloaded files
worker|N|Number of concurrency, default to CPU nums
resume|N|Continue from last time breakpoint or not, default to true

Output:

```
abc/go1.6.3.darwin-amd64.pkg.part2   --- [==================>-------------------------------------------------]  30%
abc/go1.6.3.darwin-amd64.pkg.part4   --- [=========================>------------------------------------------]  39%
abc/go1.6.3.darwin-amd64.pkg.part1   --- [=================>--------------------------------------------------]  28%
abc/go1.6.3.darwin-amd64.pkg.part0   --- [======================>---------------------------------------------]  35%
abc/go1.6.3.darwin-amd64.pkg.part3   --- [======================>---------------------------------------------]  35%
```

## License

MIT
