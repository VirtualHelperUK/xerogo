package accounting_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/datag8r/xerogo/accountingAPI/contacts"
)

func TestGetContacts(t *testing.T) {
	conf, token, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	c, pData, err := contacts.GetContacts(conf.TenantID, token.AccessToken, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if pData != nil {
		t.Fatal(*pData)
	}
	if len(c) == 0 {
		t.Fatal("len of contacts is 0")
	}
}
func TestGetContactsWithPagination(t *testing.T) {
	conf, token, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	var currentPage uint = 1
	var lastCall time.Time = time.Now()
	var rateLimit time.Duration = time.Second
	var contactList = []contacts.Contact{}
	for {
		// rate limit handling
		next := lastCall.Add(rateLimit) // Min Time Of Next Call
		current := time.Now()
		toWait := next.Sub(current)
		<-time.After(toWait)
		lastCall = time.Now()
		// actual fetching
		c, pData, err := contacts.GetContacts(conf.TenantID, token.AccessToken, &currentPage, nil)
		if err != nil {
			t.Fatal(err)
		}
		if pData == nil {
			t.Fatal("no pageData returned")
		}
		fmt.Println("called page:", currentPage, "/", pData.PageCount)
		contactList = append(contactList, c...)
		if pData.PageCount == currentPage {
			break
		}
		currentPage++
	}
	if len(contactList) == 0 {
		t.Fatal("len of contactList is 0")
	}
}

func TestGetContact(t *testing.T) {
	conf, token, err := setup()
	if err != nil {
		t.Fatal(err)
	}
	id := "a852a44c-3d8f-4c4b-a628-3a2c2121b9b1"
	c, err := contacts.GetContact(conf.TenantID, token.AccessToken, id)
	if err != nil {
		t.Fatal(err)
	}
	if c.ContactID != id {
		t.Fatal("mismatched ids", c.ContactID)
	}
}
