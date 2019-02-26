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

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
	pbOrch "github.com/RTradeLtd/grpc/nexus"
)

func Test_API_Routes_IPFS_Private_Network_Management(t *testing.T) {
	type args struct {
		name          string
		swarmKey      string
		bootStrapPeer string
		user          string
		peerID        string
	}
	tests := []struct {
		name                string
		args                args
		wantStatusCreate    int
		wantStatusStart     int
		wantStatusStop      int
		wantStatusRemove    int
		createResponseError error
		startResponseError  error
		stopResponseError   error
		removeResponseError error
	}{
		{"Success", args{"testnetwork", "", "", "", "mypeer"}, 200, 200, 200, 200, nil, nil, nil, nil},
		{"Succes-Params", args{"testnetwork2", testSwarmKey, testBootstrapPeer1, "", "mypeer"}, 200, 200, 200, 200, nil, nil, nil, nil},
		{"Failure-Bad-Response-Stop", args{"testnetwork3", testSwarmKey, testBootstrapPeer2, "", "mypeer"}, 200, 200, 400, 200, nil, nil, errors.New("bad"), nil},
		{"Failure-Bad-Response-Start", args{"testnetwork4", testSwarmKey, testBootstrapPeer2, "", "mypeer"}, 200, 400, 200, 200, nil, errors.New("bad"), nil, nil},
		{"Failure-Bad-Resposne-Remove", args{"testnetwork5", testSwarmKey, testBootstrapPeer2, "", "mypeer"}, 200, 200, 200, 400, nil, nil, nil, errors.New("bad")},
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
			defer db.Close()

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
			// seutp fake returns
			fakeOrch.StartNetworkReturnsOnCall(
				0,
				&pbOrch.StartNetworkResponse{
					PeerId:   tt.args.peerID,
					SwarmKey: tt.args.swarmKey,
				},
				tt.createResponseError,
			)
			fakeOrch.StartNetworkReturnsOnCall(
				1,
				&pbOrch.StartNetworkResponse{
					PeerId:   tt.args.peerID,
					SwarmKey: tt.args.swarmKey,
				},
				tt.startResponseError,
			)
			fakeOrch.StopNetworkReturnsOnCall(0, nil, tt.stopResponseError)
			fakeOrch.RemoveNetworkReturnsOnCall(0, nil, tt.removeResponseError)
			// authentication
			header, err := loginHelper(api, testUser, testUserPass)
			if err != nil {
				t.Fatal(err)
			}
			var intAPIResponse interfaceAPIResponse
			urlValues := url.Values{}
			urlValues.Add("network_name", tt.args.name)
			if tt.args.swarmKey != "" {
				urlValues.Add("swarm_key", tt.args.swarmKey)
			}
			if tt.args.bootStrapPeer != "" {
				urlValues.Add("bootstrap_peers", tt.args.bootStrapPeer)
			}
			if err := sendRequestWithAuth(
				api, "POST", "/v2/ipfs/private/network/new", header, tt.wantStatusCreate, nil, urlValues, &intAPIResponse,
			); err != nil {
				t.Fatal(err)
			}
			intAPIResponse = interfaceAPIResponse{}
			urlValues = url.Values{}
			urlValues.Add("network_name", tt.args.name)
			if err := sendRequestWithAuth(
				api, "POST", "/v2/ipfs/private/network/stop", header, tt.wantStatusStop, nil, urlValues, &intAPIResponse,
			); err != nil {
				t.Fatal(err)
			}
			intAPIResponse = interfaceAPIResponse{}
			urlValues = url.Values{}
			urlValues.Add("network_name", tt.args.name)
			if err := sendRequestWithAuth(
				api, "POST", "/v2/ipfs/private/network/start", header, tt.wantStatusStart, nil, urlValues, &intAPIResponse,
			); err != nil {
				t.Fatal(err)
			}
			intAPIResponse = interfaceAPIResponse{}
			urlValues = url.Values{}
			urlValues.Add("network_name", tt.args.name)
			if err := sendRequestWithAuth(
				api, "DELETE", "/v2/ipfs/private/network/remove", header, tt.wantStatusRemove, nil, urlValues, &intAPIResponse,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_API_Routes_IPFS_Private(t *testing.T) {
	t.Skip("disabled pending refactor")
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

	//nm := models.NewHostedIPFSNetworkManager(db)

	// get private network information
	// /v2/ipfs/private/network/:name
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/ipfs/private/network/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /v2/ipfs/private/network/abc123")
	}

	// get all authorized private networks
	// /v2/ipfs/private/networks
	var stringSliceAPIResp stringSliceAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/ipfs/private/networks", 200, nil, nil, &stringSliceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if stringSliceAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /v2/ipfs/private/networks")
	}
	if len(stringSliceAPIResp.Response) == 0 {
		t.Fatal("failed to find any from /v2/ipfs/private/networks")
	}
	var found bool
	for _, v := range stringSliceAPIResp.Response {
		if v == "abc123" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("failed to find correct network from /v2/ipfs/private/networks")
	}

	// add a file normally
	// /v2/ipfs/private/file/add
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
	req := httptest.NewRequest("POST", "/v2/ipfs/private/file/add", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("network_name", "abc123")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /v2/ipfs/private/file/add")
	}
	apiResp := apiResponse{}
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
		t.Fatal("bad api status code from /v2/ipfs/private/file/add")
	}
	hash = apiResp.Response

	// test pinning
	// /v2/ipfs/private/pin
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/private/pin/"+hash, 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from  /v2/ipfs/private/pin")
	}

	// test pin check
	// /v2/ipfs/private/check/pin
	var boolAPIResp boolAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/ipfs/private/pin/check/"+hash+"/abc123", 200, nil, nil, &boolAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if boolAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /v2/ipfs/private/check/pin")
	}

	// test pubsub publish
	// /v2/ipfs/private/publish/topic
	mapAPIResp := mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/private/pubsub/publish/foo", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /v2/ipfs/private/pubsub/publish/topic")
	}
	if mapAPIResp.Response["topic"] != "foo" {
		t.Fatal("bad response")
	}
	if mapAPIResp.Response["message"] != "bar" {
		t.Fatal("bad response")
	}

	// test object stat
	// /v2/ipfs/private/stat
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/v2/ipfs/private/stat/"+hash+"/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /v2/ipfs/private/stat")
	}

	// test get dag
	// /v2/ipfs/private/dag
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/v2/ipfs/private/dag/"+hash+"/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /v2/ipfs/private/dag/")
	}

	// test download
	// /v2/ipfs/utils/download
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/utils/download/"+hash, 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test get authorized networks
	// /v2/ipfs/private/networks
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/v2/ipfs/private/networks", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /v2/ipfs/private/networks/")
	}

	// test get authorized networks
	// /v2/ipfs/private/networks
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/v2/ipfs/private/uploads/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /v2/ipfs/private/uploads")
	}

	// test private network beam - source private, dest public
	urlValues = url.Values{}
	urlValues.Add("source_network", "abc123")
	urlValues.Add("destination_network", "public")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test private network beam - source public, dest private
	// /v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("destination_network", "abc123")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test private network beam - source private, dest private
	// /v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "abc123")
	urlValues.Add("destination_network", "abc123")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

}
