## Prerequisites

Right now, it's built using several external Go libraries. That
requires running `go get` to install them before trying to build
the application.

```bash
go get git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git
go get gopkg.in/inconshreveable/log15.v2
go get gopkg.in/yaml.v2
go get gopkg.in/gorp.v1
go get github.com/go-sql-driver/mysql
```

## Development

### Building and Running

Make sure you have the `$GOROOT` variable set up correctly.

Then run `go install github.com/bradberger/mqtt-client-logger && mqtt-client-logger` ;

### Installing

Run the install script from the base directory. It should take care
of everything, at least on Ubuntu.

If the version of Ubuntu is using Upstart (14.04), use `install-upstart`

```bash
./install-upstart
```

Otherwise, for newer versions use systemd:

```bash
./install-systemd
```

## Service Configuration

### Running

```bash
sudo service mqttlogger start|stop|restart
```

### Questions

Best to stay connected or to disconnect after the <interval> duration and then reconnect?

### Test SQL.

```sql
INSERT INTO brokers(Server,Username,Password,ClientID,Reconnect,CleanSession) VALUES("tcp://test.mqtt.dev:1883","superuser","password","go-logger",1,0);
INSERT INTO topics(Broker, Topic) VALUES(1, "$SYS/broker/bytes/received");
INSERT INTO topics(Broker, Topic) VALUES(1, "$SYS/broker/clients/total");
INSERT INTO topics(Broker, Topic) VALUES(1, "$SYS/#");
```
