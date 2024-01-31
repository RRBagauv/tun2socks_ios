package tun2socks

import (
	"context"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
	"github.com/Jason-Stan-Lee/go-tun2socks/v2/common/log"
	"github.com/Jason-Stan-Lee/go-tun2socks/v2/core"
	"github.com/Jason-Stan-Lee/go-tun2socks/v2/proxy/v2ray"
	vcore "github.com/v2fly/v2ray-core/v5"
	vproxyman "github.com/v2fly/v2ray-core/v5/app/proxyman"
)

type PacketFlow interface {
	WritePacket(packet []byte)
}

func InputPacket(data []byte) {
	lwipStack.Write(data)
}

var lwipStack core.LWIPStack

func StartV2Ray(packetFlow PacketFlow, configBytes []byte) {
	if packetFlow == nil {
		return
	}

	lwipStack = core.NewLWIPStack()
	// v, err := vcore.StartInstance("json", configBytes)
	v, err := vcore.StartInstance("json", configBytes)
	if err != nil {
		log.Fatalf("start V instance failed: %v", err)
	}

	sniffingConfig := &vproxyman.SniffingConfig{
		Enabled:             true,
		DestinationOverride: strings.Split("tls,http", ","),
	}

	debug.SetGCPercent(5)
	ctx := vproxyman.ContextWithSniffingConfig(context.Background(), sniffingConfig)

	// Register tun2socks connection handlers.
	core.RegisterTCPConnectionHandler(v2ray.NewTCPHandler(ctx, v))
	core.RegisterUDPConnectionHandler(v2ray.NewUDPHandler(ctx, v, 30*time.Second))

	core.RegisterOutputFn(func(data []byte) (int, error) {
		// Write IP packets back to TUN.
		packetFlow.WritePacket(data)
		runtime.GC()
		debug.FreeOSMemory()
		return len(data), nil
	})
}
