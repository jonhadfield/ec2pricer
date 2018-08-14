package ec2pricer

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/davecgh/go-spew/spew"
	"github.com/jonhadfield/aws-pricing-typer"
	"github.com/olekukonko/tablewriter"
)

type InstanceAppConfig struct {
	InstanceType    string
	Location        string
	Tenancy         string
	PreInstalledSw  string
	OperatingSystem string
	Output          string
	Debug           bool
}

type GetEC2InstancePriceInput struct {
	Location        string
	InstanceType    string
	OperatingSystem string
	Tenancy         string
	PreInstalledSw  string
	Term            string
}

var (
	pricingAPIRegion = "us-east-1"
)

func getStrPtr(input string) *string {
	return &input
}

type Product struct {
	ProductFamily string
	SKU           string
	Attributes    struct {
		NetworkPerformance          string
		VCPU                        string
		CapacityStatus              string
		OperatingSystem             string
		PhysicalProcessor           string
		ECU                         string
		PreInstalledSw              string
		ProcessorArchitecture       string
		EnhancedNetworkingSupported string
		Storage                     string
		ClockSpeed                  string
		Tenancy                     string
		LicenseModel                string
		ServiceCode                 string
		CurrentGeneration           string
		DedicatedEbsThroughput      string
		ServiceName                 string
		InstanceType                string
		NormalizationSizeFactor     string
		ProcessorFeatures           string
		Operation                   string
		Memory                      string
		LocationType                string
		InstanceFamily              string
		UsageType                   string
		Location                    string
	}
}

type GetEC2InstancePricingOutput struct {
	Product        Product
	OnDemandPrices string
	ReservedPrices string
}

