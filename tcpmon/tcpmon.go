package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type ConfigVars struct {
	AMQP_URL         string
	AMQP_EXCHANGE    string
	AMQP_ROUTING_KEY string
	HOSTS            []string
	TIMEOUT_SEC      int
	OUTPUT_PREFIX    string
	DEBUG            bool
}

func main() {

	cv := getEnvVars()

	for {
		start := time.Now()

		log.Print("Gathering TCP status.")
		messages := TestHosts(cv)
		log.Printf("Collected info for %d hosts", len(messages))
		SendToAmqp(cv, strings.Join(messages, "\n"))

		sleeptime := 1*time.Minute - time.Now().Sub(start)
		log.Printf("Sleeping for %s", sleeptime)

		time.Sleep(sleeptime)
	}

}

func TestHosts(cv ConfigVars) []string {

	messages := make([]string, 0)

	for h := range cv.HOSTS {
		currenthost := cv.HOSTS[h]
		status := TestTCPPort(cv, currenthost)
		var statvalue string
		if status == true {
			statvalue = "10"
		} else {
			statvalue = "0"
		}
		host, port, err := net.SplitHostPort(currenthost)
		if err != nil {
			log.Fatal("Host doesn't seem to be in the right format.")
		}
		hostname := strings.Split(host, ".")[0]
		output := "%s.%s.monitor.tcp.%s.open %s %s"
		ts := strconv.FormatInt(time.Now().Unix(), 10)
		message := fmt.Sprintf(output, cv.OUTPUT_PREFIX, hostname, port, statvalue, ts)
		messages = append(messages, message)
	}

	return messages

}

func TestTCPPort(cv ConfigVars, currenthost string) bool {
	timeout := time.Duration(cv.TIMEOUT_SEC) * time.Second
	conn, err := net.DialTimeout("tcp", currenthost, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func SendToAmqp(cv ConfigVars, message string) {
	conn, err := amqp.Dial(cv.AMQP_URL)
	if err != nil {
		log.Fatal("Cannot connect to AMQP_URL")
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Cannot create channel.")
	}

	var amqp_message amqp.Publishing
	amqp_message.Headers = amqp.Table{}
	amqp_message.ContentType = "text/plain"
	amqp_message.Body = []byte(message)
	amqp_message.DeliveryMode = 2

	err = ch.Publish(cv.AMQP_EXCHANGE, cv.AMQP_ROUTING_KEY, false, false, amqp_message)
	if err != nil {
		log.Fatal("Cannot send message.")
	}
}

func getEnvVars() ConfigVars {
	cv := ConfigVars{}
	var err error

	cv.AMQP_URL = os.Getenv("AMQP_URL")
	if cv.AMQP_URL == "" {
		log.Fatal("Env var AMQP_URL must be set.")
	}
	cv.AMQP_EXCHANGE = os.Getenv("AMQP_EXCHANGE")
	if cv.AMQP_EXCHANGE == "" {
		log.Fatal("Env var AMQP_EXCHANGE must bet set.")
	}
	cv.AMQP_ROUTING_KEY = os.Getenv("AMQP_ROUTING_KEY")
	if cv.AMQP_ROUTING_KEY == "" {
		log.Fatal("Env var AMQP_ROUTING_KEY must be set.")
	}
	cv.TIMEOUT_SEC, err = strconv.Atoi(os.Getenv("TIMEOUT_SEC"))
	if err != nil {
		log.Fatal("Env var TIMEOUT_SEC not set or invalid integer.")
	}
	for i := 1; ; i++ {
		host := os.Getenv(fmt.Sprintf("CHECK_HOST_%d", i))
		if host != "" {
			log.Printf("Found CHECK_HOST_%d -> %s", i, host)
			cv.HOSTS = append(cv.HOSTS, host)
		} else {
			log.Print("End of CHECK_HOST_*")
			break
		}
	}
	if len(cv.HOSTS) == 0 {
		log.Fatal("Env var HOSTS must be set.")
	}
	cv.OUTPUT_PREFIX = os.Getenv("OUTPUT_PREFIX")
	if cv.OUTPUT_PREFIX == "" {
		log.Fatal("Env var OUTPUT_PREFIX must be set")
	}
	cv.DEBUG, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		cv.DEBUG = false
	}
	return cv
}
