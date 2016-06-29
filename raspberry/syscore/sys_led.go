package syscore

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/gkiryaziev/go-mqtt-client/service"
)

/*
Struct led provides system LED[0,1] control

Topics:
	Subscribe:
		name + "/SYSTEM/LED[0,1]/ACTION		{0, 1, STATUS}
	Publish:
		name + "/SYSTEM/LED[0,1]/STATUS		{0, 1}

Methods:
	Subscribe
	Unsubscribe
	PublishStatus

Functions:
	Set trigger to [none] when subscribe
		echo none | sudo tee /sys/class/leds/led0/trigger
	Set trigger to [mmc0] when unsubscribe
		echo mmc0 | sudo tee /sys/class/leds/led0/trigger
	Set brightness to 1 when ON
		echo 1 | sudo tee /sys/class/leds/led0/brightness
	Set brightness to 0 when OFF
		echo 0 | sudo tee /sys/class/leds/led0/brightness
	Get brightness status
		sudo cat /sys/class/leds/led0/brightness
*/
type led struct {
	client mqtt.Client
	debug  bool
	topic  string
	status string
	ledID  string
}

// a[len(a)-1:] last char

// newLed return new led object.
func newLed(c mqtt.Client, topic string, debug bool) *led {
	return &led{
		client: c,
		debug:  debug,
		topic:  topic,
		status: "0",
		ledID:  topic[len(topic)-1:],
	}
}

// Subscribe to topic
func (l *led) Subscribe(qos byte) {

	topic := l.topic + "/ACTION"

	log.Println("[RUN] Subscribing:", qos, topic)

	if token := l.client.Subscribe(topic, qos, l.ledMessageHandler); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}
}

// UnSubscribe from topic
func (l *led) UnSubscribe() {

	topic := l.topic + "/ACTION"

	log.Println("[RUN] UnSubscribing:", topic)

	if token := l.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}
}

// PublishStatus Led status
func (l *led) PublishStatus(qos byte) {

	topic := l.topic + "/STATUS"

	// publish result
	if token := l.client.Publish(topic, qos, false, l.status); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}

	// debug
	if l.debug {
		log.Println("[PUB]", qos, topic, l.status)
	}
}

// ledMessageHandler set Led to ON or OFF and get STATUS
func (l *led) ledMessageHandler(client mqtt.Client, msg mqtt.Message) {

	// debug
	if l.debug {
		log.Println("[SUB]", msg.Qos(), msg.Topic(), string(msg.Payload()))
	}

	// receive message and DO
	switch string(msg.Payload()) {
	case "0":
		// logic when OFF
		ledAction(l.ledID, string(msg.Payload()))
		l.status = ledStatus(l.ledID)
		l.PublishStatus(0)
	case "1":
		// logic when ON
		ledAction(l.ledID, string(msg.Payload()))
		l.status = ledStatus(l.ledID)
		l.PublishStatus(0)
	case "STATUS":
		// publish status
		l.status = ledStatus(l.ledID)
		l.PublishStatus(0)
	}
}

// ledAction
func ledAction(ledID, action string) {
	file := "/sys/class/leds/led" + ledID + "/brightness"
	service.WriteFile(file, action)
}

// ledStatus
func ledStatus(ledID string) string {
	file := "/sys/class/leds/led" + ledID + "/brightness"
	s, _ := service.ReadFile(file, 1)
	return s
}
