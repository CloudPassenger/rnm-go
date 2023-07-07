package config

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CloudPassenger/rnm-go/infra/lru"
)

type Config struct {
	ConfPath string `json:"-"`
	// HttpClient *http.Client `json:"-"`
	Groups []Group `json:"groups"`
}

type Server struct {
	Name        string `json:"name"`
	Target      string `json:"target"`
	XVer        int    `json:"xver"`
	TCPFastOpen bool   `json:"TCPFastOpen"`
	PassKey     string `json:"privKey"`
	PrivateKey  []byte `json:"-"`
}

type Group struct {
	Name                string           `json:"name"`
	Port                int              `json:"port"`
	ListenerTCPFastOpen bool             `json:"listenerTCPFastOpen"`
	AcceptProxyProtocol bool             `json:"acceptProxyProtocol"`
	Servers             []Server         `json:"servers"`
	FallbackServer      string           `json:"fallback"`
	UserContextPool     *UserContextPool `json:"-"`

	// AuthTimeoutSec sets a TCP read timeout to drop connections that fail to finish auth in time.
	// Default: no timeout
	// outline-ss-server uses 59s, which is claimed to be the most common timeout for servers that do not respond to invalid requests.
	AuthTimeoutSec int `json:"authTimeoutSec"`

	// DialTimeoutSec sets the connect timeout when dialing the target server.
	// Default: no timeout (respect system default behavior)
	// Set to a value greater than zero to override the platform's default behavior.
	DialTimeoutSec int `json:"dialTimeoutSec"`

	// DrainOnAuthFail controls whether to fallback to the first server in the group when authentication fails.
	// Default: fallback to 1st server
	// Set to true to drain the connection when authentication fails.
	DrainOnAuthFail bool `json:"drainOnAuthFail"`
}

type UpstreamConf struct {
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Settings     json.RawMessage `json:"settings"`
	PullingError error           `json:"-"`
	Upstream     Upstream        `json:"-"`
}

func (uc UpstreamConf) Equal(that UpstreamConf) bool {
	return uc.Name == that.Name && uc.Type == that.Type && uc.Upstream.Equal(that.Upstream)
}

const (
	LRUTimeout = 30 * time.Minute
)

var (
	config  *Config
	Version = "v0.2.3"
)

func (g *Group) BuildPrivateKeys() {
	servers := g.Servers
	for j := range servers {
		s := &servers[j]
		// s.MasterKey = cipher.EVPBytesToKey(s.Password, cipher.CiphersConf[s.Method].KeyLen)
		s.PrivateKey, _ = base64.RawURLEncoding.DecodeString(s.PassKey)
	}
}

func (g *Group) BuildUserContextPool(timeout time.Duration) {
	g.UserContextPool = (*UserContextPool)(lru.New(lru.FixedTimeout, int64(timeout)))
}

func (config *Config) CheckPrivkeyLength() error {
	for _, g := range config.Groups {
		for _, s := range g.Servers {
			privKey, err := base64.RawURLEncoding.DecodeString(s.PassKey)
			if err != nil {
				return err
			}
			if len(privKey) != 32 {
				return fmt.Errorf("invalid private key in server: %s", s.Name)
			}
		}
	}
	return nil
}

func (config *Config) CheckDuplicatedPrivKey() error {
	for _, g := range config.Groups {
		m := make(map[string]struct{})
		for _, s := range g.Servers {
			mp := s.PassKey
			if _, exists := m[mp]; exists {
				return fmt.Errorf("make sure the privKey in the same group are diverse: %s", s.PassKey)
			}
		}
	}
	return nil
}

func (config *Config) CheckXver() error {
	for _, g := range config.Groups {
		for _, s := range g.Servers {
			xver := s.XVer
			if xver > 0 {
				if xver != 1 && xver != 2 {
					return fmt.Errorf("make sure the proxyprotocol version is correct: %s", s.Name)
				}
			}
		}
	}
	return nil
}

func check(config *Config) (err error) {
	if err = config.CheckXver(); err != nil {
		return
	}
	if err = config.CheckPrivkeyLength(); err != nil {
		return
	}
	if err = config.CheckDuplicatedPrivKey(); err != nil {
		return
	}
	return
}

func build(config *Config) {
	for i := range config.Groups {
		g := &config.Groups[i]
		g.BuildUserContextPool(LRUTimeout)
		g.BuildPrivateKeys()
	}
}

func BuildConfig(confPath string) (conf *Config, err error) {
	conf = new(Config)
	conf.ConfPath = confPath
	// conf.HttpClient = c
	b, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, conf); err != nil {
		return nil, err
	}
	if err = check(conf); err != nil {
		return nil, err
	}
	build(conf)
	return
}

func SetConfig(conf *Config) {
	config = conf
}

func NewConfig() *Config {
	var err error

	version := flag.Bool("v", false, "version")
	confPath := flag.String("conf", "config.json", "config file path")
	suppressTimestamps := flag.Bool("suppress-timestamps", false, "do not include timestamps in log")
	flag.Parse()

	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	if *suppressTimestamps {
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	}

	if config, err = BuildConfig(*confPath); err != nil {
		log.Fatalln(err)
	}
	return config
}
