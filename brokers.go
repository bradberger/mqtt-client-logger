package main

import (
	"fmt"
	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v1"
	log "gopkg.in/inconshreveable/log15.v2"
	"time"
)

type Broker struct {
	Id, LoggerID                         int64
	Server, Username, Password, ClientID string
	Reconnect, CleanSession              bool
	Dbm                                  *gorp.DbMap     `db:"-"`
	Topics                               map[string]byte `db:"-"`
}

type Topic struct {
	Id     int64
	Broker int64
	Topic  string
}

type Message struct {
	Id       int64
	Broker   int64
	Topic    string
	Data     string
	LoggerID int64
	Created  int64
}

type Status struct {
	Topics  map[string][]string
	Brokers map[string]Broker
}

type BrokerStatus struct {
	Topics []string
	Broker *Broker
	Client *mqtt.Client
}

var status = make(map[string]BrokerStatus)

func (m *Message) PreInsert(s gorp.SqlExecutor) error {
	m.Created = time.Now().UnixNano()
	return nil
}

// Callback when a message is received from a subscription.
// Right now, it inserts the message in the messages table of
// the database.
func (b Broker) OnMessage(client *mqtt.Client, msg mqtt.Message) {
	log.Info(fmt.Sprintf("Message received: %s", msg))

	err := b.Dbm.Insert(&Message{Topic: msg.Topic(), Data: string(msg.Payload()), Broker: b.Id, LoggerID: loggerID})
	checkErr(err, "Failed to insert message into database")

}

// Connects a the broker
func (b Broker) Connect(c *mqtt.Client) {

	go func() {
		log.Info(fmt.Sprintf("Connecting as %s to %s\n", b.ClientID, b.Server))
		token := c.Connect()
		if token.Wait() && token.Error() != nil {
			checkErr(token.Error(), fmt.Sprintf("Could not connect to %s", b.Server))
		}
	}()

}

// This looks up and stores the topics the given broker
// should be subscribing to.
func (b Broker) GetTopics() (topics map[string]byte, err error) {

	// Return array of topics based on SELECT * FROM topics WHERE Broker = b.Id
	var t []Topic

	sql := fmt.Sprintf("SELECT * FROM mqtt_logger_topics WHERE Broker = %v", b.Id)
	_, err = b.Dbm.Select(&t, sql)

	// Initialize the map here to prevent panics on assignment to nil.
	topics = make(map[string]byte)
	for _, item := range t {
		topics[item.Topic] = 1
	}

	return

}

// Simple helper function to check if a value is contained
// in a slice of strings.
func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// The connection callback when a client successfully connects.
// It subscribes to the topics that are defined in the MySQL table
// for the given broker
func (b Broker) OnClientConnect(c *mqtt.Client) {

	b.updateSubscriptions(c)

	s := status[b.Server]
	s.Client = c
	s.Broker = &b

	status[b.Server] = s

}

// Handles the subscribing and unsubscribing of topics for
// a given broker. Is called on connect and when re-loading
// configuration from the database.
func (b Broker) updateSubscriptions(c *mqtt.Client) {

	topics, err := b.GetTopics()
	checkErr(err, "Could not get topics")

	// Check for new entries to subscribe.
	for t, _ := range topics {

		if !isValueInList(t, status[b.Server].Topics) {

			log.Info(fmt.Sprintf("Subscribing to %s on %s", t, b.Server))

			// If don't do it this way, get a cannot assign to error.
			s := status[b.Server]
			s.Topics = append(status[b.Server].Topics, t)
			status[b.Server] = s

			go func() {
				token := c.Subscribe(t, 1, b.OnMessage)
				if token.Wait() && token.Error() != nil {
					delete(status, b.Server)
					checkErr(token.Error(), "Error subscribing to topic(s)")
				}
			}()

		}

	}

	// Now check to unsubscribe if topic has been deleted.
	// Create a new empty slice for topics, and if found, we
	// append to it, if not found just skip it.
	newTopics := []string{}
	for _, t := range status[b.Server].Topics {

		found := false
		for k, _ := range topics {
			if t == k {
				found = true
			}
		}

		// No longer subscribed, so unsubcribe now
		if !found {
			log.Info(fmt.Sprintf("Unsubscribing from %s on %s", t, b.Server))
			token := status[b.Server].Client.Unsubscribe(t)
			checkErr(token.Error(), "Error unsubscribing")
		} else {
			newTopics = append(newTopics, t)
		}

	}

	// Now update the status variable with the new Topics list for this server.
	s := status[b.Server]
	s.Topics = newTopics
	status[b.Server] = s

}

// Loads the brokers from the database and connects to them.
func loadBrokers() {

	initDb()

	_, err := dbmap.Select(&cfg.Brokers, fmt.Sprintf("SELECT * FROM mqtt_logger_config WHERE LoggerID = %v", loggerID))
	fatalErr(err, "Could not get the list of brokers")

	for _, broker := range cfg.Brokers {

		// Check to see if already status, and if so, just
		// skip ahead.
		s, okay := status[broker.Server]
		if okay {
			fmt.Sprintf("Re-checking config for %s", broker.Server)
			s.Broker.updateSubscriptions(s.Client)
			continue
		}

		status[broker.Server] = BrokerStatus{Topics: []string{}}

		broker.Dbm = dbmap

		opts := mqtt.NewClientOptions()
		opts.AddBroker(broker.Server)

		if broker.ClientID != "" {
			opts.SetClientID(broker.ClientID)
		}

		if broker.Username != "" {
			opts.SetUsername(broker.Username)
		}

		if broker.Password != "" {
			opts.SetPassword(broker.Password)
		}

		opts.SetAutoReconnect(broker.Reconnect)
		opts.SetCleanSession(broker.CleanSession)
		opts.OnConnect = broker.OnClientConnect
		client := mqtt.NewClient(opts)

		go broker.Connect(client)

	}

}
