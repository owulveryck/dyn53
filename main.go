package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	fqdn := flag.String("fqdn", "", "Source file")
	hzID := flag.String("hzid", "", "Hosted Zone ID")
	flag.Parse()
	if *fqdn == "" {
		log.Fatal("Please provide a fqdn")
	}
	if *hzID == "" {
		log.Fatal("Please provide a Hosted Zone ID")
	}

	// Get the public ip from EC2 metadata
	mURL := "http://169.254.169.254/latest/meta-data/public-ipv4"
	timeout := time.Duration(30 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	res, err := client.Get(mURL)
	if err != nil {
		log.Fatal("cannot get ip from metadata", err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("cannot get ip from metadata", err)
	}
	IP := net.ParseIP(string(b))
	if IP == nil {
		log.Fatal("IP is nil or invalid")
	}

	svc := route53.New(session.New())

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{ // Required
			Changes: []*route53.Change{ // Required
				{ // Required
					Action: aws.String("UPSERT"), // Required
					ResourceRecordSet: &route53.ResourceRecordSet{ // Required
						Name: aws.String(*fqdn), // Required
						Type: aws.String("A"),   // Required
						ResourceRecords: []*route53.ResourceRecord{
							{ // Required
								Value: aws.String(IP.String()), // Required
							},
						},
						//TTL: aws.Int64(111),
					},
				},
			},
			Comment: aws.String("Updated on " + time.Now().String()),
		},
		HostedZoneId: aws.String(*hzID), // Required
	}
	_, err = svc.ChangeResourceRecordSets(params)
	if err != nil {
		log.Fatal("Unable to update DNS: ", err)
	}
}
