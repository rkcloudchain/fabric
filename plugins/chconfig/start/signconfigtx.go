package start

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hyperledger/fabric/common/configtx"
	localsigner "github.com/hyperledger/fabric/common/localmsp"
	"github.com/hyperledger/fabric/common/util"
	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
)

// SignConfigTxRequest ...
type SignConfigTxRequest struct {
	Data    []byte `json:"data,omitempty"`
	Channel string `json:"channel,omitempty"`
}

func signConfigTx(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error with read request body: %s\n", err)
		return
	}

	var req SignConfigTxRequest
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

	sCtxEnvData := utils.MarshalOrPanic(sCtxEnv)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(sCtxEnvData)
}

func sanityCheckAndSignConfigTx(envConfigUpdate *cb.Envelope, chainID string) (*cb.Envelope, error) {
	payload, err := utils.ExtractPayload(envConfigUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "bad payload")
	}

	if payload.Header == nil || payload.Header.ChannelHeader == nil {
		return nil, errors.New("bad header")
	}

	ch, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshall channel header")
	}

	if ch.Type != int32(cb.HeaderType_CONFIG_UPDATE) {
		return nil, errors.New("bad channel header type")
	}

	if ch.ChannelId == "" {
		return nil, errors.New("empty channel id")
	}
	if ch.ChannelId != chainID {
		return nil, errors.Errorf("mismatched channel ID %s != %s", ch.ChannelId, chainID)
	}

	configUpdateEnv, err := configtx.UnmarshalConfigUpdateEnvelope(payload.Data)
	if err != nil {
		return nil, errors.Wrap(err, "Bad config update envelope")
	}

	signer := localsigner.NewSigner()
	sigHeader, err := signer.NewSignatureHeader()
	if err != nil {
		return nil, err
	}

	configSig := &cb.ConfigSignature{
		SignatureHeader: utils.MarshalOrPanic(sigHeader),
	}

	configSig.Signature, err = signer.Sign(util.ConcatenateBytes(configSig.SignatureHeader, configUpdateEnv.ConfigUpdate))
	if err != nil {
		return nil, err
	}

	configUpdateEnv.Signatures = append(configUpdateEnv.Signatures, configSig)
	return utils.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, chainID, signer, configUpdateEnv, 0, 0)
}
