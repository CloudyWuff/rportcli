package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/models"

	"github.com/cloudradar-monitoring/rportcli/internal/pkg/utils"
	"github.com/stretchr/testify/assert"
)

var clientsStub = []*models.Client{
	{
		ID:       "123",
		Name:     "Client 123",
		Os:       "Windows XP",
		OsArch:   "386",
		OsFamily: "Windows",
		OsKernel: "windows",
		Hostname: "localhost",
		Ipv4:     []string{"127.0.0.1"},
		Ipv6:     nil,
		Tags:     []string{"one"},
		Version:  "1",
		Address:  "12.2.2.3",
		Tunnels: []*models.Tunnel{
			{
				ID:          "1",
				Lhost:       "localhost",
				Lport:       "80",
				Rhost:       "rhost",
				Rport:       "81",
				LportRandom: false,
				Scheme:      "https",
				ACL:         "acl123",
			},
		},
	},
	{
		ID:       "124",
		Name:     "Client 124",
		Os:       "Linux Ubuntu",
		OsArch:   "x64",
		OsFamily: "Linux",
		OsKernel: "ubuntu",
		Hostname: "localhost",
		Ipv4:     []string{"127.0.0.1", "127.0.0.2"},
		Ipv6:     nil,
		Tags:     []string{"one", "two"},
		Version:  "2",
		Address:  "12.2.2.4",
		Tunnels: []*models.Tunnel{
			{
				ID:          "1",
				Lhost:       "localhost",
				Lport:       "80",
				Rhost:       "rhost",
				Rport:       "81",
				LportRandom: false,
				Scheme:      "https",
				ACL:         "acl123",
			},
			{
				ID:          "2",
				Lhost:       "localhost",
				Lport:       "66",
				Rhost:       "somehost",
				Rport:       "67",
				LportRandom: true,
				Scheme:      "http",
				ACL:         "acl124",
			},
		},
	},
}

func TestClientsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Basic bG9nMTE2Njo1NjQzMjI=", authHeader)

		assert.Equal(t, ClientsURL+"?fields%5Bclients%5D=id%2Cname%2Ctimezone%2Ctunnels%2Caddress%2Chostname%2Cos_kernel%2Cconnection_state", r.URL.String())
		jsonEnc := json.NewEncoder(rw)
		e := jsonEnc.Encode(ClientsResponse{Data: clientsStub})
		assert.NoError(t, e)
	}))
	defer srv.Close()

	cl := New(srv.URL, &utils.StorageBasicAuth{
		AuthProvider: func() (login, pass string, err error) {
			login = "log1166"
			pass = "564322"
			return
		},
	})

	clientsResp, err := cl.Clients(context.Background())
	assert.NoError(t, err)
	if err != nil {
		return
	}

	assert.Equal(t, clientsStub, clientsResp.Data)
}

func TestGetClientsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		jsonEnc := json.NewEncoder(rw)
		e := jsonEnc.Encode(ClientsResponse{Data: clientsStub})
		assert.NoError(t, e)
	}))
	defer srv.Close()

	cl := New(srv.URL, &utils.StorageBasicAuth{
		AuthProvider: func() (login, pass string, err error) {
			login = "log1155"
			pass = "564314"
			return
		},
	})
	actualClients, err := cl.GetClients(context.Background())
	assert.NoError(t, err)

	expectedClients := make([]*models.Client, 0, len(clientsStub))

	expectedClients = append(expectedClients, clientsStub...)

	assert.Equal(t, expectedClients, actualClients)
}
