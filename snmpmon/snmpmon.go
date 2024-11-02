package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	g "github.com/soniah/gosnmp"
	"log"
	"os"
	"strconv"
	s "strings"
	"time"
)

/*
Three snmpd config settings are needed:

agentAddress  udp:127.0.0.1
rocommunity public  localhost
includeAllDisks 10000

*/

/*

Graphite metric name rules

Metric names must start with a letter
Can only contain ascii alphanumerics, underscore and periods (other characters will get converted to underscores)
Should not exceed 200 characters (though less than 100 is genearlly preferred from a UI perspective)
Unicode is not supported
We recommend avoiding spaces

*/

// prefix, o.i.d.1=name
// Note the first entry of a multi-entry line will be used as the key
var OidList = [][]string{
	{"disk.space", ".1.3.6.1.4.1.2021.9.1.2=path,6=total,7=avail,8=used,9=percent"},
	{"net.packets.in", ".1.3.6.1.2.1.31.1.1.1.1=name,7=unicast,8=multicast,9=broadcast"},
	{"net.packets.out", ".1.3.6.1.2.1.31.1.1.1.1=name,11=unicast,12=multicast,13=broadcast"},
	{"net.octets.in", ".1.3.6.1.2.1.31.1.1.1.1=name,6=count"},
	{"net.octets.out", ".1.3.6.1.2.1.31.1.1.1.1=name,10=count"},
	{"disk.io.read", ".1.3.6.1.4.1.2021.13.15.1.1.2=name,5=accesses,12=bytes"},
	{"disk.io.write", ".1.3.6.1.4.1.2021.13.15.1.1.2=name,6=accesses,13=bytes"},
	{"disk.io.loadavg", ".1.3.6.1.4.1.2021.13.15.1.1.2=name,9=1,10=5,11=15"},
	{"loadavg.1", ".1.3.6.1.4.1.2021.10.1.5.1"},
	{"loadavg.5", "1.3.6.1.4.1.2021.10.1.5.2"},
	{"loadavg.15", "1.3.6.1.4.1.2021.10.1.5.3"},
	{"swap.total", "1.3.6.1.4.1.2021.4.3"},
	{"swap.avail", "1.3.6.1.4.1.2021.4.4"},
	{"mem.total", "1.3.6.1.4.1.2021.4.5"},
	{"mem.avail", "1.3.6.1.4.1.2021.4.6"},
	{"mem.free", "1.3.6.1.4.1.2021.4.11"},
	{"mem.shared", "1.3.6.1.4.1.2021.4.13"},
	{"mem.buffered", "1.3.6.1.4.1.2021.4.14"},
	{"mem.cached", "1.3.6.1.4.1.2021.4.15"},
	{"mem.swap.in", "1.3.6.1.4.1.2021.11.3"},
	{"mem.swap.out", "1.3.6.1.4.1.2021.11.4"},
	{"sys.interrupts", "1.3.6.1.4.1.2021.11.7"},
	{"sys.context", "1.3.6.1.4.1.2021.11.8"},
	{"cpu.user", "1.3.6.1.4.1.2021.11.50"},
	{"cpu.nice", "1.3.6.1.4.1.2021.11.51"},
	{"cpu.system", "1.3.6.1.4.1.2021.11.52"},
	{"cpu.idle", "1.3.6.1.4.1.2021.11.53"},
	{"cpu.wait", "1.3.6.1.4.1.2021.11.54"},
	{"cpu.kernel", "1.3.6.1.4.1.2021.11.55"},
	{"cpu.interrupt", "1.3.6.1.4.1.2021.11.56"},
}

type ConfigVars struct {
	AWSRegion      string
	SQSQueueURL    string
	SNMPHost       string
	SNMPPort       string
	SNMPCommunity  string
	OutputIDPrefix string
	Debug          bool
}

var ResultData = make([]g.SnmpPDU, 0)

func main() {

	var category string
	var value string
	var data string

	cv := getEnvVars()

	for {
		fmt.Println("Gathering metrics.")
		start := time.Now()
		data = ""

		for p := range OidList {
			category = OidList[p][0]
			value = OidList[p][1]

			ResultData = make([]g.SnmpPDU, 0) // to clear it each iteration

			if s.Contains(value, "=") {
				data = data + correlate(category, value, cv)
			} else {
				data = data + get_value(category, value, cv)
			}
		}
		if cv.Debug {
			fmt.Print(data)
			fmt.Println("Debug mode: Data not sent to SQS.")
		} else {
			fmt.Print("Sending to SQS: ")
			sendToSQS(data, cv)
		}
		sleeptime := 1*time.Minute - time.Now().Sub(start)
		fmt.Println("Sleeping for", sleeptime)

		time.Sleep(sleeptime)
	}

}

func getEnvVars() ConfigVars {
	vars := ConfigVars{}
	var err error

	vars.AWSRegion = os.Getenv("AWS_REGION")
	if vars.AWSRegion == "" {
		log.Fatal("Env var AWS_REGION must be set.")
	}
	vars.SQSQueueURL = os.Getenv("SQS_QUEUE_URL")
	if vars.SQSQueueURL == "" {
		log.Fatal("Env var SQS_QUEUE_URL must be set.")
	}
	vars.SNMPHost = os.Getenv("SNMP_HOST")
	if vars.SNMPHost == "" {
		log.Fatal("Env var SNMP_HOST must be set.")
	}
	vars.SNMPPort = os.Getenv("SNMP_PORT")
	if vars.SNMPPort == "" {
		fmt.Println("Env var SNMP_PORT not set, defaulting to 161.")
		vars.SNMPPort = "161"
	}
	vars.SNMPCommunity = os.Getenv("SNMP_COMMUNITY")
	if vars.SNMPCommunity == "" {
		fmt.Println("Env var SNMP_COMMUNITY not set, defaulting to public.")
		vars.SNMPCommunity = "public"
	}
	vars.OutputIDPrefix = os.Getenv("OUTPUT_ID_PREFIX")
	if vars.OutputIDPrefix == "" {
		log.Fatal("Env var OUTPUT_ID_PREFIX not set.")
	}
	vars.Debug, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		vars.Debug = false
	}

	return vars
}

