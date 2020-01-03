# 生成火焰图，用于分析cpu/内存(使用go内置的net/http/pprof)
analysis:
	go-torch -u http://localhost:6060 -t 120 -f analysis.svg

#生成临时分配内存的火焰图
heap_alloc_space:
	go-torch -alloc_space http://127.0.0.1:6060/debug/pprof/heap --colors=mem -t 120 -f heap_alloc_space.svg

#生成常驻内存的火焰图
heap_inuse_space:
	go-torch -inuse_space http://127.0.0.1:6060/debug/pprof/heap --colors=mem -t 120 -f heap_inuse_space.svg

#使用go tool查看内存信息
heap:
	go tool pprof -inuse_space http://127.0.0.1:6060/debug/pprof/heap

#生成内存svg图
heap_svg:
	go tool pprof -alloc_space -cum -svg http://127.0.0.1:6060/debug/pprof/heap > heap.svg
