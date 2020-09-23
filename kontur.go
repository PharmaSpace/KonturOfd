package KonturOfd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type UserData struct {
	Email string
	Password string
	ApiKey string
	Sid string
}

type Receipts struct {
	Data []*Receipt
}

type Receipt struct {
	ReceiptCode        int    `json:"receiptCode"`
	User               string `json:"user"`
	UserInn            string `json:"userInn"`
	RequestNumber      int    `json:"requestNumber"`
	DateTime           string `json:"dateTime"`
	ShiftNumber        int    `json:"shiftNumber"`
	OperationType      int    `json:"operationType"`
	TaxationType       int    `json:"taxationType"`
	Operator           string `json:"operator"`
	KktRegID           string `json:"kktRegId"`
	FiscalDriveNumber  string `json:"fiscalDriveNumber"`
	RetailPlaceAddress string `json:"retailPlaceAddress"`
	Items              []Items `json:"items"`
	Nds18                int   `json:"nds18"`
	TotalSum             int   `json:"totalSum"`
	CashTotalSum         int   `json:"cashTotalSum"`
	EcashTotalSum        int   `json:"ecashTotalSum"`
	FiscalDocumentNumber int   `json:"fiscalDocumentNumber"`
	FiscalSign           int64 `json:"fiscalSign"`
}

type Items struct {
	Items []Item
}

type Item struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	Sum      int    `json:"sum"`
}

type Organizations struct {
	Data []*Organization
}

type Organization struct {
	ID        string `json:"id"`
	Inn       string `json:"inn"`
	Kpp       string `json:"kpp"`
	Ogrn      string `json:"ogrn"`
	ShortName string `json:"shortName"`
	FullName  string `json:"fullName"`
}

type Cashboxes struct {
	Data []*Cashbox
}

type Cashbox struct {
	RegNumber    string `json:"regNumber"`
	SerialNumber string `json:"serialNumber"`
	Address      string `json:"address"`
	Name         string `json:"name"`
	Kpp          string `json:"kpp"`
	FiscalDrive  struct {
		FiscalDriverNumber        string `json:"fiscalDriverNumber"`
		EarliestDocumentTimestamp string `json:"earliestDocumentTimestamp"`
	} `json:"fiscalDrive"`
	SalesPointName string `json:"salesPointName"`
}

type authResp struct {
	Sid string `json:"Sid"`
}

func (u *UserData)auth() error {
	reqBody := bytes.Buffer{}
	reqBody.Write([]byte(u.Password))
	resp, err := http.Post(fmt.Sprintf("api.kontur.ru/auth/authenticate-by-pass?login=testlogin@%s",u.Email ), "text/plain", &reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	result := authResp{}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return err
	}
	u.Sid = result.Sid
	return nil
}

func (u *UserData) GetReceipts(date string) ([]*Receipt, error) {
	err := u.auth()
	if err != nil {
		return nil, err
	}
	organizations, err := u.getOrganizations()
	if err != nil {
		return nil, err
	}
	allReceipts := make([]*Receipt, 0)
	for _, organization := range organizations.Data {
		cashboxes, err := u.getCashboxes(organization.ID)
		if err != nil {
			return nil, err
		}
		for _, cashbox := range cashboxes.Data {
			receipts, err := u.getReceipts(organization.ID, cashbox.RegNumber, date)
			if err != nil {
				return nil, err
			}
			allReceipts = append(allReceipts, receipts...)
		}
	}
	return allReceipts, nil
}

func (u *UserData) getOrganizations() (*Organizations, error) {
	resp, err := u.doGetRequest("/v2/organizations")
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	organizations := Organizations{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &organizations)
	return &organizations, err
}

func (u *UserData) getCashboxes(organizationId string) (*Cashboxes, error){
	resp, err := u.doGetRequest(fmt.Sprintf("/v2/organizations/%s/cashboxes", organizationId))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cashboxes := Cashboxes{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &cashboxes)
	return &cashboxes, err
}

func (u *UserData) getReceipts(organizationId, kktRegId, date string) ([]*Receipt, error) {
	resp, err := u.doGetRequest(fmt.Sprintf("v2/organizations/%s/cashboxes/%s/documents?date=%s", organizationId, kktRegId, date))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	receipts := make([]*Receipt, 0)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &receipts)
	return receipts, err
}
func (u *UserData) doGetRequest(path string) (*http.Response, error){
	client := http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://ofd-api.kontur.ru%s", path), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("ofd_api_key", u.ApiKey)
	req.Header.Add("auth.sid", u.Sid)

	resp, err := client.Do(req)
	return resp, err
}
