package network

import (
	"net"
	"testing"
)

func TestAllocate(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("192.0.0.0/16")
	ip, _ := ipAllocator.Allocate(ipnet)
	t.Logf("alloc ip: %v", ip)
	//_, ipnet2, _ := net.ParseCIDR("192.0.0.0/16")
	//ip2, _ := ipAllocator.Allocate(ipnet2)
	//t.Logf("alloc ip2: %v", ip2)
}

func TestRelease(t *testing.T) {
	//ip, ipnet, _ := net.ParseCIDR("192.168.0.0/16")
	//ipAllocator.Release(ipnet, &ip)
}
