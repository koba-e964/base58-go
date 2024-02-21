# base58 ![Go](https://github.com/koba-e964/base58-go/actions/workflows/go.yml/badge.svg?query=branch%3Amain)
`base58` is a fast implementation of [Base58](https://en.bitcoin.it/wiki/Base58Check_encoding) with resistance to side-channel attacks.

# Benchmarks
```console
$ go test -bench=5K .
goos: darwin
goarch: arm64
pkg: github.com/koba-e964/base58-go
BenchmarkDecode_ConstantTime_5K-8            660           1772633 ns/op           3.91 MB/s        5376 B/op          1 allocs/op
BenchmarkDecode_VariableTime_5K-8           8433            134081 ns/op          51.63 MB/s      338044 B/op        126 allocs/op
BenchmarkEncode_ConstantTime_5K-8            409           2905121 ns/op           1.72 MB/s       12288 B/op          2 allocs/op
BenchmarkEncode_VariableTime_5K-8            982           1216755 ns/op           4.11 MB/s       24664 B/op        686 allocs/op
PASS
ok      github.com/koba-e964/base58-go  5.459s
```

The performance of base58.VartimeEncode and base58.VartimeDecode is comparable to that of [github.com/btcsuite/btcd/btcutil/base58](https://github.com/btcsuite/btcd/tree/13152b35e191385a874294a9dbc902e48b1d71b0/btcutil/base58):

```console
$ go test -bench=5K .
goos: darwin
goarch: arm64
pkg: github.com/btcsuite/btcd/btcutil/base58
BenchmarkBase58Encode_5K-8           974           1222350 ns/op           4.09 MB/s
BenchmarkBase58Decode_5K-8          8229            131085 ns/op          52.10 MB/s
PASS
ok      github.com/btcsuite/btcd/btcutil/base58 3.228s
$ git log --max-count=1
commit 13152b35e191385a874294a9dbc902e48b1d71b0 (HEAD -> master, origin/master, origin/HEAD)
Merge: 8e53942b e307ad12
Author: Olaoluwa Osuntokun <laolu32@gmail.com>
Date:   Tue Jan 23 20:25:55 2024 -0800

    Merge pull request #2089 from kcalvinalvin/2024-01-03-add-last-flush-time-on-initconsistentstate
    
    blockchain: set the lastflushtime when setting the lastflushhash
$ go version
go version go1.22.0 darwin/arm64
```

Besides, this package offers the constant-time version of these functions, namely base58.Encode and base58.Decode, which prevents side-channel attacks such as timing attacks.

# Docs
https://pkg.go.dev/github.com/koba-e964/base58-go
