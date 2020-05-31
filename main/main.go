package main

import (
	"log"

	"github.com/anyktx/proxy"
)

func main() {
	// 这里指定的是与SOCKS代理建立的连接类型，只可以使用 TCP(4/6)
	dialer, e := proxy.SOCKS5("tcp", "127.0.0.1:5719", nil, nil)
	if e != nil {
		log.Fatalln(e)
	}
	// 这里是需要进行的数据传输类型，为 udp(4/6) 或 tcp(4/6)
	conn, e := dialer.Dial("udp", "qhelper.top:5530")
	if e != nil {
		log.Fatalln(e)
	}
	defer conn.Close()

	_, e = conn.Write([]byte("Hi, I'm client from socks5 UDP!"))
	if e != nil {
		log.Fatalln(e)
	}

	buf := make([]byte, 64)
	n, e := conn.Read(buf)
	if e != nil {
		log.Fatalln(e)
	}
	log.Println(string(buf[:n]))
	return
}
