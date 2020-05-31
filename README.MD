# package proxy

基于 golang.org/x/net/proxy 包，添加了UDP支持。
未实现部分：
1. UDP 数据的分片传输。其实不是很难，就是现在懒得做;
2. 