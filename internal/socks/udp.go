// file udp.go
// create by ayk
// 实现 SOCKS5 的 UDP 穿透

package socks

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func newUDPConn(tcpConn net.Conn, addr net.Addr, originAddr net.Addr) (_ net.Conn, err error) {
	uc := &udpconn{
		tcpConn: tcpConn,
		proxy:   addr.(*Addr),
	}
	// 建立 UDP 连接
	uc.udpConn, err = net.Dial("udp", addr.String())
	if origin, ok := originAddr.(*Addr); ok {
		uc.header = make([]byte, 0, 6+len(origin.IP)+len(origin.Name))
		uc.header = append(uc.header, 0x00, 0x00, 0x00)
		if origin.IP != nil {
			if ip4 := origin.IP.To4(); ip4 != nil {
				uc.header = append(uc.header, AddrTypeIPv4)
				uc.header = append(uc.header, ip4...)
			} else if ip6 := origin.IP.To16(); ip6 != nil {
				uc.header = append(uc.header, AddrTypeIPv6)
				uc.header = append(uc.header, ip6...)
			} else {
				return nil, fmt.Errorf("Invalid IP: %v", origin.IP)
			}
		} else if origin.Name != "" {
			uc.header = append(uc.header, AddrTypeFQDN)
			uc.header = append(uc.header, byte(len(origin.Name)))
			uc.header = append(uc.header, []byte(origin.Name)...)
		} else {
			return nil, fmt.Errorf("Invalid addr: %v", originAddr)
		}
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(origin.Port))
		uc.header = append(uc.header, b[:2]...)
	}
	return uc, err
}

// 记录UDP连接信息
type udpconn struct {
	tcpConn net.Conn // 与代理握手使用的TCP连接
	udpConn net.Conn // 与代理进行通信的UDP连接
	remote  net.Addr // 远程目标地址
	local   net.Addr // 本地地址
	proxy   *Addr    // 代理服务器地址
	header  []byte   // UDP 头信息。不支持分片时为固定，支持分片此字段酌情处理
}

// +----+------+------+----------+----------+----------+
// |RSV | FRAG | ATYP | DST.ADDR | DST.PORT |　　DATA　|
// +----+------+------+----------+----------+----------+
// |　2 |　 1　| 　1　| Variable | 　2　　　| Variable |
// +----+------+------+----------+----------+----------+

// TODO: 目前不支持分片数据，待有需要时再完成
func (u *udpconn) Read(data []byte) (n int, err error) {
	n, err = u.udpConn.Read(data)
	if n < 6 {
		return 0, fmt.Errorf("Invalid data length")
	}
	// 不支持分片，所以需要[2]为0
	if data[2] != 0 {
		return 0, fmt.Errorf("Not support fragment")
	}

	addr := Addr{}
	cursor := 4 // 传输的数据起始位置
	addrType := data[3]
	switch addrType {
	case AddrTypeIPv4:
		addr.IP = data[cursor : cursor+4]
		cursor += 4
	case AddrTypeIPv6:
		addr.IP = data[cursor : cursor+16]
		cursor += 16
	case AddrTypeFQDN:
		addr.Name = string(data[cursor+1 : cursor+1+int(data[4])])
		cursor += int(data[4]) + 1
	default:
		return 0, fmt.Errorf("Invalid address type: %d", addrType)
	}
	addr.Port = int(binary.BigEndian.Uint16(data[cursor : cursor+2]))
	fmt.Println("Target address: ", addr.String())
	// 端口
	cursor += 2
	copy(data, data[cursor:n])
	return n - cursor, nil
}

// TODO: 目前不支持分片数据，待有需要时再完成
func (u *udpconn) Write(b []byte) (n int, err error) {
	n, err = u.udpConn.Write(append(u.header, b[:]...))
	if err != nil {
		return 0, err
	}
	if n == len(u.header)+len(b) {
		return n - len(u.header), nil
	}
	err = fmt.Errorf("invalid write size: should %d but %d", len(u.header)+len(b), n)
	return 0, err
}

func (u *udpconn) Close() error {
	var e error
	if u.udpConn != nil {
		e = u.udpConn.Close()
		if e != nil {
			return e
		}
	}
	if u.tcpConn != nil {
		e = u.tcpConn.Close()
	}
	return e
}

func (u *udpconn) LocalAddr() net.Addr {
	return u.local
}

func (u *udpconn) RemoteAddr() net.Addr {
	return u.remote
}

func (u *udpconn) SetDeadline(t time.Time) error {
	return u.udpConn.SetDeadline(t)
}

func (u *udpconn) SetReadDeadline(t time.Time) error {
	return u.udpConn.SetReadDeadline(t)
}

func (u *udpconn) SetWriteDeadline(t time.Time) error {
	return u.udpConn.SetWriteDeadline(t)
}
