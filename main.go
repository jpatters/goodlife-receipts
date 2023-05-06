package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func BeginningOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func EndOfMonth(t time.Time) time.Time {
	return BeginningOfMonth(t).AddDate(0, 1, 0).Add(-time.Second)
}

func main() {
	uri := "https://receipts.goodlifefitness.com/"

	currentDate := time.Now()

	isFirstOfMonth := currentDate.AddDate(0, 0, -1).Month() != currentDate.Month()
	var startDate, endDate time.Time

	if isFirstOfMonth {
		startDate = BeginningOfMonth(currentDate.AddDate(0, -1, 0))
		endDate = startDate.AddDate(0, 0, 13)
	} else {
		startDate = BeginningOfMonth(currentDate.AddDate(0, -1, 0)).AddDate(0, 0, 14)
		endDate = EndOfMonth(currentDate.AddDate(0, -1, 0))
	}

	resp, err := http.Get(uri)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	parsedValues := getPageValues(resp.Body, []string{"__VIEWSTATE", "__EVENTVALIDATION", "__VIEWSTATEGENERATOR"})

	for key, value := range parsedValues {
		if value == "" {
			panic(fmt.Sprintf("Could not find %s", key))
		}
	}

	log.Println("Got values from page")

	form := url.Values{}

	for key, value := range parsedValues {
		form.Add(key, value)
	}

	form.Add("__EVENTTARGET", "ctl00$Copy$btnSubmit3day")
	form.Add("ctl00$ScriptManager1", "")
	form.Add("ctl00$Copy$first_name", os.Getenv("FIRST_NAME"))
	form.Add("ctl00$Copy$last_name", os.Getenv("LAST_NAME"))
	form.Add("ctl00$Copy$revcan_number", os.Getenv("BARCODE_NUMBER"))
	form.Add("ctl00$Copy$member_number", "")
	form.Add("ctl00$Copy$drpBirthMonth", os.Getenv("BIRTH_MONTH"))
	form.Add("ctl00$Copy$drpBirthDay", os.Getenv("BIRTH_DAY"))
	form.Add("ctl00$Copy$drpBirthYear", os.Getenv("BIRTH_YEAR"))
	form.Add("ctl00$Copy$PayorFirstName", "")
	form.Add("ctl00$Copy$PayorLastName", "")
	form.Add("ctl00$Copy$ParentBarcode", "")
	form.Add("ctl00$Copy$ParentMembershipNumber", "")
	form.Add("ctl00$Copy$street", os.Getenv("STREET_ADDRESS"))
	form.Add("ctl00$Copy$city", os.Getenv("CITY"))
	form.Add("ctl00$Copy$province", os.Getenv("PROVINCE"))
	form.Add("ctl00$Copy$postal_code", os.Getenv("POSTAL_CODE"))
	form.Add("ctl00$Copy$email", os.Getenv("EMAIL_ADDRESS"))
	form.Add("ctl00$Copy$email2", os.Getenv("EMAIL_ADDRESS"))
	form.Add("ctl00$Copy$telephone", os.Getenv("PHONE_NUMBER"))
	form.Add("ctl00$Copy$startdate", startDate.Format("01/02/2006"))
	form.Add("ctl00$Copy$enddate", endDate.Format("01/02/2006"))
	form.Add("ctl00$Copy$CheckBox1", "on")

	log.Println("Sending request")
	response, err := http.Post(uri, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		panic(fmt.Sprintf("Status code error: %d %s", response.StatusCode, response.Status))
	}

	log.Println("Success")
}

func getPageValues(body io.Reader, searchValues []string) map[string]string {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}

	results := make(map[string]string)

	for _, searchValue := range searchValues {
		results[searchValue] = doc.Find(fmt.Sprintf("#%s", searchValue)).AttrOr("value", "")
	}

	return results
}
