package start

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/plugins/chconfig/common"
	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
)

// UpdateAddedFromConfigsRequest ...
type UpdateAddedFromConfigsRequest struct {
	Channel string `json:"channel,omitempty"`
	Config  []byte `json:"config,omitempty"`
	MSPID   string `json:"msp_id,omitempty"`
}

func computeAddedUpdateFromConfigs(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with read request body: %s\n", err)
		return
	}

	var req UpdateAddedFromConfigsRequest
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

	updated, err := common.ComputeAddedUpdateFromConfigs(configEnvelope.Config, updatedConifg, req.Channel, req.MSPID)
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

func generateConfigEnvelope(updated *cb.ConfigUpdate, chainID string) (*cb.Envelope, error) {
	encoded, err := proto.Marshal(updated)
	if err != nil {
		return nil, errors.Errorf("Error marshaling config update: %s", err)
	}

	cheader := &cb.ChannelHeader{ChannelId: chainID, Type: int32(cb.HeaderType_CONFIG_UPDATE)}
	cheaderBytes, err := proto.Marshal(cheader)
	if err != nil {
		return nil, errors.Errorf("Error with marshal channel header: %s", err)
	}

	cfgEnvp := &cb.ConfigUpdateEnvelope{ConfigUpdate: encoded}
	data, err := proto.Marshal(cfgEnvp)
	if err != nil {
		return nil, errors.Errorf("Error marshaling ConfigUpdateEnvelope: %s", err)
	}

	payload := &cb.Payload{
		Header: &cb.Header{ChannelHeader: cheaderBytes},
		Data:   data,
	}

	data, err = proto.Marshal(payload)
	if err != nil {
		return nil, errors.Errorf("Error marshaling Payload: %s", err)
	}

	return &cb.Envelope{Payload: data}, nil
}
