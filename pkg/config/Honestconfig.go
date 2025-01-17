package config

import (
	"encoding/base64"
	"fmt"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign/bls"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Implement Config interface in local linux machine setting
type HonestConfig struct {
	N int `yaml:"N"` //每个分片中的节点数
	F int `yaml:"F"` //每个分片中的恶意节点数
	M int `yaml:"m"` //分片个数

	IPList   []string `yaml:"IPList"`
	PortList []string `yaml:"PortList"`
	Txnum    int      `yaml:"Txnum"`
	Crate    float64  `yaml:"Crate"`
	// judge if execute read config function before
	// default is false in golang structure declare
	isRead    bool
	PID       int      `yaml:"PID"`  //节点在整体中的编号
	Snumber   int      `yaml:"Snum"` //节点所在的分片编号
	SID       int      `yaml:"SID"`  //节点在分片内的编号
	Statistic string   `yaml:"Statistic"`
	PK        []string `yaml:"PK"`
	SK        string   `yaml:"SK"`
	// server start time
	PrepareTime int `yaml:"PrepareTime"`
	WaitTime    int `yaml:"WaitTime"`

	TestEpochs int `yaml:"TestEpochs"`
}

func NewHonestConfig(configName string, isLocal bool) (HonestConfig, error) {
	c := HonestConfig{}
	err := c.ReadHonestConfig(configName, isLocal)
	if err != nil {
		return HonestConfig{}, err
	}
	return c, err
}

// read config from ConfigName file location
func (c *HonestConfig) ReadHonestConfig(ConfigName string, isLocal bool) error {
	byt, err := ioutil.ReadFile(ConfigName)
	if err != nil {
		goto ret
	}

	err = yaml.Unmarshal(byt, c)

	c.isRead = true

	if !isLocal {
		if err != nil {
			goto ret
		}

		if c.N <= 0 || c.F < 0 {
			return errors.Wrap(errors.New("N or F is negative"),
				ConfigReadError.Error())
		}

		if c.N != len(c.IPList) || c.N != len(c.PortList) {
			return errors.Wrap(errors.New("ip list"+
				" length or port list length isn't match N"),
				ConfigReadError.Error())
		}
		// id is begin from 0 to ... N-1
		if c.PID >= c.N || c.PID < 0 {
			return errors.New("ID is begin from 0 to N-1")
		}
	}

	return nil
ret:
	return errors.Wrap(err, ConfigReadError.Error())
}

// Achieve numbers of total nodes
// the return value is a positive integer
func (c *HonestConfig) GetN() (int, error) {
	if !c.isRead {
		return 0, NotReadFileError
	}
	return c.N, nil
}

// Achieve number of corrupted nodes
// return value is a positive integer
func (c *HonestConfig) GetF() (int, error) {
	if !c.isRead {
		return 0, NotReadFileError
	}
	return c.F, nil
}

// Achieve ip list if defined
// return a ip list of defined ip in config file
func (c *HonestConfig) GetIPList() ([]string, error) {
	if !c.isRead {
		return nil, NotReadFileError
	}
	if len(c.IPList) == 0 {
		return nil, NotDefined
	}
	return c.IPList, nil
}

// Achieve port list if defined
// return a port list of defined port in config file
func (c *HonestConfig) GetPortList() ([]string, error) {
	if !c.isRead {
		return nil, NotReadFileError
	}
	if len(c.PortList) == 0 {
		return nil, NotDefined
	}
	return c.PortList, nil
}

func (c *HonestConfig) GetMyID() (int, error) {
	if !c.isRead {
		return 0, NotReadFileError
	}
	return c.PID, nil
}

func (c *HonestConfig) Marshal(location string) error {
	byts, err := yaml.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "marshal config fail")
	}
	err = ioutil.WriteFile(location, byts, 0777)
	if err != nil {
		return errors.Wrap(err, "marshal config fail")
	}
	return nil
}

func (c *HonestConfig) RemoteHonestGen(dir string) error {
	//suite := bn256.NewSuite()
	suite := pairing.NewSuiteBn256()
	randomStream := suite.RandomStream()
	var pks []string
	var sks []string

	for i := 0; i < c.N*c.M; i++ {
		sk, pk := bls.NewKeyPair(suite, randomStream)
		skBytes, _ := sk.MarshalBinary()
		pkBytes, _ := pk.MarshalBinary()
		sks = append(sks, base64.StdEncoding.EncodeToString(skBytes))
		pks = append(pks, base64.StdEncoding.EncodeToString(pkBytes))
	}

	for i := 0; i < c.N*c.M; i++ {
		c.PID = i
		c.SID = i % c.N
		c.Snumber = i / c.N

		c.SK = sks[i]
		c.PK = pks

		err := c.Marshal(dir + "/config_" + strconv.Itoa(i) + ".yaml")
		if err != nil {
			fmt.Println(dir)
			fmt.Println("marshal config fail")
			return errors.Wrap(err, "marshal config fail")
		}
	}
	return nil
}
