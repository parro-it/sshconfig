package sshconfig

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

// SSHHost defines a single host entry in a ssh config
type SSHHost struct {
	Host              []string
	HostName          string
	User              string
	Port              int
	ProxyCommand      string
	HostKeyAlgorithms string
	IdentityFile      string
	LocalForwards     []Forward
	RemoteForwards    []Forward
	DynamicForwards   []DForward
}

// Forward defines a single port forward entry
type Forward struct {
	InHost  string
	InPort  int
	OutHost string
	OutPort int
}

// NewForward returns Forward object parsed from LocalForward or RemoteForward string
func NewForward(f string) (Forward, error) {
	r := regexp.MustCompile(`((\S+):)?(\d+)\s+(\S+):(\d+)`)
	m := r.FindStringSubmatch(f)

	InPort, err := strconv.Atoi(m[3])
	if err != nil {
		return Forward{}, err
	}

	OutPort, err := strconv.Atoi(m[5])
	if err != nil {
		return Forward{}, err
	}

	return Forward{
		InHost:  m[2],
		InPort:  InPort,
		OutHost: m[4],
		OutPort: OutPort,
	}, nil
}

// DForward defines a single dynamic port forward entry
type DForward struct {
	Host string
	Port int
}

// NewDForward returns DForward object parsed from DynamicForward string
func NewDForward(f string) (DForward, error) {
	r := regexp.MustCompile(`((\S+):)?(\d+)`)
	m := r.FindStringSubmatch(f)

	InPort, err := strconv.Atoi(m[3])
	if err != nil {
		return DForward{}, err
	}

	return DForward{
		Host: m[2],
		Port: InPort,
	}, nil
}

// MustParseSSHConfig must parse the SSH config given by path or it will panic
func MustParseSSHConfig(path string) []*SSHHost {
	config, err := ParseSSHConfig(path)
	if err != nil {
		panic(err)
	}
	return config
}

// ParseSSHConfig parses a SSH config given by path.
func ParseSSHConfig(path string) ([]*SSHHost, error) {
	// read config file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parse(string(content))
}

// parses an openssh config file
func parse(input string) ([]*SSHHost, error) {
	sshConfigs := []*SSHHost{}
	var next item
	var sshHost *SSHHost

	lexer := lex(input)
Loop:
	for {
		token := lexer.nextItem()

		if sshHost == nil && token.typ != itemHost {
			return nil, fmt.Errorf("config variable before Host variable")
		}

		switch token.typ {
		case itemHost:
			if sshHost != nil {
				sshConfigs = append(sshConfigs, sshHost)
			}

			sshHost = &SSHHost{Host: []string{}, Port: 22}
		case itemHostValue:
			sshHost.Host = strings.Split(token.val, " ")
		case itemHostName:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.HostName = next.val
		case itemUser:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.User = next.val
		case itemPort:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			port, err := strconv.Atoi(next.val)
			if err != nil {
				return nil, err
			}
			sshHost.Port = port
		case itemProxyCommand:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.ProxyCommand = next.val
		case itemHostKeyAlgorithms:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.HostKeyAlgorithms = next.val
		case itemIdentityFile:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			sshHost.IdentityFile = next.val
		case itemLocalForward:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			f, err := NewForward(next.val)
			if err != nil {
				return nil, err
			}
			sshHost.LocalForwards = append(sshHost.LocalForwards, f)
		case itemRemoteForward:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			f, err := NewForward(next.val)
			if err != nil {
				return nil, err
			}
			sshHost.RemoteForwards = append(sshHost.RemoteForwards, f)
		case itemDynamicForward:
			next = lexer.nextItem()
			if next.typ != itemValue {
				return nil, fmt.Errorf(next.val)
			}
			f, err := NewDForward(next.val)
			if err != nil {
				return nil, err
			}
			sshHost.DynamicForwards = append(sshHost.DynamicForwards, f)
		case itemError:
			return nil, fmt.Errorf("%s at pos %d", token.val, token.pos)
		case itemEOF:
			if sshHost != nil {
				sshConfigs = append(sshConfigs, sshHost)
			}
			break Loop
		default:
			// continue onwards
		}
	}
	return sshConfigs, nil
}
