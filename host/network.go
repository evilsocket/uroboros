package host

import (
	"bufio"
	"fmt"
	"github.com/evilsocket/islazy/str"
	"net"
	"os"
	"regexp"
)

type NetworkINodes map[int]NetworkEntry

// https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tree/include/net/tcp_states.h
const (
	TCP_ESTABLISHED = iota + 1
	TCP_SYN_SENT
	TCP_SYN_RECV
	TCP_FIN_WAIT1
	TCP_FIN_WAIT2
	TCP_TIME_WAIT
	TCP_CLOSE
	TCP_CLOSE_WAIT
	TCP_LAST_ACK
	TCP_LISTEN
	TCP_CLOSING /* Now a valid state */
	TCP_NEW_SYN_RECV
)

var sockStates = map[uint]string{
	TCP_ESTABLISHED:  "ESTABLISHED",
	TCP_SYN_SENT:     "SYN_SENT",
	TCP_SYN_RECV:     "SYN_RECV",
	TCP_FIN_WAIT1:    "FIN_WAIT1",
	TCP_FIN_WAIT2:    "FIN_WAIT2",
	TCP_TIME_WAIT:    "TIME_WAIT",
	TCP_CLOSE:        "CLOSE",
	TCP_CLOSE_WAIT:   "CLOSE_WAIT",
	TCP_LAST_ACK:     "LAST_ACK",
	TCP_LISTEN:       "LISTENING",
	TCP_CLOSING:      "CLOSING",
	TCP_NEW_SYN_RECV: "NEW_SYN_RECV",
}

var sockTypes = map[uint]string{
	1: "SOCK_STREAM",
	2: "SOCK_DGRAM",
	5: "SOCK_SEQPACKET",
}

// Entry holds the information of a /proc/net/* entry.
// For example, /proc/net/tcp:
// sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
// 0:  0100007F:13AD 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 18083222
// or /proc/net/unix
// Num       RefCount Protocol Flags    Type St Inode Path
// 0000000000000000: 00000002 00000000 00010000 0005 01 15878 /run/udev/control
// 0000000000000000: 00000002 00000000 00000000 0002 01 28261 /run/user/1000/systemd/notify
// 0000000000000000: 00000002 00000000 00010000 0001 01 28264 /run/user/1000/systemd/private
// 0000000000000000: 00000002 00000000 00010000 0001 01 28270 /run/user/1000/bus
// 0000000000000000: 00000002 00000000 00010000 0001 01 28271 /run/user/1000/gnupg/S.dirmngr
// 0000000000000000: 00000002 00000000 00010000 0001 01 28272 /run/user/1000/gnupg/S.gpg-agent.browser
// 0000000000000000: 00000002 00000000 00010000 0001 01 28273 /run/user/1000/gnupg/S.gpg-agent.extra
type NetworkEntry struct {
	Proto string
	// for unix
	Type       uint
	TypeString string
	Path       string
	// for everythign else
	State       uint
	StateString string
	SrcIP       net.IP
	SrcPort     uint
	DstIP       net.IP
	DstPort     uint
	UserId      int
	INode       int
}

func (e NetworkEntry) String() string {
	if e.Proto == "unix" {
		return fmt.Sprintf("(%s) %s path='%s'", e.Proto, e.TypeString, e.Path)
	} else if e.State == TCP_LISTEN {
		return fmt.Sprintf("(%s) %s:%d", e.Proto, e.SrcIP, e.SrcPort)
	}
	return fmt.Sprintf("(%s) %s:%d <-> %s:%d", e.Proto, e.SrcIP, e.SrcPort, e.DstIP, e.DstPort)
}

func (e NetworkEntry) InfoString() string {
	if e.Proto == "unix" {
		if e.Path != "" {
			if info, err := os.Stat(e.Path); err == nil {
				return info.Mode().String()
			}
		}
	} else {
		return e.StateString
	}
	return ""
}

var (
	unixParser = regexp.MustCompile(`(?i)` +
		`[a-f0-9]+:\s+` + // num
		`[a-f0-9]+\s+` + // ref count
		`[a-f0-9]+\s+` + // protocol
		`[a-f0-9]+\s+` + // flags
		`([a-f0-9]+)\s+` + // type
		`([a-f0-9]+)\s+` + // state
		`(\d+)\s*` + // inode
		`(.*)`, // path
	)

	tcpAndUdpParser = regexp.MustCompile(`(?i)` +
		`\d+:\s+` + // number of entry
		`([a-f0-9]{8,32}):([a-f0-9]{4})\s+` + // local_address
		`([a-f0-9]{8,32}):([a-f0-9]{4})\s+` + // rem_address
		`([a-f0-9]{2})\s+` + // connection state
		`[a-f0-9]{8}:[a-f0-9]{8}\s+` + // tx_queue rx_queue
		`[a-f0-9]{2}:[a-f0-9]{8}\s+` + // tr tm->when
		`[a-f0-9]{8}\s+` + // retrnsmt
		`(\d+)\s+` + // uid
		`\d+\s+` + // timeout
		`(\d+)\s+` + // inode
		`.+`) // stuff we don't care about
)

// Parse scans and retrieves the opened connections, from /proc/net/ files
func parseNetworkForProtocol(proto string) ([]NetworkEntry, error) {
	filename := fmt.Sprintf("%s/net/%s", ProcFS, proto)
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	entries := make([]NetworkEntry, 0)
	scanner := bufio.NewScanner(fd)
	for lineno := 0; scanner.Scan(); lineno++ {
		// skip column names
		if lineno == 0 {
			continue
		}

		var entry NetworkEntry
		line := str.Trim(scanner.Text())
		if proto == "unix" {
			m := unixParser.FindStringSubmatch(line)
			if m == nil {
				panic(fmt.Errorf("could not parse netstat line from %s: %s", filename, line))
				continue
			}
			entry = NetworkEntry{
				Proto: proto,
				Type:  hexToInt(m[1]),
				State: hexToInt(m[2]),
				INode: decToInt(m[3]),
				Path:  m[4],
			}
			entry.TypeString = sockTypes[entry.Type]
			entry.StateString = sockStates[entry.State]
		} else {
			m := tcpAndUdpParser.FindStringSubmatch(line)
			if m == nil {
				panic(fmt.Errorf("could not parse netstat line from %s: %s", filename, line))
				continue
			}

			entry = NetworkEntry{
				Proto:   proto,
				SrcIP:   hexToIP(m[1]),
				SrcPort: hexToInt(m[2]),
				DstIP:   hexToIP(m[3]),
				DstPort: hexToInt(m[4]),
				State:   hexToInt(m[5]),
				UserId:  decToInt(m[6]),
				INode:   decToInt(m[7]),
			}
			entry.StateString = sockStates[entry.State]
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func buildNetworkINodes() (NetworkINodes, error) {
	byInode := make(NetworkINodes)
	protos := []string{"tcp", "tcp6", "udp", "udp6", "unix"}

	for i := range protos {
		if entries, err := parseNetworkForProtocol(protos[i]); err != nil {
			return nil, err
		} else {
			for _, entry := range entries {
				byInode[entry.INode] = entry
			}
		}
	}
	return byInode, nil
}
