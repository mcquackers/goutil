package goutil

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
	"gopkg.in/alexcesaro/statsd.v2"
)

type InternalErrorer interface {
	InternalError()
}

type Pulser interface {
	Pulse()
}

type Heart struct {
	Pulser
	env     string
	subenv  string
	address string
	prefix  string
}

func NewHeartBeat(address, prefix string) *Heart {
	return &Heart{
		env:     os.Getenv("ENVIRONMENT"),
		subenv:  os.Getenv("SUBENVIRONMENT"),
		address: address,
		prefix:  prefix,
	}
}

type UtilKit struct {
	InternalErrorer
	*Heart
	Address string
	Prefix  string
	Env     string
	Subenv  string
}

func NewUtilKit(address string, prefix string) *UtilKit {
	utilKit := &UtilKit{
		Address: address,
		Prefix:  prefix,
		Env:     os.Getenv("ENVIRONMENT"),
		Subenv:  os.Getenv("SUBENVIRONMENT"),
	}

	utilKit.Heart = NewHeartBeat(address, prefix)

	return utilKit
}

func Un(origin string) {
	glog.V(2).Infof("Leaving %s\n", origin)
}

func Trace(origin string) string {
	glog.V(2).Infof("Entering %s\n", origin)
	return origin
}

func (this *UtilKit) InternalError() {
	prefix := fmt.Sprintf("%s.%s.%s", this.Prefix, this.Env, this.Subenv)
	c, err := statsd.New(statsd.Prefix(prefix), statsd.Address(this.Address))
	if err != nil {
		glog.Errorf("[SERVER#internalError] Statsd creation error:\n%v\n", err)
	}
	defer c.Close()

	c.Increment("500_count")
}

func (this *Heart) Pulse() {
	prefix := fmt.Sprintf("%s.%s.%s", this.prefix, this.env, this.subenv)
	c, err := statsd.New(statsd.Prefix(prefix), statsd.Address(this.address))
	if err != nil {
		glog.Errorf("[SERVER#pulse] Statsd creation error:\n%v\n", err)
	}
	ticker := time.NewTicker(time.Second * 10)
	for _ = range ticker.C {
		c.Gauge("pulse", 1)
	}
}

func HandleStatusCode(code int) error {
	switch code {
	case 200, 201, 203:
		return nil
	case 400:
		return errors.New("400 Bad Request")
	case 401:
		return errors.New("401 Unauthorised")
	case 403:
		return errors.New("403 Forbidden")
	case 404:
		return errors.New("404 Not found")
	default:
		return errors.New(strconv.Itoa(code) + " Other error")
	}
}
