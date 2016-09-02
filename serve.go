/*
 * btrfscue version 0.3
 * Copyright (c)2011-2016 Christian Blichmann
 *
 * Sub-command to provide a FS index metadata over IPC
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *     * Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"flag"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"

	"blichmann.eu/code/btrfscue/btrfs"
	"blichmann.eu/code/btrfscue/subcommand"
)

type MetadataService struct {
	Index btrfs.Index
	ch    chan os.Signal
}

func NewMetadataService() *MetadataService {
	return &MetadataService{Index: btrfs.NewIndex(),
		ch: make(chan os.Signal, 1)}
}

func Wait(s *MetadataService) os.Signal {
	signal.Notify(s.ch, os.Interrupt, os.Kill)
	return <-s.ch

}

type RangeArgs struct {
	KeyFirst, KeyLast btrfs.Key
}

type RangeReply struct {
	Low, High int
}

func (s *MetadataService) Range(args *RangeArgs, reply *RangeReply) error {
	reply.Low, reply.High = s.Index.Range(args.KeyFirst, args.KeyLast)
	return nil
}

func (s *MetadataService) Quit(args *struct{}, reply *bool) error {
	*reply = true
	go func() { s.ch <- os.Kill }()
	return nil
}

type serveCommand struct {
	port *int
}

func (c *serveCommand) DefineFlags(fs *flag.FlagSet) {
	c.port = fs.Int("port", 7077, "listen port for RPC")
}

func (c *serveCommand) Run([]string) {
	if len(*metadata) == 0 {
		fatalf("missing metadata option\n")
	}

	s := NewMetadataService()
	{
		m, err := os.Open(*metadata)
		reportError(err)
		defer m.Close()

		reportError(ReadIndex(m, &s.Index))
	}

	rpc.Register(s)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", net.JoinHostPort("::1",
		strconv.Itoa(*c.port)))
	reportError(err)
	go http.Serve(l, nil)

	sig := Wait(s)
	verbosef("Got signal: %s\n", sig)
}

func init() {
	subcommand.Commands.RegisterHidden("serve", &serveCommand{})
}
