package accounts

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/datag8r/xerogo/accountingAPI/endpoints"
	"github.com/datag8r/xerogo/filter"
	"github.com/datag8r/xerogo/utils"
)

type Account struct {
	Code                    string            // Required For Creation // maxlen = 10
	Name                    string            // Required For Creation // maxlen = 150
	Type                    accountType       // Required For Creation
	BankAccountNumber       *string           `json:",omitempty"` // Required For Creation If Type == Bank
	Status                  accountStatusCode `json:"-"`
	Description             *string           `json:",omitempty"`
	BankAccountType         *bankAccountType  `json:",omitempty"`
	CurrencyCode            *string           `json:",omitempty"`
	TaxType                 string            // Will Be a taxType type when i make them
	EnablePaymentsToAccount bool
	ShowInExpenseClaims     bool
	AccountID               string `json:",omitempty"`
	Class                   accountClassType
	SystemAccount           *systemAccountType `json:",omitempty"`
	ReportingCode           string             `json:",omitempty"`
	ReportingCodeName       string             `json:",omitempty"`
	HasAttachments          bool
	UpdatedDateUTC          string
	AddToWatchlist          bool
}

type accountForUpdate struct {
	Code              string           // Required For Creation // maxlen = 10
	Name              string           // Required For Creation // maxlen = 150
	Type              accountType      // Required For Creation
	BankAccountNumber *string          `json:",omitempty"` // Required For Creation If Type == Bank
	Description       *string          `json:",omitempty"`
	BankAccountType   *bankAccountType `json:",omitempty"`
	CurrencyCode      *string          `json:",omitempty"`
	AccountID         string
	Class             accountClassType
	SystemAccount     *systemAccountType `json:",omitempty"`
	ReportingCode     string
	ReportingCodeName string
	HasAttachments    bool
	UpdatedDateUTC    string
	AddToWatchlist    bool
}

func (a Account) toUpdate() accountForUpdate {
	return accountForUpdate{
		Code:              a.Code,
		Name:              a.Name,
		Type:              a.Type,
		BankAccountNumber: a.BankAccountNumber,
		Description:       a.Description,
		BankAccountType:   a.BankAccountType,
		CurrencyCode:      a.CurrencyCode,
		AccountID:         a.AccountID,
		Class:             a.Class,
		SystemAccount:     a.SystemAccount,
		ReportingCode:     a.ReportingCode,
		ReportingCodeName: a.ReportingCodeName,
		HasAttachments:    a.HasAttachments,
		UpdatedDateUTC:    a.UpdatedDateUTC,
		AddToWatchlist:    a.AddToWatchlist,
	}
}

type accountForCreate struct {
	Code              string      // Required For Creation // maxlen = 10
	Name              string      // Required For Creation // maxlen = 150
	Type              accountType // Required For Creation
	BankAccountNumber *string     `json:",omitempty"` // Required For Creation If Type == Bank
}

func (a Account) toCreate() accountForCreate {
	return accountForCreate{
		Code:              a.Code,
		Name:              a.Name,
		Type:              a.Type,
		BankAccountNumber: a.BankAccountNumber,
	}
}

func (a Account) validForCreation() bool {
	// validate Code
	if a.Code == "" || len(a.Code) > 10 {
		return false
	}
	// validate Name
	if a.Name == "" || len(a.Name) > 150 {
		return false
	}
	// validate Type
	if !validateAccountType(a.Type) {
		return false
	}
	if a.Type == AccountTypeBank {
		if a.BankAccountNumber == nil || len(*a.BankAccountNumber) == 0 {
			return false
		}
	}
	return true
}

func (a Account) validForUpdate() bool {
	// validate AccountID
	if a.AccountID == "" || len(a.AccountID) != len("297c2dc5-cc47-4afd-8ec8-74990b8761e9") { // figure out the number
		return false
	}
	return true
}

var (
	ErrInvalidAccountForCreation = errors.New("one or more required fields were invalid to create this account")
	ErrInvalidAccountForUpdating = errors.New("a valid account id is required to update an account")
	ErrInvalidAccountID          = errors.New("invalid account id for request")
)

type accountStatusCode string
type accountClassType string
type accountType string

type bankAccountType string
type systemAccountType string

// source: https://developer.xero.com/documentation/api/accounting/types#accounts
const (
	AccountClassAsset     accountClassType = "ASSET"
	AccountClassEquity    accountClassType = "EQUITY"
	AccountClassExpense   accountClassType = "EXPENSE"
	AccountClassLiability accountClassType = "LIABILITY"
	AccountClassRevenue   accountClassType = "REVENUE"

	AccountTypeBank                accountType = "BANK"
	AccountTypeCurrent             accountType = "CURRENT"
	AccountTypeCurrentLiability    accountType = "CURRLIAB"
	AccountTypeDepreciation        accountType = "DEPRECIATN"
	AccountTypeDirectCosts         accountType = "DIRECTCOSTS"
	AccountTypeEquity              accountType = "EQUITY"
	AccountTypeExpense             accountType = "EXPENSE"
	AccountTypeFixedAsset          accountType = "FIXED"
	AccountTypeInventoryAsset      accountType = "INVENTORY"
	AccountTypeLiability           accountType = "LIABILITY"
	AccountTypeNonCurrentAsset     accountType = "NONCURRENT"
	AccountTypeOtherIncome         accountType = "OTHERINCOME"
	AccountTypeOverHeads           accountType = "OVERHEADS"
	AccountTypePrepayment          accountType = "PREPAYMENT"
	AccountTypeRevenue             accountType = "REVENUE"
	AccountTypeSales               accountType = "SALES"
	AccountTypeNonCurrentLiability accountType = "TERMLIAB"

	AccountStatusCodeActive   accountStatusCode = "ACTIVE"
	AccountStatusCodeArchived accountStatusCode = "ARCHIVED"

	BankAccountTypeBank       bankAccountType = "BANK"
	BankAccountTypeCreditCard bankAccountType = "CREDITCARD"
	BankAccountTypePaypal     bankAccountType = "PAYPAL"

	SystemAccountTypeAccountsReceivable systemAccountType = "DEBTORS"
	// More of these
)