func sendToSQS(message string, cv ConfigVars) {

	session_options := session.Options{
		Config: aws.Config{Region: aws.String(cv.AWSRegion)},
		//SharedConfigState: session.SharedConfigEnable,
	}

	sess := session.Must(session.NewSessionWithOptions(session_options))

	svc := sqs.New(sess)

	qURL := cv.SQSQueueURL

	sqs_message := &sqs.SendMessageInput{
		MessageBody: aws.String(message),
		QueueUrl:    &qURL,
	}

	result, err := svc.SendMessage(sqs_message)

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Println("Success", *result.MessageId)
}

func correlate(category string, oidpattern string, cv ConfigVars) string {
	var key_name string
	var key_number string
	var key_map = make(map[string]string)
	var out string
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	splitstr := s.Split(oidpattern, ".")
	numbers_and_names := s.Split(splitstr[len(splitstr)-1], ",")
	baseoid := s.Join(splitstr[:len(splitstr)-1], ".")
	getData(baseoid, cv)
	//fmt.Println(baseoid, numbers_and_names)
	for n := range numbers_and_names {
		nnsplit := s.Split(numbers_and_names[n], "=")
		number := nnsplit[0]
		name := nnsplit[1]
		//fmt.Println(baseoid, number, name)
		if key_name == "" && key_number == "" { // key info has not been set
			key_name = name
			key_number = number
			//fmt.Printf("Key is %s, num is %s\n", key_name, key_number)
			//continue
		}
		//fmt.Printf("int.system.%s.%s.%s\n", category, key, name)
		current_node := s.Join([]string{baseoid, number}, ".")
		for i := range ResultData {
			result_value := returnValue(ResultData[i])
			current_oid := ResultData[i].Name
			current_split := s.Split(current_oid, ".")
			current_instance := current_split[len(current_split)-1]
			current_item := current_split[len(current_split)-2]
			if s.Index(current_oid, current_node) == 0 {
				//fmt.Println(current_oid, current_node)
				if current_item == key_number {
					//fmt.Println(number, key_number, result_value)
					key_map[current_instance] = s.Replace(result_value, "/", "_", -1)
					continue
				} else if current_item == number {
					out = out + fmt.Sprintf("%s.%s.%s.%s %s %s\n", cv.OutputIDPrefix, category, key_map[current_instance], name, result_value, ts)
				}
			} else {
				//fmt.Println("No match")
			}
		}
	}
	//fmt.Print(out)
	return out
}

func getData(oid string, cv ConfigVars) {

	var params *g.GoSNMP

	snmpport, err := strconv.ParseUint(cv.SNMPPort, 10, 16)
	if err != nil {
		log.Fatalf("Error parsing SNMP_PORT: %v", err)
	}

	if cv.Debug {
		params = &g.GoSNMP{
			Target:    cv.SNMPHost,
			Port:      uint16(snmpport),
			Community: cv.SNMPCommunity,
			Version:   g.Version2c,
			Timeout:   time.Duration(5) * time.Second,
			Logger:    log.New(os.Stdout, "", 0),
		}
	} else {
		params = &g.GoSNMP{
			Target:    cv.SNMPHost,
			Port:      uint16(snmpport),
			Community: cv.SNMPCommunity,
			Version:   g.Version2c,
			Timeout:   time.Duration(5) * time.Second,
			//Logger: log.New(os.Stdout, "", 0),
		}
	}
	err = params.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer params.Conn.Close()
	if cv.Debug {
		fmt.Println("Starting BulkWalk")
	}
	err = params.BulkWalk(oid, addValue)
	if cv.Debug {
		fmt.Println("BulkWalk finished")
	}
	if err != nil {
		fmt.Printf("Walk Error: %v\n", err)
		os.Exit(1)
	}
	if cv.Debug {
		fmt.Println("Printing ResultData")
		printValues()
		fmt.Println("finished printing ResultData")
	}
}

func printValues() error {
	var err error
	for p := range ResultData {
		err = printValue(ResultData[p])
	}
	return err
}

func addValue(pdu g.SnmpPDU) error {
	ResultData = append(ResultData, pdu)
	return nil

}

func returnValue(pdu g.SnmpPDU) string {
	switch pdu.Type {
	case g.OctetString:
		b := pdu.Value.([]byte)
		return string(b)
	default:
		return g.ToBigInt(pdu.Value).String()
	}
}

func printValue(pdu g.SnmpPDU) error {
	fmt.Printf("NAME: %s ", pdu.Name)
	switch pdu.Type {
	case g.OctetString:
		b := pdu.Value.([]byte)
		fmt.Printf("STRING: %s\n", string(b))
	default:
		fmt.Printf("TYPE: %d  VALUE: %s\n", pdu.Type, g.ToBigInt(pdu.Value).String())
	}
	return nil
}

func get_value(category, oid string, cv ConfigVars) string {
	var out string
	getData(oid, cv)
	base := "%s.%s %s %s\n"

	ts := strconv.FormatInt(time.Now().Unix(), 10)

	for i := range ResultData {

		out = out + fmt.Sprintf(base, cv.OutputIDPrefix, category, returnValue(ResultData[i]), ts)
	}
	if cv.Debug {
		fmt.Print(out)
	}
	return out
}
