package v2

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
)

func Test_API_Routes_Payments_Dash(t *testing.T) {
	// this test isn't fully implemented due to dependencies on chainrider
	type args struct {
		creditValue string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{"Success", args{"10"}, 200},
	}
	for _, tt := range tests {
		// load configuration
		cfg, err := config.LoadConfig("../../testenv/config.json")
		if err != nil {
			t.Fatal(err)
		}
		db, err := loadDatabase(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		// setup fake mock clients
		fakeLens := &mocks.FakeLensV2Client{}
		fakeOrch := &mocks.FakeServiceClient{}
		fakeSigner := &mocks.FakeSignerClient{}

		api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
		if err != nil {
			t.Fatal(err)
		}
		testRecorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v2/payments/dash/create", nil)
		req.Header.Add("Authorization", authHeader)
		urlValues := url.Values{}
		urlValues.Add("credit_value", tt.args.creditValue)
		req.PostForm = urlValues
		api.r.ServeHTTP(testRecorder, req)
	}
}

func Test_API_Routes_Payments_ETH_RTC(t *testing.T) {
	// TODO: fully implement
	type args struct {
		creditValue   string
		senderAddress string
		paymentType   string
		paymentNumber string
		txHash        string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{"Success-ETH", args{"10", "0x0", "eth", "1", "0x0"}, 200},
		{"Success-RTC", args{"10", "0x0", "rtc", "1", "0x0"}, 200},
	}
	for _, tt := range tests {
		// load configuration
		cfg, err := config.LoadConfig("../../testenv/config.json")
		if err != nil {
			t.Fatal(err)
		}
		db, err := loadDatabase(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		// setup fake mock clients
		fakeLens := &mocks.FakeLensV2Client{}
		fakeOrch := &mocks.FakeServiceClient{}
		fakeSigner := &mocks.FakeSignerClient{}

		api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
		if err != nil {
			t.Fatal(err)
		}
		header, err := loginHelper(api, testUser, testUserPass)
		if err != nil {
			t.Fatal(err)
		}
		// test the payment request
		testRecorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v2/payments/eth/request", nil)
		req.Header.Add("Authorization", header)
		urlValues := url.Values{}
		urlValues.Add("credit_value", tt.args.creditValue)
		urlValues.Add("payment_type", tt.args.paymentType)
		urlValues.Add("sender_address", tt.args.senderAddress)
		req.PostForm = urlValues
		api.r.ServeHTTP(testRecorder, req)
		// test the payment confirm
		testRecorder = httptest.NewRecorder()
		req = httptest.NewRequest("POST", "/v2/payments/eth/confirm", nil)
		req.Header.Add("Authorization", header)
		urlValues = url.Values{}
		urlValues.Add("payment_number", tt.args.paymentNumber)
		urlValues.Add("tx_hash", tt.args.txHash)
		req.PostForm = urlValues
		api.r.ServeHTTP(testRecorder, req)
	}
}