var (
	allAccountTypes = []accountType{
		AccountTypeBank,
		AccountTypeCurrent,
		AccountTypeCurrentLiability,
		AccountTypeDepreciation,
		AccountTypeDirectCosts,
		AccountTypeEquity,
		AccountTypeExpense,
		AccountTypeFixedAsset,
		AccountTypeInventoryAsset,
		AccountTypeLiability,
		AccountTypeNonCurrentAsset,
		AccountTypeOtherIncome,
		AccountTypeOverHeads,
		AccountTypePrepayment,
		AccountTypeRevenue,
		AccountTypeSales,
		AccountTypeNonCurrentLiability,
	}
)

func validateAccountType(accType accountType) bool {
	for _, aT := range allAccountTypes {
		if aT == accType {
			return true
		}
	}
	return false
}

func GetAccounts(tenantID, accessToken string, where *filter.Filter) (accounts []Account, err error) {
	url := endpoints.EndpointAccounts
	var request *http.Request
	if where != nil {
		request, err = where.BuildRequest("GET", url, nil)
	} else {
		request, err = http.NewRequest("GET", url, nil)
	}
	if err != nil {
		return
	}
	utils.AddXeroHeaders(request, accessToken, tenantID)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	var responseBody struct {
		Accounts []Account
	}
	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	if response.StatusCode != 200 {
		err = errors.New(string(b))
		return
	}
	err = json.Unmarshal(b, &responseBody)
	accounts = responseBody.Accounts
	return
}

func GetAccount(accountID string, tenantID, accessToken string) (acc Account, err error) {
	if len(accountID) != len("297c2dc5-cc47-4afd-8ec8-74990b8761e9") { // figure out the number
		err = ErrInvalidAccountID
		return
	}
	url := endpoints.EndpointAccounts + "/" + accountID
	var request *http.Request
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	utils.AddXeroHeaders(request, accessToken, tenantID)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	if response.StatusCode != 200 {
		err = errors.New(string(b))
		return
	}
	err = json.Unmarshal(b, &acc)
	return
}

func CreateAccount(account Account, tenantID string, accessToken string) (acc Account, err error) {
	url := endpoints.EndpointAccounts
	if !account.validForCreation() {
		err = ErrInvalidAccountForCreation
		return
	}
	b, err := json.Marshal(account.toCreate())
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(b)
	var request *http.Request
	request, err = http.NewRequest("PUT", url, buf)
	if err != nil {
		return
	}
	utils.AddXeroHeaders(request, accessToken, tenantID)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	b, err = io.ReadAll(response.Body)
	if err != nil {
		return
	}
	if response.StatusCode != 200 {
		err = errors.New(string(b))
		return
	}
	var responseBody struct {
		Accounts []Account
	}
	err = json.Unmarshal(b, &responseBody)
	if err != nil {
		return
	}
	if len(responseBody.Accounts) == 1 {
		acc = responseBody.Accounts[0]
	}
	return
}

func UpdateAccount(account Account, tenantID string, accessToken string) (err error) {
	url := endpoints.EndpointAccounts
	if !account.validForUpdate() {
		err = ErrInvalidAccountForUpdating
		return
	}
	b, err := json.Marshal(account.toUpdate())
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(b)
	var request *http.Request
	request, err = http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	utils.AddXeroHeaders(request, accessToken, tenantID)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	b, err = io.ReadAll(response.Body)
	if err != nil {
		return
	}
	if response.StatusCode != 200 {
		err = errors.New(string(b))
		return
	}
	return nil
}

func ArchiveAccount(accountID string, tenantID, accessToken string) (err error) {
	if len(accountID) != len("297c2dc5-cc47-4afd-8ec8-74990b8761e9") { // figure out the number
		err = ErrInvalidAccountID
		return
	}
	url := endpoints.EndpointAccounts + "/" + accountID
	var requestBody struct {
		AccountID string
		Status    string
	}
	requestBody.AccountID = accountID
	requestBody.Status = string(AccountStatusCodeArchived)
	b, err := json.Marshal(requestBody)
	if err != nil {
		return
	}
	buf := bytes.NewBuffer(b)
	request, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	utils.AddXeroHeaders(request, accessToken, tenantID)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	b, err = io.ReadAll(response.Body)
	if err != nil {
		return
	}
	if response.StatusCode != 200 {
		return errors.New(string(b))
	}
	return
}

// System accounts and accounts used on transactions can not be deleted using the delete method.
// If an account is not able to be deleted you can update the status to ARCHIVED using the accounts.ArchiveAccount Function
func DeleteAccount(accountID string, tenantID, accessToken string) (err error) {
	if len(accountID) != len("297c2dc5-cc47-4afd-8ec8-74990b8761e9") { // figure out the number
		err = ErrInvalidAccountID
		return
	}
	url := endpoints.EndpointAccounts + "/" + accountID
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return
	}
	utils.AddXeroHeaders(request, accessToken, tenantID)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	if response.StatusCode != 200 {
		return errors.New(string(b))
	}
	return
}

// TODO Add Attachments to Account
// TODO Get Attachments from Account