func GetInstancePricing(config *InstanceAppConfig) {
	sess, err := session.NewSession(&aws.Config{Region: &pricingAPIRegion})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	svc := pricing.New(sess)
	ec2ServiceCode := "AmazonEC2"
	formatVer := "aws_v1"
	typeTerm := pricing.FilterTypeTermMatch
	var getEC2InstancePriceFilters []*pricing.Filter

	getEC2InstancePriceFilters = append(getEC2InstancePriceFilters, &pricing.Filter{
		Type:  &typeTerm,
		Field: getStrPtr("location"),
		Value: getStrPtr(config.Location),
	})

	if config.InstanceType != "" {
		getEC2InstancePriceFilters = append(getEC2InstancePriceFilters, &pricing.Filter{
			Type:  &typeTerm,
			Field: getStrPtr("instanceType"),
			Value: getStrPtr(config.InstanceType),
		})
	}

	if config.OperatingSystem != "" {
		getEC2InstancePriceFilters = append(getEC2InstancePriceFilters, &pricing.Filter{
			Type:  &typeTerm,
			Field: getStrPtr("operatingSystem"),
			Value: getStrPtr(config.OperatingSystem),
		})
	}

	if config.Tenancy != "" {
		getEC2InstancePriceFilters = append(getEC2InstancePriceFilters, &pricing.Filter{
			Type:  &typeTerm,
			Field: getStrPtr("tenancy"),
			Value: getStrPtr(config.Tenancy),
		})
	}

	if config.PreInstalledSw != "" {
		getEC2InstancePriceFilters = append(getEC2InstancePriceFilters, &pricing.Filter{
			Type:  &typeTerm,
			Field: getStrPtr("preInstalledSw"),
			Value: getStrPtr(config.PreInstalledSw),
		})
	}

	if config.Debug {
		fmt.Printf("Filters: %+v\n\n", getEC2InstancePriceFilters)
	}
	getProductsOutput, descErr := svc.GetProducts(&pricing.GetProductsInput{
		ServiceCode:   &ec2ServiceCode,
		FormatVersion: &formatVer,
		Filters:       getEC2InstancePriceFilters,
	})
	if descErr != nil {
		fmt.Println(descErr)
		os.Exit(1)
	}
	pricingData, getTypedErr := awsPricingTyper.GetTypedPricingData(*getProductsOutput)
	if config.Debug {
		spew.Dump(pricingData)
	}
	if getTypedErr != nil {
		log.Fatal(getTypedErr)
	}
	if pricingData != nil {
		outputTypeInfo(&pricingData[0])
	} else {
		fmt.Println("No results found.")
		os.Exit(0)
	}
	for i := range pricingData {
		item := &pricingData[i]
		// Process OS value
		var license string
		switch strings.ToLower(item.Product.Attributes.LicenseModel) {
		case "no license required":
			license = "NA"
		case "bring your own license":
			license = "BYOL"
		default:
			license = item.Product.Attributes.LicenseModel
		}

		fmt.Printf("OS: %s | Tenancy: %s | SW: %s | License: %s\n", item.Product.Attributes.OperatingSystem,
			item.Product.Attributes.Tenancy, item.Product.Attributes.PreInstalledSw, license)
		//var productData [][]string
		//
		//itemData := []string{item.Product.Attributes.OperatingSystem, item.Product.Attributes.Tenancy, item.Product.Attributes.PreInstalledSw, license}
		//productData = append(productData, itemData)
		//productTable := tablewriter.NewWriter(os.Stdout)
		//productTable.SetHeader([]string{"OS", "Tenancy", "SW", "License"})
		//productTable.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: false})
		//productTable.SetCenterSeparator("|")
		//productTable.AppendBulk(productData) // Add Bulk Data
		//productTable.Render()
		// output terms
		var termsData [][]string

		//termsData = append(termsData)
		//ODTermDesc := item.Terms.OnDemand
		//fmt.Printf("%+v\n", item.Terms.OnDemand)
		for _, term := range item.Terms.OnDemand {
			var upFrontCost, pricePerHour awsPricingTyper.PricePerUnit
			// loop through dimensions
			for _, pd := range term.PriceDimensions {
				for _, pdV := range pd {
					if strings.ToLower(pdV.Unit) == "quantity" {
						for _, unitPrice := range pdV.PricePerUnit {
							upFrontCost = unitPrice
						}
					} else if strings.ToLower(pdV.Unit) == "hrs" {
						for _, unitPrice := range pdV.PricePerUnit {
							pricePerHour = unitPrice
						}
					}
				}
			}
			termDesc := "On Demand"
			termType := "NA"
			termData := []string{termDesc, termType, fmt.Sprintf("%.2f", upFrontCost["USD"]), fmt.Sprintf("%.3f", pricePerHour["USD"])}
			termsData = append(termsData, termData)
		}

		for _, term := range item.Terms.Reserved {
			var upFrontCost, pricePerHour awsPricingTyper.PricePerUnit
			// loop through dimensions
			for _, pd := range term.PriceDimensions {
				for _, pdV := range pd {
					if strings.ToLower(pdV.Unit) == "quantity" {
						for _, unitPrice := range pdV.PricePerUnit {
							upFrontCost = unitPrice
						}
					} else if strings.ToLower(pdV.Unit) == "hrs" {
						for _, unitPrice := range pdV.PricePerUnit {
							pricePerHour = unitPrice
						}
					}
				}
			}
			termDesc := fmt.Sprintf("%s %s", term.TermAttributes.LeaseContractLength, term.TermAttributes.PurchaseOption)
			termType := term.TermAttributes.OfferingClass
			termData := []string{termDesc, termType, fmt.Sprintf("%.2f", upFrontCost["USD"]), fmt.Sprintf("%.3f", pricePerHour["USD"])}
			termsData = append(termsData, termData)
		}
		termsTable := tablewriter.NewWriter(os.Stdout)
		termsTable.SetHeader([]string{"Term", "Type", "Up Front ($)", "Hourly ($)"})
		termsTable.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
		termsTable.SetCenterSeparator("|")
		termsTable.AppendBulk(termsData) // Add Bulk Data
		termsTable.Render()
		fmt.Println()
	}
}

func outputTypeInfo(doc *awsPricingTyper.PricingDocument) {
	//var productData [][]string
	fmt.Println()
	fmt.Printf("TYPE      %s\n", doc.Product.Attributes.InstanceType)
	fmt.Printf("LOCATION  %s\n", doc.Product.Attributes.Location)

	//productData = append(productData, []string{"TYPE", doc.Product.Attributes.InstanceType})
	//productData = append(productData, []string{"LOCATION", doc.Product.Attributes.Location})

	//itemData := []string{"TYPE", doc.Product.Attributes.InstanceType}
	//itemData := []string{"LOCATION", doc.Product.Attributes.Location}
	//productData = append(productData, itemData)
	//productTable := tablewriter.NewWriter(os.Stdout)
	//productTable.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	//productTable.SetCenterSeparator("|")
	//productTable.SetBorder(false)
	//productTable.AppendBulk(productData) // Add Bulk Data
	//productTable.Render()
	fmt.Println()
}
