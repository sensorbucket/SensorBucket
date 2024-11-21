package service

import (
	"bytes"
	"context"
	"log"
	"time"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

// MQTTProcessor is mochi-mqtt hook that authenticates mqtt clients using the username and password as
// pipelineID and APIKey respectively. The hook builds an IngressDTO for each publish, and forwards this to the
// AMQP Message Queue.
type MQTTProcessor struct {
	mqtt.HookBase
	ctx       context.Context
	clients   *ClientRegistry
	publisher chan<- processing.IngressDTO
}

type MQTTProcessorOptions struct {
	Context      context.Context
	APIKeyTrader APIKeyTrader
	Publisher    chan<- processing.IngressDTO
}

func (h *MQTTProcessor) Init(_opts any) error {
	if _, ok := _opts.(*MQTTProcessorOptions); !ok && _opts != nil {
		return mqtt.ErrInvalidConfigType
	}
	opts := &MQTTProcessorOptions{}
	if _opts != nil {
		opts = _opts.(*MQTTProcessorOptions)
	}

	h.ctx = opts.Context
	h.clients = CreateClientRegistry(h.ctx, opts.APIKeyTrader, 5*time.Minute)
	h.publisher = opts.Publisher

	return nil
}

func (h *MQTTProcessor) ID() string {
	return "auth-sensorbucket"
}

func (h *MQTTProcessor) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnectAuthenticate,
		mqtt.OnClientExpired,
		mqtt.OnPublish,
	}, []byte{b})
}

func (h *MQTTProcessor) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	cntClientAuth.Add(context.Background(), 1)
	if err := h.clients.Authenticate(cl.ID, string(cl.Properties.Username), string(pk.Connect.Password)); err != nil {
		log.Printf("Error authenticating APIKey: %s\n", err.Error())
		return false
	}
	cntClientAuthSuccess.Add(context.Background(), 1)
	return true
}

func (h *MQTTProcessor) OnClientExpired(cl *mqtt.Client) {
	h.clients.Destroy(cl.ID)
}

type DTOMetadata struct {
	Topic string `json:"topic"`
}

func (h *MQTTProcessor) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	cntMQTTPublishes.Add(context.Background(), 1)
	defer timer(histProcessDuration)()
	pk.Ignore = true

	client, err := h.clients.GetClient(cl.ID)
	if err != nil {
		return pk, err
	}

	dto := processing.CreateIngressDTO(client.AccessToken, client.PipelineID, client.TenantID, pk.Payload)
	dto.Metadata["mqtt"] = DTOMetadata{
		Topic: pk.TopicName,
	}
	dto.CreatedAt = time.Unix(pk.Created, 0)
	h.publisher <- dto

	return pk, nil
}
