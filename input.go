package main

import (
	"errors"
	"fmt"
	"time"
)

type Input struct {
	TestCase       string
	Goroutines     int
	From           int
	To             int
	BlockSize      int
	ReplicationMin int
	ReplicationMax int
	HostPort       string
}

func (o *Input) info(ts bool) string {
	ti := fmt.Sprintf("%s%d-%d_g%d", o.TestCase, o.From, o.To, o.Goroutines)

	if o.BlockSize > 0 {
		ti += fmt.Sprintf("_bs%d", o.BlockSize)
	}

	if o.ReplicationMin > 0 && o.ReplicationMax > 0 {
		ti += fmt.Sprintf("_r%d-%d", o.ReplicationMin, o.ReplicationMax)
	}

	if ts {
		ti += fmt.Sprintf("_%d", time.Now().Unix())
	}

	return ti
}

func (o *Input) check() error {
	if o.From >= o.To {
		return errors.New(fmt.Sprintf("wrong [from, to), from:%d, to:%d", o.From, o.To))
	}

	if (o.To - o.From) < o.Goroutines {
		return errors.New(fmt.Sprintf("goroutines must <= (to-from), check the parameters"))
	}

	return nil
}
