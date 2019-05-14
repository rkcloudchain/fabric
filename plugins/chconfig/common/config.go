package common

import (
	"bytes"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/common/tools/configtxlator/update"
	"github.com/hyperledger/fabric/common/tools/protolator"
	peercomm "github.com/hyperledger/fabric/peer/common"
	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("chconfig.common")

// GetConfigBlockFromOrderer gets the config block from orderer's deliver service
func GetConfigBlockFromOrderer(chainID string) (*cb.Block, error) {
	client, err := peercomm.NewDeliverClientForOrderer(chainID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	iBlock, err := client.GetNewestBlock()
	if err != nil {
		logger.Errorf("GetNewestBlock failed: %s\n", err)
		return nil, err
	}

	lc, err := utils.GetLastConfigIndexFromBlock(iBlock)
	if err != nil {
		logger.Errorf("GetLastConfigIndexFromBlock failed: %s\n", err)
		return nil, err
	}

	return client.GetSpecifiedBlock(lc)
}

// ExtractConfigEnvelope extracts configuration envelope proto
func ExtractConfigEnvelope(data []byte) (*cb.ConfigEnvelope, error) {
	envelope := &cb.Envelope{}
	if err := proto.Unmarshal(data, envelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope from config block failed")
	}

	payload := &cb.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return nil, errors.Wrap(err, "unmarshal payload from envelope failed")
	}

	channelHeader := &cb.ChannelHeader{}
	if err := proto.Unmarshal(payload.Header.ChannelHeader, channelHeader); err != nil {
		return nil, errors.Wrap(err, "unmarshal header from payload failed")
	}
	if cb.HeaderType(channelHeader.Type) != cb.HeaderType_CONFIG {
		return nil, errors.New("block must be of type 'CONFIG'")
	}

	configEnvelope := &cb.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, configEnvelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	return configEnvelope, nil
}

// EncodeOrdererOrgGroup encode org group from json
func EncodeOrdererOrgGroup(data []byte) (*cb.ConfigGroup, error) {
	reader := bytes.NewBuffer(data)
	group := &cb.DynamicConsortiumOrgGroup{ConfigGroup: &cb.ConfigGroup{}}

	if err := protolator.DeepUnmarshalJSON(reader, group); err != nil {
		return nil, errors.Wrap(err, "malformed org definition")
	}

	return group.ConfigGroup, nil
}

// ComputeAddedUpdateFromConfigs compute updated config
func ComputeAddedUpdateFromConfigs(originalConfig *cb.Config, updatedConfig *cb.ConfigGroup, chainID, mspID string) (*cb.ConfigUpdate, error) {
	updated := proto.Clone(originalConfig).(*cb.Config)
	groups := updated.ChannelGroup.Groups
	applicaton := groups["Application"]
	applicaton.Groups[mspID] = updatedConfig

	configUpdate, err := update.Compute(originalConfig, updated)
	if err != nil {
		return nil, err
	}

	configUpdate.ChannelId = chainID
	return configUpdate, nil
}

// ComputeRemovedUpdateFromConfigs compute updated config
func ComputeRemovedUpdateFromConfigs(originalConfig *cb.Config, chainID, mspID string) (*cb.ConfigUpdate, error) {
	updated := proto.Clone(originalConfig).(*cb.Config)
	groups := updated.ChannelGroup.Groups
	application := groups["Application"]

	if _, ok := application.Groups[mspID]; !ok {
		return nil, errors.Errorf("Could't find any msp id with %s", mspID)
	}

	delete(application.Groups, mspID)
	configUpdate, err := update.Compute(originalConfig, updated)
	if err != nil {
		return nil, err
	}

	configUpdate.ChannelId = chainID
	return configUpdate, nil
}
