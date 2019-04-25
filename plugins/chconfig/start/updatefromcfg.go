package start

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/plugins/chconfig/common"
)

// UpdateFromConfigsRequest ...
type UpdateFromConfigsRequest struct {
	Channel string `json:"channel,omitempty"`
	Config  []byte `json:"config,omitempty"`
	MSPID   string `json:"msp_id,omitempty"`
}

func computeUpdateFromConfigs(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with read request body: %s\n", err)
		return
	}

	var req UpdateFromConfigsRequest
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

	updatedConifg, err := common.EncodeOrdererOrgGroup(req.Config)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with encode orderer org group: %s\n", err)
		return
	}

	updated, err := common.ComputeUpdateFromConfigs(configEnvelope.Config, updatedConifg, req.Channel, req.MSPID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with compute update config: %s\n", err)
		return
	}

	encoded, err := proto.Marshal(updated)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error marshaling config update: %s\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(encoded)
}
