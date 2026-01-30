package mqtt

import (
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client  mqtt.Client
	builder *MQTTBuilder

	builderLock sync.Mutex
}

// Publish sends a message to the specified topic using the publication spec identified by operationID.
// It does not validate the topic or payload.
func (c *MQTTClient) Publish(operationID string, actualTopic string, payload any) error {
	c.builderLock.Lock()
	pub, ok := c.builder.publications[operationID]
	c.builderLock.Unlock()
	if !ok {
		return fmt.Errorf("publication not found for operationID %s", operationID)
	}

	bytes, err := utils.ToJSON(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %w", err)
	}

	token := c.client.Publish(actualTopic, byte(pub.QoS), pub.Retained, bytes)
	token.Wait()

	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", actualTopic, err)
	}

	return nil
}
