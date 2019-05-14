package start

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/plugins/chconfig/common"
)

// UpdateRemovedFromConfigsRequest ...
type UpdateRemovedFromConfigsRequest struct {
	Channel string `json:"channel,omitempty"`
	MSPID   string `json:"msp_id,omitempty"`
}

func computeRemovedUpdateFromConfigs(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with read request body: %s\n", err)
		return
	}

	var req UpdateRemovedFromConfigsRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with unmarshal request body: %s\n", err)
		return
	}

	configBlock, err := common.GetConfigBlockFromOrderer(req.Channel)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with get config block: %s\n", err)
		return
	}

	configEnvelope, err := common.ExtractConfigEnvelope(configBlock.Data.Data[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with extract config block: %s\n", err)
		return
	}

	updated, err := common.ComputeRemovedUpdateFromConfigs(configEnvelope.Config, req.Channel, req.MSPID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with compute update config: %s\n", err)
		return
	}

	envp, err := generateConfigEnvelope(updated, req.Channel)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s\n", err)
		return
	}

	data, err := proto.Marshal(envp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error marshaling Envelope: %s\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}
