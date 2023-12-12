//go:build linux

/*
 * Copyright (C) 2020-2022, IrineSistiana
 *
 * This file is part of mosdns.
 *
 * mosdns is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * mosdns is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package iptoshell

import (
        "context"
        "fmt"
        "github.com/IrineSistiana/mosdns/v5/pkg/query_context"
        "github.com/miekg/dns"
        "github.com/nadoo/ipset"
        "net/netip"
        "os/exec"
        "strconv"
)

type iptoshellPlugin struct {
        args *Args
        nl   *ipset.NetLink
}

func newIpToshellPlugin(args *Args) (*iptoshellPlugin, error) {
        if args.Mask4 == 0 {
                args.Mask4 = 24
        }
        if args.Mask6 == 0 {
                args.Mask6 = 32
        }

        return &iptoshellPlugin{
                args: args,
                nl:   nl,
        }, nil
}

func (p *iptoshellPlugin) Exec(_ context.Context, qCtx *query_context.Context) error {
        r := qCtx.R()
        if r != nil {
                if err := p.addIPtoshell(r); err != nil {
                        return fmt.Errorf("iptoshell: %w", err)
                }
        }
        return nil
}

func (p *iptoshellPlugin) Close() error {
        return p.nl.Close()
}

func (p *iptoshellPlugin) addIPtoshell(r *dns.Msg) error {
        for i := range r.Answer {
                switch rr := r.Answer[i].(type) {
                case *dns.A:
                        if len(p.args.SetName4) == 0 {
                                continue
                        }
                        addr, ok := netip.AddrFromSlice(rr.A.To4())
                        if !ok {
                                return fmt.Errorf("iptoshell invalid A record with ip: %s", rr.A)
                        }
                        var okprefix error
                        prefix, okprefix := addr.Prefix(p.args.Mask4)
                        if okprefix != nil {
                                return fmt.Errorf("iptoshell to Prefix  invalid A record with ip: %s", rr.A)
                        }  
                        cmd := exec.Command(p.args.SetName4, rr.A.String(), strconv.Itoa(p.args.Mask4), prefix.String())
                        err := cmd.Run()
                        if err != nil {
                              return fmt.Errorf(" RUN iptoshell invalid  A record with ip: %s", rr.A)
                        } 


                case *dns.AAAA:
                        if len(p.args.SetName6) == 0 {
                                continue
                        }
                        addr, ok := netip.AddrFromSlice(rr.AAAA.To16())
                        if !ok {
                                return fmt.Errorf("iptoshell invalid AAAA record with ip: %s", rr.AAAA)
                        }
                        var okprefix error
                        prefix, okprefix := addr.Prefix(p.args.Mask6)
                        if okprefix != nil {
                                return fmt.Errorf("iptoshell to Prefix  invalid AAAA record with ip: %s", rr.AAAA)
                        }
                        cmd := exec.Command(p.args.SetName6, rr.AAAA.String(), strconv.Itoa(p.args.Mask6), prefix.String())
                        err := cmd.Run()
                        if err != nil {
                             return fmt.Errorf("Run iptoshell AAAA record with ip: %s", rr.AAAA)
                        } 
                default:
                        continue
                }
        }

        return nil
}
