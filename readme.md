Concurrent and resumable downloader

```
fget --resourceUrl=https://storage.googleapis.com/golang/go1.6.3.darwin-amd64.pkg --saveDir=abc --workers=5 --resume=true
abc/go1.6.3.darwin-amd64.pkg.part2   --- [==================>-------------------------------------------------]  30%
abc/go1.6.3.darwin-amd64.pkg.part4   --- [=========================>------------------------------------------]  39%
abc/go1.6.3.darwin-amd64.pkg.part1   --- [=================>--------------------------------------------------]  28%
abc/go1.6.3.darwin-amd64.pkg.part0   --- [======================>---------------------------------------------]  35%
abc/go1.6.3.darwin-amd64.pkg.part3   --- [======================>---------------------------------------------]  35%
```
