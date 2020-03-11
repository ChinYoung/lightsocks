package tun

import (
	"errors"
	"io"
	"net"
	"os"
	"time"

	"github.com/eycorsican/go-tun2socks/core"
	"github.com/eycorsican/go-tun2socks/proxy/socks"
	"github.com/gwuhaolin/lightsocks/local"
)

const listenAddr = "127.0.0.1:0"

func StartTunServer(password string, remoteAddr string, fd int) (close func(), err error) {
	lsLocal, err := local.NewLsLocal(password, listenAddr, remoteAddr)
	if err != nil {
		return
	}
	lwipStack := core.NewLWIPStack()
	f := os.NewFile(uintptr(fd), "tun")
	if f == nil {
		return nil, errors.New("无法打开VPN虚拟网卡")
	}
	return lsLocal.Listen(func(listenAddr *net.TCPAddr) {
		// 成功启动LS服务器，开启转发
		core.RegisterTCPConnHandler(socks.NewTCPHandler(listenAddr.IP.String(), uint16(listenAddr.Port)))
		core.RegisterUDPConnHandler(socks.NewUDPHandler(listenAddr.IP.String(), uint16(listenAddr.Port), 2*time.Minute))
		core.RegisterOutputFn(func(data []byte) (int, error) {
			return f.Write(data) // 写入tun
		})
		io.Copy(lwipStack, f) // 从tun中读
	})
}
