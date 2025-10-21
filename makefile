# 生成CPU性能分析文件  可以使用 go tool pprof -http=:9998 cpu-std.pprof 图形化查看
gen-cpu-pprof:
	go test -benchmem -run=^$$ -bench ^BenchmarkCopyStructMy$$ . -v -cpuprofile cpu-y.pprof  -benchtime=5s
	go test -benchmem -run=^$$ -bench ^BenchmarkCopyStructStd$$ . -v -cpuprofile cpu-std.pprof  -benchtime=5s
	go test -benchmem -run=^$$ -bench ^BenchmarkJsonMarshalCopy$$ . -v -cpuprofile cpu-json.pprof  -benchtime=5s
