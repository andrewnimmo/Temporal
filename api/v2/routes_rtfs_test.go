package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/c2h5oh/datasize"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database/models"
	shell "github.com/RTradeLtd/go-ipfs-api"
)

const (
	goodTestPinHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
	badTestPinHash  = "QmnotARealHash"
)

func Test_API_Routes_IPFS_PIN(t *testing.T) {
	type args struct {
		hash            string
		holdTime        string
		size            int
		firstStatError  error
		secondStatError error
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{"Success", args{goodTestPinHash, "1", 100000, nil, nil}, 200},
		{"Failure-Bad-Hash", args{badTestPinHash, "1", 100000, nil, nil}, 400},
		{"Failure-Bad-Hold-Time", args{goodTestPinHash, "bilboisnottime", 10000, nil, nil}, 400},
		{"Failure-Bad-Hold-Time-Length", args{goodTestPinHash, "10000000", 10000, nil, nil}, 400},
		{"Failure-Size-To-Big", args{goodTestPinHash, "1", int(datasize.TB.Bytes()), nil, nil}, 400},
		{"Failure-Object-Stat-Error", args{goodTestPinHash, "1", 1000, errors.New("bad"), nil}, 400},
		{"Failure-Object-Stat-Error", args{goodTestPinHash, "1", 1000, nil, errors.New("bad")}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// load configuration
			cfg, err := config.LoadConfig("../../testenv/config.json")
			if err != nil {
				t.Fatal(err)
			}
			db, err := loadDatabase(cfg)
			if err != nil {
				t.Fatal(err)
			}

			// setup fake mock clients
			fakeLens := &mocks.FakeLensV2Client{}
			fakeOrch := &mocks.FakeServiceClient{}
			fakeSigner := &mocks.FakeSignerClient{}
			fakeIPFS := &mocks.FakeManager{}
			// setup fake api
			api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
			if err != nil {
				t.Fatal(err)
			}
			// load fake rtfs
			api.ipfs = fakeIPFS
			// a successful response needs to setup 2 mock stat calls
			fakeIPFS.StatReturnsOnCall(0, &shell.ObjectStats{
				CumulativeSize: tt.args.size,
			}, tt.args.firstStatError)
			fakeIPFS.StatReturnsOnCall(1, &shell.ObjectStats{
				CumulativeSize: tt.args.size,
			}, tt.args.secondStatError)
			urlValues := url.Values{}
			urlValues.Add("hold_time", tt.args.holdTime)
			var apiResp apiResponse
			if err := sendRequest(
				api, "POST", "/v2/ipfs/public/pin/"+tt.args.hash, tt.wantStatus, nil, urlValues, &apiResp,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_API_Routes_IPFS_Public(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// setup fake mock clients
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	api, testRecorder, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	// update the users tier
	if err := api.usage.UpdateTier("testuser", models.Plus); err != nil {
		t.Fatal(err)
	}

	// add a file normally
	// /v2/ipfs/public/file/add
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", "../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	fh, err := os.Open("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	if _, err = io.Copy(fileWriter, fh); err != nil {
		t.Fatal(err)
	}
	bodyWriter.Close()
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v2/ipfs/public/file/add", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /v2/ipfs/public/file/add")
	}
	var apiResp apiResponse
	// unmarshal the response
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/ipfs/public/file/add")
	}
	hash = apiResp.Response

	// temporary fix for a badly written this
	// this will be solved in test refactoring
	models.NewUploadManager(db).NewUpload(
		hash, "file", models.UploadOptions{
			Username:    "testuser",
			NetworkName: "public",
			Encrypted:   false,
		},
	)
	// add a zip file (only do a partial test since this is being weird)
	// /v2/ipfs/public/file/add/directory
	bodyBuf = &bytes.Buffer{}
	bodyWriter = multipart.NewWriter(bodyBuf)
	fileWriter, err = bodyWriter.CreateFormFile("file", "../../testfiles/testenv.zip")
	if err != nil {
		t.Fatal(err)
	}
	fh, err = os.Open("../../testfiles/testenv.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	if _, err = io.Copy(fileWriter, fh); err != nil {
		t.Fatal(err)
	}
	bodyWriter.Close()
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/v2/ipfs/public/file/add/directory", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test pubsub publish (success)
	// /v2/ipfs/pubsub/publish/topic
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	var mapAPIResp mapAPIResponse
	if err := sendRequest(
		api, "POST", "/v2/ipfs/public/pubsub/publish/foo", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /v2/pubsub/publish/topic")
	}
	if mapAPIResp.Response["topic"] != "foo" {
		t.Fatal("bad response")
	}
	if mapAPIResp.Response["message"] != "bar" {
		t.Fatal("bad response")
	}

	// test pubsub publish (fail)
	// /v2/ipfs/pubsub/publish/topic
	urlValues = url.Values{}
	urlValues.Add("message", "")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/ipfs/public/pubsub/publish/foo", 400, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	if apiResp.Code != 400 {
		t.Fatal("bad response status code from /v2/ipfs/public/pubsub/publish")
	}

	// test object stat (success)
	// /v2/ipfs/stat
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/ipfs/public/stat/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /v2/ipfs/public/stat")
	}

	// test object stat (fail)
	// /v2/ipfs/stat
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/v2/ipfs/public/stat/notarealhash", 400, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 400 {
		t.Fatal("bad response status code from /v2/ipfs/public/stat")
	}

	// test get dag (success)
	// /v2/ipfs/public/dag
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/v2/ipfs/public/dag/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /v2/ipfs/public/dag/")
	}

	// test get dag (fail)
	// /v2/ipfs/public/dag
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/v2/ipfs/public/dag/notarealhash", 400, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 400 {
		t.Fatal("bad response status code from /v2/ipfs/public/dag/")
	}

	// test download
	// /v2/ipfs/utils/download
	if err := sendRequest(
		api, "POST", "/v2/ipfs/utils/download/"+hash, 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test public network beam
	// /v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("destination_network", "public")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}
	// test extend pin
	// /v2/ipfs/public/pin/:hash/extend
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/ipfs/public/pin/"+hash+"/extend", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
}
