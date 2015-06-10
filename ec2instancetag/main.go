package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"io/ioutil"
	"net/http"
	"os"
)

func fetchMetadata(path string) (string, error) {
	url := fmt.Sprintf("http://169.254.169.254/latest/meta-data/%s", path)

	resp, err := http.Get(url)

	if nil != err {
		return "", err
	}

	defer resp.Body.Close()

	if b, err := ioutil.ReadAll(resp.Body); nil == err {
		return string(b), nil
	} else {
		return "", err
	}
}

func getEc2Service() (*ec2.EC2, error) {
	region, err := fetchMetadata("placement/availability-zone")

	if nil != err {
		return nil, err
	}

	region = region[0 : -1+len(region)]

	return ec2.New(&aws.Config{Region: region}), nil
}

func main() {
	if 2 != len(os.Args) {
		fmt.Println("Usage: ec2instancetag KEY")
		os.Exit(1)
	}

	keyToLookup := os.Args[1]

	ec2service, err := getEc2Service()

	if nil != err {
		panic(err)
	}

	instanceId, err := fetchMetadata("instance-id")

	if nil != err {
		panic(err)
	}

	result, err := ec2service.DescribeTags(&ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(instanceId)},
			},
		},
	})

	if nil != err {
		panic(err)
	}

	tags := make(map[string]string)

	for _, tag := range result.Tags {
		tags[*tag.Key] = *tag.Value
	}

	if v, ok := tags[keyToLookup]; !ok {
		os.Exit(127)
	} else {
		fmt.Println(v)
		os.Exit(0)
	}
}
