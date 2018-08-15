# aws pricing typer
[![Build Status](https://travis-ci.org/jonhadfield/aws-pricing-typer.svg?branch=master)](https://travis-ci.org/jonhadfield/aws-pricing-typer)
[![Coverage Status](https://coveralls.io/repos/github/jonhadfield/aws-pricing-typer/badge.svg?branch=master)](https://coveralls.io/github/jonhadfield/aws-pricing-typer?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonhadfield/aws-pricing-typer)](https://goreportcard.com/report/github.com/jonhadfield/aws-pricing-typer)

## about

Using the [Go SDK](https://aws.amazon.com/sdk-for-go/) to query the AWS Pricing API for Products (metadata plus pricing detail) returns a slice of type: aws.JSONValue that's defined as:
``` map[string]interface{} ```.
This library takes the output and returns the documents in structs with their concrete types.

## note

- AWS API credentials are required in order to query the pricing API
- >"The Price List Service API provides pricing details for your information only. If there is a discrepancy between the offer file and a service pricing page, AWS charges the prices that are listed on the service pricing page. For more information about AWS service pricing, see Cloud Services Pricing." 

## usage

```go
package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/jonhadfield/aws-pricing-typer"
)

func main() {
	pricingAPIRegion := "us-east-1"

	// get pricing client
	sess, _ := session.NewSession(&aws.Config{Region: &pricingAPIRegion})
	svc := pricing.New(sess)

	// create request criteria
	ec2ServiceCode := "AmazonEC2"
	formatVer := "aws_v1"
	typeTerm := pricing.FilterTypeTermMatch

	// create filters
	locationKey := "location"
	locationValue := "EU (Ireland)"
	instanceTypeKey := "instanceType"
	instanceTypeValue := "m4.large"
	var priceFilters []*pricing.Filter
	priceFilters = append(priceFilters, &pricing.Filter{
		Type:  &typeTerm,
		Field: &locationKey,
		Value: &locationValue,
	})
	priceFilters = append(priceFilters, &pricing.Filter{
		Type:  &typeTerm,
		Field: &instanceTypeKey,
		Value: &instanceTypeValue,
	})

	// make request
	productsOutput, _ := svc.GetProducts(&pricing.GetProductsInput{
		ServiceCode:   &ec2ServiceCode,
		FormatVersion: &formatVer,
		Filters:       priceFilters,
	})

	// transform
	priceData, _ := awsPricingTyper.GetTypedPricingData(*productsOutput)
}
```