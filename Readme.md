当我们需要对go test中某些test/benchmark进行profiling时，我们可以使用类似的方法。例如我们可以先使用go test 内置的参数生成pprof数据，然后借助go tool pprof/go-torch来分析。

#生成cpu、mem的pprof文件
go test -bench=BenchmarkStorageXXX -cpuprofile cpu.out -memprofile mem.out

此时会生成一个二进制文件和2个pprof数据文件，例如
storage.test cpu.out mem.out

然后使用go-torch来分析，二进制文件放前面
#分析cpu
go-torch storage.test cpu.out
#分析内存
go-torch --colors=mem -alloc_space storage.test mem.out

go-torch --colors=mem -inuse_space storage.test mem.out
