package service

import (
	"bytes"
	"context"
	"log"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type Auther struct {
	mqtt.HookBase
	ctx        context.Context
	clientKeys *ClientKeySets
}

type AuthHookOptions struct {
	Context      context.Context
	APIKeyTrader APIKeyTrader
}

func (h *Auther) Init(_opts any) error {
	if _, ok := _opts.(*AuthHookOptions); !ok && _opts != nil {
		return mqtt.ErrInvalidConfigType
	}
	opts := &AuthHookOptions{}
	if _opts != nil {
		opts = _opts.(*AuthHookOptions)
	}

	h.ctx = opts.Context
	h.clientKeys = CreateClientKeySets(h.ctx, opts.APIKeyTrader)

	return nil
}

func (h *Auther) ID() string {
	return "auth-sensorbucket"
}

func (h *Auther) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnectAuthenticate,
		mqtt.OnClientExpired,
	}, []byte{b})
}

func (h *Auther) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	if err := h.clientKeys.Authenticate(cl.ID, string(pk.Connect.Password)); err != nil {
		log.Printf("Error authenticating APIKey: %s\n", err.Error())
		return false
	}
	return true
}

func (h *Auther) OnClientExpired(cl *mqtt.Client) {
	h.clientKeys.Destroy(cl.ID)
}
