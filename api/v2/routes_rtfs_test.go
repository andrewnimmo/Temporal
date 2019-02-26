package v2

import (
	"bytes"
	"errors"
	"io"
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

func Test_API_Routes_IPFS_Public_PIN(t *testing.T) {
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
			// authentication
			header, err := loginHelper(api, testUser, testUserPass)
			if err != nil {
				t.Fatal(err)
			}
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
			if err := sendRequestWithAuth(
				api, "POST", "/v2/ipfs/public/pin/"+tt.args.hash, header, tt.wantStatus, nil, urlValues, &apiResp,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_API_Routes_IPFS_Public_File_Add(t *testing.T) {

	type args struct {
		filePath string
		holdTime string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		addError   error
	}{
		{"Success", args{"../../testenv/config.json", "1"}, 200, nil},
		{"Failure-Bad-Hold-Time", args{"../../testenv/config.json", "notatime"}, 400, nil},
		{"Failure-Hold-Time-To-Long", args{"../../testenv/config.json", "1000000"}, 400, nil},
		{"Failure-Bad-Add", args{"../../testenv/config.json", "1"}, 400, errors.New("bad")},
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
			fakeManager := &mocks.FakeManager{}

			api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
			if err != nil {
				t.Fatal(err)
			}
			api.ipfs = fakeManager
			fakeManager.AddReturnsOnCall(0, "hashyboi", tt.addError)
			// authentication
			header, err := loginHelper(api, testUser, testUserPass)
			if err != nil {
				t.Fatal(err)
			}
			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)
			fileWriter, err := bodyWriter.CreateFormFile("file", tt.args.filePath)
			if err != nil {
				t.Fatal(err)
			}
			fh, err := os.Open(tt.args.filePath)
			if err != nil {
				t.Fatal(err)
			}
			defer fh.Close()
			if _, err = io.Copy(fileWriter, fh); err != nil {
				t.Fatal(err)
			}
			bodyWriter.Close()
			testRecorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/v2/ipfs/public/file/add", bodyBuf)
			req.Header.Add("Authorization", header)
			req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
			urlValues := url.Values{}
			urlValues.Add("hold_time", tt.args.holdTime)
			req.PostForm = urlValues
			api.r.ServeHTTP(testRecorder, req)
			if testRecorder.Code != tt.wantStatus {
				t.Fatal("bad http status code recovered from /v2/ipfs/public/file/add")
			}
		})
	}
}

func Test_API_Routes_IPFS_Public_Add_Directory(t *testing.T) {
	devValue := dev
	type args struct {
		filePath string
		holdTime string
	}
	tests := []struct {
		name        string
		args        args
		wantStatus  int
		addDirError error
		isDev       bool
	}{
		{"Success", args{"../../testfiles/testenv.zip", "1"}, 200, nil, true},
		{"Failure-Not-Dev", args{"../../testfiles/testenv.zip", "1"}, 400, nil, false},
		{"Failure-Not-Zip", args{"../../testenv/config.json", "1"}, 400, nil, true},
		{"Failure-Bad-Hold-Time", args{"../../testenv/config.json", "notatime"}, 400, nil, true},
		{"Failure-Hold-Time-To-Long", args{"../../testenv/config.json", "1000000"}, 400, nil, true},
		{"Failure-Bad-Add", args{"../../testfiles/testenv.zip", "1"}, 400, errors.New("bad"), true},
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
			fakeManager := &mocks.FakeManager{}

			api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
			if err != nil {
				t.Fatal(err)
			}
			api.ipfs = fakeManager
			fakeManager.AddDirReturnsOnCall(0, "hashyboi", tt.addDirError)
			// authentication
			header, err := loginHelper(api, testUser, testUserPass)
			if err != nil {
				t.Fatal(err)
			}
			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)
			fileWriter, err := bodyWriter.CreateFormFile("file", tt.args.filePath)
			if err != nil {
				t.Fatal(err)
			}
			fh, err := os.Open(tt.args.filePath)
			if err != nil {
				t.Fatal(err)
			}
			defer fh.Close()
			if _, err = io.Copy(fileWriter, fh); err != nil {
				t.Fatal(err)
			}
			bodyWriter.Close()
			testRecorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/v2/ipfs/public/file/add/directory", bodyBuf)
			req.Header.Add("Authorization", header)
			req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
			urlValues := url.Values{}
			urlValues.Add("hold_time", tt.args.holdTime)
			req.PostForm = urlValues
			dev = tt.isDev
			api.r.ServeHTTP(testRecorder, req)
			if testRecorder.Code != tt.wantStatus {
				t.Fatal("bad http status code recovered from /v2/ipfs/public/file/add")
			}
		})
	}
	// reset dev value
	dev = devValue
}

