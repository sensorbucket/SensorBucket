package service

import (
	"bytes"
	"context"
	"log"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

type Auther struct {
	mqtt.HookBase
	ctx       context.Context
	clients   *ClientKeySets
	publisher chan<- processing.IngressDTO
}

type AuthHookOptions struct {
	Context      context.Context
	APIKeyTrader APIKeyTrader
	Publisher    chan<- processing.IngressDTO
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
	h.clients = CreateClientKeySets(h.ctx, opts.APIKeyTrader)
	h.publisher = opts.Publisher

	return nil
}

func (h *Auther) ID() string {
	return "auth-sensorbucket"
}

func (h *Auther) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnectAuthenticate,
		mqtt.OnClientExpired,
		mqtt.OnPublish,
	}, []byte{b})
}

func (h *Auther) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	if err := h.clients.Authenticate(cl.ID, string(cl.Properties.Username), string(pk.Connect.Password)); err != nil {
		log.Printf("Error authenticating APIKey: %s\n", err.Error())
		return false
	}
	return true
}

func (h *Auther) OnClientExpired(cl *mqtt.Client) {
	h.clients.Destroy(cl.ID)
}

type IngressPayload struct {
	Topic string
}

func (h *Auther) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	pk.Ignore = true

	client, err := h.clients.GetClient(cl.ID)
	if err != nil {
		return packets.Packet{}, err
	}

	h.publisher <- processing.CreateIngressDTO(client.APIKey, client.PipelineID, client.TenantID, pk.Payload)
	// fmt.Printf("dto: %v\n", dto)

	return pk, nil
}
