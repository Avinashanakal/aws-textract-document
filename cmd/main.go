package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
)

func main() {
	file, err := os.ReadFile("kitas.png")
	if err != nil {
		panic(err)
	}

	cfg, _ := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("ap-southeast-1"),
	)
	client := textract.NewFromConfig(cfg)

	resp, err := client.DetectDocumentText(context.Background(), &textract.DetectDocumentTextInput{
		Document: &types.Document{
			Bytes: file,
		},
	})

	if err != nil {
		panic(err)
	}

	type Block struct {
		Text string `json:"text"`
	}
	blocks := make([]Block, 0)
	for _, block := range resp.Blocks {
		if block.BlockType == "WORD" {
			blocks = append(blocks, Block{Text: *block.Text})
		}
	}

	var result []string
	for _, v := range blocks {
		result = append(result, v.Text)
	}

	var niora, permitNumber, fullName, place, dob, passport, passportExpiry, status, nationality, gender, address, occupation string

	nioraIndex := indexOf(result, "NIORA")
	if nioraIndex == -1 {
		panic("invalid kitas doc")
	}
	if nioraIndex != -1 && nioraIndex+2 < len(result) {
		niora = result[nioraIndex+2]
		if len(niora) < 11 {
			fmt.Println("invalid niora")
		}
	}

	permitNumberIndex := indexOf(result, "Number")
	if result[permitNumberIndex-1] == "Permit" {
		if permitNumberIndex != -1 && permitNumberIndex+2 < len(result) {
			permitNumber = result[permitNumberIndex+2]
		}
	}

	fullNameIndex := indexOf(result, "Full")
	if fullNameIndex != -1 && fullNameIndex+4 < len(result) {
		firstName := result[fullNameIndex+3]
		lastName := result[fullNameIndex+4]
		fullName = fmt.Sprintf("%s %s", firstName, lastName)
	}

	placeIndex := indexOf(result, "Place")
	if fullNameIndex != -1 && placeIndex+6 < len(result) {
		place = result[placeIndex+6]
	}

	dobIndex := indexOf(result, "Date")
	if dobIndex != -1 && dobIndex+6 < len(result) {
		dob = result[dobIndex+6]
	}

	passportIndex := indexOf(result, "Passport")
	if passportIndex != -1 && passportIndex+3 < len(result) {
		passport = result[passportIndex+3]
	}

	passportExpiryIndexes := FindIndexes(result, "Expiry")

	for _, index := range passportExpiryIndexes {
		if result[index-1] == "Passport" {
			if index != -1 && index+2 < len(result) {
				passportExpiry = result[index+2]
			}
		}

	}

	NationalityIndex := indexOf(result, "Nationality")
	if NationalityIndex != -1 && NationalityIndex+2 < len(result) {
		nationality = result[NationalityIndex+2]
	}

	GenderIndex := indexOf(result, "Gender")
	if GenderIndex != -1 && GenderIndex+2 < len(result) {
		gender = result[GenderIndex+2]
	}

	AddressIndex := indexOf(result, "Address")
	OccupationIndex := indexOf(result, "Occupation")

	if AddressIndex != -1 && OccupationIndex != -1 && AddressIndex+2 < len(result) && AddressIndex+2 < OccupationIndex {
		address = strings.Join(result[AddressIndex+2:OccupationIndex], " ")
	}
	StatusIndex := indexOf(result, "Status")
	if OccupationIndex != -1 && StatusIndex != -1 && OccupationIndex+2 < len(result) && OccupationIndex+2 < StatusIndex {
		occupation = strings.Join(result[OccupationIndex+2:StatusIndex], " ")
	}

	if StatusIndex != -1 && StatusIndex+2 < len(result) {
		status = result[StatusIndex+2]
	}

	data := Kitos{
		Niora:          niora,
		PermitNumber:   permitNumber,
		FullName:       fullName,
		Place:          place,
		DateOfBirth:    dob,
		PassportNumber: passport,
		PassportExpiry: passportExpiry,
		Nationality:    nationality,
		Gender:         gender,
		Address:        address,
		Occupation:     occupation,
		Status:         status,
	}
	kitosJsonData, _ := json.Marshal(data)
	fmt.Println(string(kitosJsonData))

}

// Function to find the index of a specific substring in a slice
func indexOf(slice []string, target string) int {
	for i, v := range slice {
		if strings.ToLower(v) == strings.ToLower(target) {
			return i
		}
	}
	return -1
}

func FindIndexes(result []string, target string) []int {
	var indexes []int
	for i, word := range result {
		if strings.ToLower(word) == strings.ToLower(target) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

type Kitos struct {
	Niora          string `json:"niora"`
	PermitNumber   string `json:"permitNumber"`
	FullName       string `json:"fullName"`
	Place          string `json:"place"`
	DateOfBirth    string `json:"dob"`
	PassportNumber string `json:"passportNumber"`
	PassportExpiry string `json:"passportExpiry"`
	Nationality    string `json:"nationality"`
	Gender         string `json:"gender"`
	Address        string `json:"address"`
	Occupation     string `json:"occupation"`
	Status         string `json:"status"`
}
