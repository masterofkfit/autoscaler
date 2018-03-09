// Copyright 2018 Drone.IO Inc
// Use of this software is governed by the Business Source License
// that can be found in the LICENSE file.

package hetznercloud

import (
	"bytes"
	"context"
	"strconv"

	"github.com/drone/autoscaler"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/rs/zerolog/log"
)

func (p *provider) Create(ctx context.Context, opts autoscaler.InstanceCreateOpts) (*autoscaler.Instance, error) {
	p.init.Do(func() {
		p.setup(ctx)
	})

	buf := new(bytes.Buffer)
	err := p.userdata.Execute(buf, &opts)
	if err != nil {
		return nil, err
	}

	req := hcloud.ServerCreateOpts{
		Name:     opts.Name,
		UserData: buf.String(),
		ServerType: &hcloud.ServerType{
			Name: p.serverType,
		},
		Image: &hcloud.Image{
			Name: p.image,
		},
		Datacenter: &hcloud.Datacenter{
			Name: p.datacenter,
		},
		SSHKeys: []*hcloud.SSHKey{
			{
				ID: p.key,
			},
		},
	}

	logger := log.Ctx(ctx).With().
		Str("datacenter", req.Datacenter.Name).
		Str("image", req.Image.Name).
		Str("serverType", req.ServerType.Name).
		Str("name", req.Name).
		Logger()

	logger.Debug().
		Msg("instance create")

	resp, _, err := p.client.Server.Create(ctx, req)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("cannot create instance")
		return nil, err
	}

	logger.Info().
		Str("name", req.Name).
		Msg("instance created")

	return &autoscaler.Instance{
		Provider: autoscaler.ProviderHetznerCloud,
		ID:       strconv.Itoa(resp.Server.ID),
		Name:     resp.Server.Name,
		Address:  resp.Server.PublicNet.IPv4.IP.String(),
		Size:     req.ServerType.Name,
		Region:   req.Datacenter.Name,
		Image:    req.Image.Name,
	}, nil
}
