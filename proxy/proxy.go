package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"github.com/openrelayxyz/cardinal-rpc"
	"github.com/openrelayxyz/cardinal-proxy/resolver"
	log "github.com/inconshreveable/log15"
	"net"
	"net/http"
	"time"
)

type OnMissinger interface{
	OnMissing(func(cctx *rpc.CallContext, method string, params []json.RawMessage) (interface{}, *rpc.RPCError, *rpc.CallMetadata))
}

func RegisterProxy(registry OnMissinger, backends map[string]string, defaultBackendURL string, r resolver.Resolver) {
	client := &http.Client{Transport:&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost:   16,
		MaxIdleConns:          16,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}}
	registry.OnMissing(func(cctx *rpc.CallContext, method string, params []json.RawMessage) (interface{}, *rpc.RPCError, *rpc.CallMetadata) {
		bn := r.Resolve(cctx, method, params)

		backendURL := defaultBackendURL
		if bn == nil {
			// log.Debug("Unable to resolve backend", "method", method, "params", params, "default", defaultBackendURL)
			bn = new(string)
			*bn = "<default>"
		} else {
			if bu, ok := backends[*bn]; ok {
				backendURL = bu
			} else {
				log.Warn("Resolved backend unavailable", "backend", *bn)
			}
		}

		call, _ := rpc.NewCallParams(method, params)
		callBytes, _ := json.Marshal(call)

		request, _ := http.NewRequest("POST", backendURL, bytes.NewReader(callBytes))
		request.Header.Add("Content-Type", "application/json")
		log.Debug("Proxying request", "method", "POST", "backend", *bn, "url", backendURL, "headers", request.Header)

		resp, err := client.Do(request)
		if err != nil {
			return nil, rpc.NewRPCError(-32503, err.Error()), cctx.Metadata()
		}
		defer resp.Body.Close()
		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, rpc.NewRPCError(-32504, err.Error()), cctx.Metadata()
		}
		var response rpc.Response
		if err := json.Unmarshal(result, &response); err != nil {
			return nil, rpc.NewRPCError(-32500, err.Error()), cctx.Metadata()
		}
		return response.Result, response.Error, cctx.Metadata()
	})
}
