package host

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

func decToInt(n string) int {
	d, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		panic(fmt.Errorf("error while parsing %s to int: %s", n, err))
	}
	return int(d)
}

func hexToInt(h string) uint {
	d, err := strconv.ParseUint(h, 16, 64)
	if err != nil {
		panic(fmt.Errorf("error while parsing %s to int: %s", h, err))
	}
	return uint(d)
}

func hexToInt2(h string) (uint, uint) {
	if len(h) > 16 {
		d, err := strconv.ParseUint(h[:16], 16, 64)
		if err != nil {
			panic(fmt.Errorf("error while parsing %s to int: %s", h[16:], err))
		}
		d2, err := strconv.ParseUint(h[16:], 16, 64)
		if err != nil {
			panic(fmt.Errorf("error while parsing %s to int: %s", h[16:], err))
		}
		return uint(d), uint(d2)
	}

	d, err := strconv.ParseUint(h, 16, 64)
	if err != nil {
		panic(fmt.Errorf("error while parsing %s to int: %s", h[16:], err))
	}
	return uint(d), 0
}

func hexToIP(h string) net.IP {
	n, m := hexToInt2(h)
	var ip net.IP
	if m != 0 {
		ip = make(net.IP, 16)
		// TODO: Check if this depends on machine endianness?
		binary.LittleEndian.PutUint32(ip, uint32(n>>32))
		binary.LittleEndian.PutUint32(ip[4:], uint32(n))
		binary.LittleEndian.PutUint32(ip[8:], uint32(m>>32))
		binary.LittleEndian.PutUint32(ip[12:], uint32(m))
	} else {
		ip = make(net.IP, 4)
		binary.LittleEndian.PutUint32(ip, uint32(n))
	}
	return ip
}
