# bbenchmark测试记录

第一次测试 2025/2/19 13:57
goos: linux
goarch: amd64
pkg: github.com/muidea/magicCommon/foundation/cache
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
=== RUN   BenchmarkKVCache_Put
BenchmarkKVCache_Put
BenchmarkKVCache_Put-8                            843104              1468 ns/op             128 B/op          5 allocs/op
=== RUN   BenchmarkKVCache_Fetch
BenchmarkKVCache_Fetch
BenchmarkKVCache_Fetch-8                          929329              1596 ns/op              64 B/op          4 allocs/op
=== RUN   BenchmarkKVCache_ConcurrentPut
BenchmarkKVCache_ConcurrentPut
BenchmarkKVCache_ConcurrentPut-8                  915916              1305 ns/op             128 B/op          5 allocs/op
=== RUN   BenchmarkKVCache_ConcurrentFetch
BenchmarkKVCache_ConcurrentFetch
BenchmarkKVCache_ConcurrentFetch-8               1000000              1268 ns/op              64 B/op          4 allocs/op
=== RUN   BenchmarkMemoryCache_Put
BenchmarkMemoryCache_Put
BenchmarkMemoryCache_Put-8                        109248             10446 ns/op             477 B/op         10 allocs/op
=== RUN   BenchmarkMemoryCache_Fetch
BenchmarkMemoryCache_Fetch
BenchmarkMemoryCache_Fetch-8                      798748              1506 ns/op              96 B/op          5 allocs/op
=== RUN   BenchmarkMemoryCache_ConcurrentPut
BenchmarkMemoryCache_ConcurrentPut
BenchmarkMemoryCache_ConcurrentPut-8              236346              5929 ns/op             467 B/op         10 allocs/op
=== RUN   BenchmarkMemoryCache_ConcurrentFetch
BenchmarkMemoryCache_ConcurrentFetch
BenchmarkMemoryCache_ConcurrentFetch-8            959468              1349 ns/op              96 B/op          5 allocs/op
PASS
ok      github.com/muidea/magicCommon/foundation/cache  11.730s

第二次测试 2025/2/19 15:01
goos: linux
goarch: amd64
pkg: github.com/muidea/magicCommon/foundation/cache
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
=== RUN   BenchmarkKVCache_Put
BenchmarkKVCache_Put
BenchmarkKVCache_Put-8                           1449297               846.2 ns/op           128 B/op          5 allocs/op
=== RUN   BenchmarkKVCache_Fetch
BenchmarkKVCache_Fetch
BenchmarkKVCache_Fetch-8                        20868300                60.57 ns/op            0 B/op          0 allocs/op
=== RUN   BenchmarkKVCache_ConcurrentPut
BenchmarkKVCache_ConcurrentPut
BenchmarkKVCache_ConcurrentPut-8                 1550900               763.8 ns/op           128 B/op          5 allocs/op
=== RUN   BenchmarkKVCache_ConcurrentFetch
BenchmarkKVCache_ConcurrentFetch
BenchmarkKVCache_ConcurrentFetch-8              29775656                35.35 ns/op            0 B/op          0 allocs/op
=== RUN   BenchmarkMemoryCache_Put
BenchmarkMemoryCache_Put
BenchmarkMemoryCache_Put-8                        237350              5769 ns/op             466 B/op         10 allocs/op
=== RUN   BenchmarkMemoryCache_Fetch
BenchmarkMemoryCache_Fetch
BenchmarkMemoryCache_Fetch-8                     1505180               791.8 ns/op            96 B/op          5 allocs/op
=== RUN   BenchmarkMemoryCache_ConcurrentPut
BenchmarkMemoryCache_ConcurrentPut
BenchmarkMemoryCache_ConcurrentPut-8              475258              3053 ns/op             466 B/op         10 allocs/op
=== RUN   BenchmarkMemoryCache_ConcurrentFetch
BenchmarkMemoryCache_ConcurrentFetch
BenchmarkMemoryCache_ConcurrentFetch-8           1773027               868.9 ns/op            96 B/op          5 allocs/op
PASS
ok      github.com/muidea/magicCommon/foundation/cache  13.724s


第三次测试 2025/2/19 15:38
goos: linux
goarch: amd64
pkg: github.com/muidea/magicCommon/foundation/cache
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
=== RUN   BenchmarkKVCache_Put
BenchmarkKVCache_Put
BenchmarkKVCache_Put-8                            924957              1096 ns/op             128 B/op          5 allocs/op
=== RUN   BenchmarkKVCache_Fetch
BenchmarkKVCache_Fetch
BenchmarkKVCache_Fetch-8                        17631177                60.07 ns/op            0 B/op          0 allocs/op
=== RUN   BenchmarkKVCache_ConcurrentPut
BenchmarkKVCache_ConcurrentPut
BenchmarkKVCache_ConcurrentPut-8                 1000000              1020 ns/op             128 B/op          5 allocs/op
=== RUN   BenchmarkKVCache_ConcurrentFetch
BenchmarkKVCache_ConcurrentFetch
BenchmarkKVCache_ConcurrentFetch-8              24898582                53.60 ns/op            0 B/op          0 allocs/op
=== RUN   BenchmarkMemoryCache_Put
BenchmarkMemoryCache_Put
BenchmarkMemoryCache_Put-8                        174165              6395 ns/op             425 B/op         10 allocs/op
=== RUN   BenchmarkMemoryCache_Fetch
BenchmarkMemoryCache_Fetch
BenchmarkMemoryCache_Fetch-8                    16614080                62.59 ns/op            0 B/op          0 allocs/op
=== RUN   BenchmarkMemoryCache_ConcurrentPut
BenchmarkMemoryCache_ConcurrentPut
BenchmarkMemoryCache_ConcurrentPut-8              268110              4239 ns/op             451 B/op         10 allocs/op
=== RUN   BenchmarkMemoryCache_ConcurrentFetch
BenchmarkMemoryCache_ConcurrentFetch
BenchmarkMemoryCache_ConcurrentFetch-8          25848128                49.55 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/muidea/magicCommon/foundation/cache  10.598s