func Test_API_Routes_IPFS_Pubsub_Publish(t *testing.T) {
	type args struct {
		topic, message string
	}
	tests := []struct {
		name                string
		args                args
		wantStatus          int
		pubSubError         error
		increasePubsubcount int64
	}{
		{"Success", args{"foo", "bar"}, 200, nil, 0},
		{"Failure-Missing-Message", args{"foo", ""}, 400, nil, 0},
		{"Failure-Bad-Publish", args{"foo", "bar"}, 400, errors.New("bad"), 0},
		{"Failure-Too-Many-Pubsubs", args{"foo", "bar"}, 400, nil, 1000},
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
			// authentication
			header, err := loginHelper(api, testUser, testUserPass)
			if err != nil {
				t.Fatal(err)
			}
			urlValues := url.Values{}
			urlValues.Add("message", tt.args.message)
			var intAPIResp interfaceAPIResponse
			fakeIPFS.PubSubPublishReturnsOnCall(0, tt.pubSubError)
			if tt.increasePubsubcount != 0 {
				if err := api.usage.IncrementPubSubUsage(testUser, tt.increasePubsubcount); err != nil {
					t.Fatal(err)
				}
			}
			if err := sendRequestWithAuth(
				api, "POST", "/v2/ipfs/public/pubsub/publish/"+tt.args.topic, header, tt.wantStatus, nil, urlValues, &intAPIResp,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_API_Routes_IPFS_Public_Object_Stat(t *testing.T) {
	type args struct {
		hash string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		statError  error
	}{
		{"Success", args{goodTestPinHash}, 200, nil},
		{"Failure-Bad-Hash", args{badTestPinHash}, 400, nil},
		{"Failure-Object-Stat-Error", args{goodTestPinHash}, 400, errors.New("bad")},
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

			api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
			if err != nil {
				t.Fatal(err)
			}
			api.ipfs = fakeIPFS
			// authentication
			header, err := loginHelper(api, testUser, testUserPass)
			if err != nil {
				t.Fatal(err)
			}
			fakeIPFS.StatReturnsOnCall(0, &shell.ObjectStats{CumulativeSize: 100}, tt.statError)
			var intAPIResp interfaceAPIResponse
			if err := sendRequestWithAuth(
				api, "GET", "/v2/ipfs/public/stat/"+tt.args.hash, header, tt.wantStatus, nil, nil, &intAPIResp,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_API_Routes_IPFS_Public_Dag_Get(t *testing.T) {
	type args struct {
		hash string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
		dagError   error
	}{
		{"Success", args{goodTestPinHash}, 200, nil},
		{"Failure-Bad-Hash", args{badTestPinHash}, 400, nil},
		{"Failure-Dag-Get-Error", args{goodTestPinHash}, 400, errors.New("bad")},
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

			api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
			if err != nil {
				t.Fatal(err)
			}
			api.ipfs = fakeIPFS
			// authentication
			header, err := loginHelper(api, testUser, testUserPass)
			if err != nil {
				t.Fatal(err)
			}
			fakeIPFS.DagGetReturnsOnCall(0, tt.dagError)
			var intAPIResp interfaceAPIResponse
			if err := sendRequestWithAuth(
				api, "GET", "/v2/ipfs/public/dag/"+tt.args.hash, header, tt.wantStatus, nil, nil, &intAPIResp,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_API_Routes_IPFS_Public(t *testing.T) {
	t.Skip("temporarily disabling these tests until they are refactored")
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

	api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	// update the users tier
	if err := api.usage.UpdateTier("testuser", models.Plus); err != nil {
		t.Fatal(err)
	}

	// temporary fix for a badly written this
	// this will be solved in test refactoring
	models.NewUploadManager(db).NewUpload(
		hash, "file", models.UploadOptions{
			Username:    "testuser",
			NetworkName: "public",
			Encrypted:   false,
		},
	)

	// test download
	// /v2/ipfs/utils/download
	if err := sendRequest(
		api, "POST", "/v2/ipfs/utils/download/"+hash, 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test public network beam
	// /v2/ipfs/utils/laser/beam
	urlValues := url.Values{}
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
	apiResp := apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/ipfs/public/pin/"+hash+"/extend", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
}
