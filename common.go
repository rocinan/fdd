package fdd

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"

	"github.com/shuLhan/share/lib/dns"
	"golang.org/x/sys/unix"
)

//socket常用操作方法封装 //todo ipv6 support
//CreateTcpListenSocket 根据地址端口创建tcp监听socket
func CreateTcpListenSocket(addr string, port int) (int, error) {
	if fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_TCP); err != nil {
		return 0, err
	} else {
		SetNoBlock(fd)
		SetReUseAddr(fd)
		socketAddr := unix.SockaddrInet4{Port: port}
		copy(socketAddr.Addr[:], net.ParseIP(addr).To4())
		if err := unix.Bind(fd, &socketAddr); err != nil {
			return 0, err
		}
		if err := unix.Listen(fd, 128); err != nil {
			return 0, err
		}
		return fd, nil
	}
}

//CreateUdpListenSocket 根据地址创建udp socket
func CreateUdpListenSocket(addr string, port int) (int, error) {
	if fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP); err != nil {
		return 0, err
	} else {
		defer SetNoBlock(fd)
		SetReUseAddr(fd)
		socketAddr := unix.SockaddrInet4{Port: port}
		copy(socketAddr.Addr[:], net.ParseIP(addr).To4())
		if err := unix.Bind(fd, &socketAddr); err != nil {
			return 0, err
		}
		return fd, nil
	}
}

//AcceptTcpConn 接受tcp链接返回链接socketFD
func AcceptTcpConn(fd int) (int, error) {
	if fd, _, err := unix.Accept(fd); err != nil {
		return 0, err
	} else {
		return fd, nil
	}
}

//CreateRemoteSocket 创建tcp连接 socketFD
func CreateRemoteSocket(remoteAddr string, remotePort int) (int, error) {
	if fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_TCP); err != nil {
		return 0, err
	} else {
		defer SetNoBlock(fd)
		socketAddr := unix.SockaddrInet4{Port: remotePort}
		copy(socketAddr.Addr[:], net.ParseIP(remoteAddr).To4())
		if err = unix.Connect(fd, &socketAddr); err != nil {
			return 0, err
		}
		return fd, nil
	}
}

//CreateRemoteSocket 创建udp socketFD
func CreateUdpRemoteSocket() (int, error) {
	if fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0); err != nil {
		return 0, err
	} else {
		defer SetNoBlock(fd)
		return fd, nil
	}
}

func SockAddrParse(addr string, port int) *unix.SockaddrInet4 {
	sa := &unix.SockaddrInet4{Port: port}
	copy(sa.Addr[:], net.ParseIP(addr).To4())
	return sa
}

func CheckError(pf string, err error) bool {
	if err != nil {
		log.Warn(pf, err)
		return false
	}
	return true
}

func BufferSend(fd int, buffer *[]byte) (int, error) {
	return unix.Write(fd, *buffer)
}

func BufferRecv(fd int, buffer *[]byte) (int, error) {
	return unix.Read(fd, *buffer)
}

func PacketSend(fd int, p *[]byte, sa *unix.SockaddrInet4) error {
	return unix.Sendto(fd, *p, 0, sa)
}

func PacketRecv(fd int, p *[]byte) (int, *unix.SockaddrInet4, error) {
	n, from, err := unix.Recvfrom(fd, *p, 0)
	return n, from.(*unix.SockaddrInet4), err
}

func SetNoBlock(fd int) error {
	return unix.SetNonblock(fd, true)
}

func SetReUseAddr(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
}

func CloseSocket(fd int) error {
	return unix.Close(fd)
}

func MD5Addr(addr []byte, port int) string {
	sum := md5.Sum(append(addr, []byte(strconv.Itoa(port))...))
	return hex.EncodeToString(sum[:])
}

func GetDomainIp(domain string) (string, error) {
	dot, err := dns.NewDoTClient("223.5.5.5:853", true)
	if err != nil {
		return "", err
	}
	defer dot.Close()
	res, err := dot.Lookup(dns.MessageQuestion{
		Name: domain,
		Type: dns.RecordTypeA,
	}, false)
	if err != nil {
		return "", err
	}
	for _, v := range res.Answer {
		if v.Type != 5 {
			return v.Value.(string), nil
		}
	}
	return "", nil
}

func Judge(v int) bool {
	return v != 0
}

func Addr2Str(sa *unix.SockaddrInet4) string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3], sa.Port)
}
