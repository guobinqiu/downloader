Concurrent and resumable downloader

How to run

```
go install github.com/guobinqiu/downloader@latest

downloader --resourceUrl=https://storage.googleapis.com/golang/go1.6.3.darwin-amd64.pkg --saveDir=abc --workers=5 --resume=true
abc/go1.6.3.darwin-amd64.pkg.part2   --- [==================>-------------------------------------------------]  30%
abc/go1.6.3.darwin-amd64.pkg.part4   --- [=========================>------------------------------------------]  39%
abc/go1.6.3.darwin-amd64.pkg.part1   --- [=================>--------------------------------------------------]  28%
abc/go1.6.3.darwin-amd64.pkg.part0   --- [======================>---------------------------------------------]  35%
abc/go1.6.3.darwin-amd64.pkg.part3   --- [======================>---------------------------------------------]  35%
```

or

```
go run main.go --resourceUrl=https://storage.googleapis.com/golang/go1.6.3.darwin-amd64.pkg --saveDir=abc --workers=5 --resume=true
```

However, you can replace [this popular process bar](https://github.com/gosuri/uiprogress) with [my process bar](https://github.com/guobinqiu/process), but it's not Windows unsupported for now.
