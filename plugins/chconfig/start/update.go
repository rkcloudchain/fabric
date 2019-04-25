package start

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hyperledger/fabric/peer/common"
	"github.com/hyperledger/fabric/protos/utils"
)

// UpdateRequest ...
type UpdateRequest struct {
	Data    []byte `json:"data,omitempty"`
	Channel string `json:"channel,omitempty"`
}

func update(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with read request body: %s\n", err)
		return
	}

	var req UpdateRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with unmarshal request body: %s\n", err)
		return
	}

	ctxEnv, err := utils.UnmarshalEnvelope(req.Data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with unmarshal envelope: %s\n", err)
		return
	}

	sCtxEnv, err := sanityCheckAndSignConfigTx(ctxEnv, req.Channel)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with check and sign config tx: %s\n", err)
		return
	}

	client, err := common.GetBroadcastClient()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error getting broadcast client: %s\n", err)
		return
	}
	defer client.Close()

	err = client.Send(sCtxEnv)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error sending configuration envelope: %s\n", err)
		return
	}

	logger.Info("Successfully submitted channel update")
	w.WriteHeader(http.StatusNoContent)
}
