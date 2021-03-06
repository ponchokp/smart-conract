/**
 * Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED
 *
 * This file is part of NDID software.
 *
 * NDID is the free software: you can redistribute it and/or modify it under
 * the terms of the Affero GNU General Public License as published by the
 * Free Software Foundation, either version 3 of the License, or any later
 * version.
 *
 * NDID is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the Affero GNU General Public License for more details.
 *
 * You should have received a copy of the Affero GNU General Public License
 * along with the NDID source code. If not, see https://www.gnu.org/licenses/agpl.txt.
 *
 * Please contact info@ndid.co.th for any further questions
 *
 */

package app

import (
	"encoding/json"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/tendermint/tendermint/abci/types"

	"github.com/ndidplatform/smart-contract/v4/abci/code"
	"github.com/ndidplatform/smart-contract/v4/abci/utils"
	"github.com/ndidplatform/smart-contract/v4/protos/data"
)

var modeFunctionMap = map[string]bool{
	"RegisterIdentity":          true,
	"AddIdentity":               true,
	"AddAccessor":               true,
	"RevokeAccessor":            true,
	"RevokeIdentityAssociation": true,
	"UpdateIdentityModeList":    true,
	"RevokeAndAddAccessor":      true,
}

var (
	masterNDIDKeyBytes   = []byte("MasterNDID")
	initStateKeyBytes    = []byte("InitState")
	lastBlockKeyBytes    = []byte("lastBlock")
	idpListKeyBytes      = []byte("IdPList")
	allNamespaceKeyBytes = []byte("AllNamespace")
)

const (
	keySeparator                = "|"
	nodeIDKeyPrefix             = "NodeID"
	behindProxyNodeKeyPrefix    = "BehindProxyNode"
	tokenKeyPrefix              = "Token"
	tokenPriceFuncKeyPrefix     = "TokenPriceFunc"
	serviceKeyPrefix            = "Service"
	serviceDestinationKeyPrefix = "ServiceDestination"
	approvedServiceKeyPrefix    = "ApproveKey"
	providedServicesKeyPrefix   = "ProvideService"
	refGroupCodeKeyPrefix       = "RefGroupCode"
	identityToRefCodeKeyPrefix  = "identityToRefCodeKey"
	accessorToRefCodeKeyPrefix  = "accessorToRefCodeKey"
	allowedModeListKeyPrefix    = "AllowedModeList"
	requestKeyPrefix            = "Request"
	dataSignatureKeyPrefix      = "SignData"
)

func (app *ABCIApplication) setMqAddresses(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("SetMqAddresses, Parameter: %s", param)
	var funcParam SetMqAddressesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
	value, _ := app.state.Get([]byte(nodeDetailKey), false)
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	var msqAddress []*data.MQ
	for _, address := range funcParam.Addresses {
		var msq data.MQ
		msq.Ip = address.IP
		msq.Port = address.Port
		msqAddress = append(msqAddress, &msq)
	}
	nodeDetail.Mq = msqAddress

	nodeDetailByte, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(nodeDetailKey), []byte(nodeDetailByte))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) getNodeMasterPublicKey(param string) types.ResponseQuery {
	app.logger.Infof("GetNodeMasterPublicKey, Parameter: %s", param)
	var funcParam GetNodeMasterPublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	value, _ := app.state.Get([]byte(key), true)
	var res GetNodeMasterPublicKeyResult
	if value == nil {
		valueJSON, err := json.Marshal(res)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	res.MasterPublicKey = nodeDetail.MasterPublicKey
	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(valueJSON, "success", app.state.Height)

}

func (app *ABCIApplication) getNodePublicKey(param string) types.ResponseQuery {
	app.logger.Infof("GetNodePublicKey, Parameter: %s", param)
	var funcParam GetNodePublicKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	value, _ := app.state.Get([]byte(key), true)
	var res GetNodePublicKeyResult
	if value == nil {
		valueJSON, err := json.Marshal(res)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	res.PublicKey = nodeDetail.PublicKey
	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(valueJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getNodeNameByNodeID(nodeID string) string {
	key := nodeIDKeyPrefix + keySeparator + nodeID
	value, _ := app.state.Get([]byte(key), true)
	if value == nil {
		return ""
	}
	var nodeDetail data.NodeDetail
	err := proto.Unmarshal([]byte(value), &nodeDetail)
	if err != nil {
		return ""
	}
	return nodeDetail.NodeName
}

func (app *ABCIApplication) getIdpNodes(param string) types.ResponseQuery {
	app.logger.Infof("GetIdpNodes, Parameter: %s", param)
	var funcParam GetIdpNodesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var returnNodes GetIdpNodesResult
	returnNodes.Node = make([]interface{}, 0)
	if funcParam.ReferenceGroupCode == "" && funcParam.IdentityNamespace == "" && funcParam.IdentityIdentifierHash == "" {
		idpsValue, _ := app.state.Get(idpListKeyBytes, true)
		var idpsList data.IdPList
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, idp := range idpsList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp
				nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
				if nodeDetailValue == nil {
					continue
				}
				var nodeDetail data.NodeDetail
				err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
				if err != nil {
					continue
				}
				// check node is active
				if !nodeDetail.Active {
					continue
				}
				// check Max IAL && AAL
				if !(nodeDetail.MaxIal >= funcParam.MinIal &&
					nodeDetail.MaxAal >= funcParam.MinAal) {
					continue
				}
				// Filter by node_id_list
				if len(funcParam.NodeIDList) > 0 {
					if !contains(idp, funcParam.NodeIDList) {
						continue
					}
				}
				// Filter by supported_request_message_data_url_type_list
				if len(funcParam.SupportedRequestMessageDataUrlTypeList) > 0 {
					// foundSupported := false
					supportedCount := 0
					for _, supportedType := range nodeDetail.SupportedRequestMessageDataUrlTypeList {
						if contains(supportedType, funcParam.SupportedRequestMessageDataUrlTypeList) {
							supportedCount++
						}
					}
					if supportedCount < len(funcParam.SupportedRequestMessageDataUrlTypeList) {
						continue
					}
				}
				var msqDesNode MsqDestinationNode
				msqDesNode.ID = idp
				msqDesNode.Name = nodeDetail.NodeName
				msqDesNode.MaxIal = nodeDetail.MaxIal
				msqDesNode.MaxAal = nodeDetail.MaxAal
				msqDesNode.SupportedRequestMessageDataUrlTypeList = append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...)
				returnNodes.Node = append(returnNodes.Node, msqDesNode)
			}
		}
	} else {
		refGroupCode := ""
		if funcParam.ReferenceGroupCode != "" {
			refGroupCode = funcParam.ReferenceGroupCode
		} else {
			identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
			refGroupCodeFromDB, _ := app.state.Get([]byte(identityToRefCodeKey), true)
			if refGroupCodeFromDB == nil {
				return app.ReturnQuery(nil, "not found", app.state.Height)
			}
			refGroupCode = string(refGroupCodeFromDB)
		}
		refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
		refGroupValue, _ := app.state.Get([]byte(refGroupKey), true)
		if refGroupValue == nil {
			return app.ReturnQuery(nil, "not found", app.state.Height)
		}
		var refGroup data.ReferenceGroup
		err := proto.Unmarshal(refGroupValue, &refGroup)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		for _, idp := range refGroup.Idps {
			nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp.NodeId
			nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
			if nodeDetailValue == nil {
				continue
			}
			var nodeDetail data.NodeDetail
			err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
			if err != nil {
				continue
			}
			// check node is active
			if !nodeDetail.Active {
				continue
			}
			// check Max IAL && AAL
			if !(nodeDetail.MaxIal >= funcParam.MinIal &&
				nodeDetail.MaxAal >= funcParam.MinAal) {
				continue
			}
			// check IdP has Association with Identity
			if !idp.Active {
				continue
			}
			// check Ial > min ial
			if idp.Ial < funcParam.MinIal {
				continue
			}
			// Filter by node_id_list
			if len(funcParam.NodeIDList) > 0 {
				if !contains(idp.NodeId, funcParam.NodeIDList) {
					continue
				}
			}
			// Filter by supported_request_message_data_url_type_list
			if len(funcParam.SupportedRequestMessageDataUrlTypeList) > 0 {
				// foundSupported := false
				supportedCount := 0
				for _, supportedType := range nodeDetail.SupportedRequestMessageDataUrlTypeList {
					if contains(supportedType, funcParam.SupportedRequestMessageDataUrlTypeList) {
						supportedCount++
					}
				}
				if supportedCount < len(funcParam.SupportedRequestMessageDataUrlTypeList) {
					continue
				}
			}
			// Filter by mode_list
			if len(funcParam.ModeList) > 0 {
				supportedModeCount := 0
				for _, mode := range idp.Mode {
					if containsInt32(mode, funcParam.ModeList) {
						supportedModeCount++
					}
				}
				if supportedModeCount < len(funcParam.ModeList) {
					continue
				}
			}
			var msqDesNode MsqDestinationNodeWithModeList
			msqDesNode.ID = idp.NodeId
			msqDesNode.Name = nodeDetail.NodeName
			msqDesNode.MaxIal = nodeDetail.MaxIal
			msqDesNode.MaxAal = nodeDetail.MaxAal
			msqDesNode.Ial = idp.Ial
			msqDesNode.ModeList = idp.Mode
			msqDesNode.SupportedRequestMessageDataUrlTypeList = append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...)
			returnNodes.Node = append(returnNodes.Node, msqDesNode)
		}
	}
	value, err := json.Marshal(returnNodes)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(returnNodes.Node) == 0 {
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	return app.ReturnQuery(value, "success", app.state.Height)
}

func (app *ABCIApplication) getAsNodesByServiceId(param string) types.ResponseQuery {
	app.logger.Infof("GetAsNodesByServiceId, Parameter: %s", param)
	var funcParam GetAsNodesByServiceIdParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	value, _ := app.state.Get([]byte(key), true)

	if value == nil {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}

	// filter serive is active
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, _ := app.state.Get([]byte(serviceKey), true)
	if serviceValue == nil {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if service.Active == false {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "service is not active", app.state.Height)
	}

	var storedData data.ServiceDesList
	err = proto.Unmarshal([]byte(value), &storedData)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var result GetAsNodesByServiceIdWithNameResult
	result.Node = make([]ASNodeResult, 0)
	for index := range storedData.Node {

		// filter service destination is Active
		if !storedData.Node[index].Active {
			continue
		}

		// Filter approve from NDID
		approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + storedData.Node[index].NodeId
		approveServiceJSON, _ := app.state.Get([]byte(approveServiceKey), true)
		if approveServiceJSON == nil {
			continue
		}
		var approveService data.ApproveService
		err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
		if err != nil {
			continue
		}
		if !approveService.Active {
			continue
		}

		nodeDetailKey := nodeIDKeyPrefix + keySeparator + storedData.Node[index].NodeId
		nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
		if nodeDetailValue == nil {
			continue
		}
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
		if err != nil {
			continue
		}

		// filter node is active
		if !nodeDetail.Active {
			continue
		}
		var newRow = ASNodeResult{
			storedData.Node[index].NodeId,
			nodeDetail.NodeName,
			storedData.Node[index].MinIal,
			storedData.Node[index].MinAal,
			storedData.Node[index].SupportedNamespaceList,
		}
		result.Node = append(result.Node, newRow)
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result.Node) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getMqAddresses(param string) types.ResponseQuery {
	app.logger.Infof("GetMqAddresses, Parameter: %s", param)
	var funcParam GetMqAddressesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	value, _ := app.state.Get([]byte(nodeDetailKey), true)
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal(value, &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if value == nil {
		value = []byte("[]")
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	var result GetMqAddressesResult
	for _, msq := range nodeDetail.Mq {
		var newRow MsqAddress
		newRow.IP = msq.Ip
		newRow.Port = msq.Port
		result = append(result, newRow)
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getRequest(param string, height int64) types.ResponseQuery {
	app.logger.Infof("GetRequest, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := requestKeyPrefix + keySeparator + funcParam.RequestID
	value, _ := app.state.GetVersioned([]byte(key), height, true)

	if value == nil {
		valueJSON := []byte("{}")
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	var res GetRequestResult
	res.IsClosed = request.Closed
	res.IsTimedOut = request.TimedOut
	res.MessageHash = request.RequestMessageHash
	res.Mode = request.Mode

	valueJSON, err := json.Marshal(res)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(valueJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getRequestDetail(param string, height int64, committedState bool) types.ResponseQuery {
	app.logger.Infof("GetRequestDetail, Parameter: %s", param)
	var funcParam GetRequestParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	key := requestKeyPrefix + keySeparator + funcParam.RequestID
	value, _ := app.state.GetVersioned([]byte(key), height, committedState)
	if value == nil {
		valueJSON := []byte("{}")
		return app.ReturnQuery(valueJSON, "not found", app.state.Height)
	}

	var result GetRequestDetailResult
	var request data.Request
	err = proto.Unmarshal([]byte(value), &request)
	if err != nil {
		value = []byte("")
		return app.ReturnQuery(value, err.Error(), app.state.Height)
	}

	result.RequestID = request.RequestId
	result.MinIdp = int(request.MinIdp)
	result.MinAal = float64(request.MinAal)
	result.MinIal = float64(request.MinIal)
	result.Timeout = int(request.RequestTimeout)
	result.IdPIDList = request.IdpIdList
	for _, dataRequest := range request.DataRequestList {
		var newRow DataRequest
		newRow.ServiceID = dataRequest.ServiceId
		newRow.As = dataRequest.AsIdList
		newRow.Count = int(dataRequest.MinAs)
		newRow.AnsweredAsIdList = dataRequest.AnsweredAsIdList
		newRow.ReceivedDataFromList = dataRequest.ReceivedDataFromList
		newRow.RequestParamsHash = dataRequest.RequestParamsHash
		if newRow.As == nil {
			newRow.As = make([]string, 0)
		}
		if newRow.AnsweredAsIdList == nil {
			newRow.AnsweredAsIdList = make([]string, 0)
		}
		if newRow.ReceivedDataFromList == nil {
			newRow.ReceivedDataFromList = make([]string, 0)
		}
		result.DataRequestList = append(result.DataRequestList, newRow)
	}
	result.MessageHash = request.RequestMessageHash
	for _, response := range request.ResponseList {
		var newRow Response
		newRow.Ial = float64(response.Ial)
		newRow.Aal = float64(response.Aal)
		newRow.Status = response.Status
		newRow.Signature = response.Signature
		newRow.IdpID = response.IdpId
		if response.ValidIal != "" {
			if response.ValidIal == "true" {
				tValue := true
				newRow.ValidIal = &tValue
			} else {
				fValue := false
				newRow.ValidIal = &fValue
			}
		}
		if response.ValidSignature != "" {
			if response.ValidSignature == "true" {
				tValue := true
				newRow.ValidSignature = &tValue
			} else {
				fValue := false
				newRow.ValidSignature = &fValue
			}
		}
		result.Responses = append(result.Responses, newRow)
	}
	result.IsClosed = request.Closed
	result.IsTimedOut = request.TimedOut
	result.Mode = request.Mode

	// Set purpose
	result.Purpose = request.Purpose

	// make nil to array len 0
	if result.IdPIDList == nil {
		result.IdPIDList = make([]string, 0)
	}
	if result.DataRequestList == nil {
		result.DataRequestList = make([]DataRequest, 0)
	}

	// Set requester_node_id
	result.RequesterNodeID = request.Owner

	// Set creation_block_height
	result.CreationBlockHeight = request.CreationBlockHeight

	// Set creation_chain_id
	result.CreationChainID = request.ChainId

	resultJSON, err := json.Marshal(result)
	if err != nil {
		value = []byte("")
		return app.ReturnQuery(value, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getNamespaceList(param string) types.ResponseQuery {
	app.logger.Infof("GetNamespaceList, Parameter: %s", param)
	value, _ := app.state.Get(allNamespaceKeyBytes, true)
	if value == nil {
		value = []byte("[]")
		return app.ReturnQuery(value, "not found", app.state.Height)
	}

	result := make([]*data.Namespace, 0)
	// filter flag==true
	var namespaces data.NamespaceList
	err := proto.Unmarshal([]byte(value), &namespaces)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for _, namespace := range namespaces.Namespaces {
		if namespace.Active {
			result = append(result, namespace)
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getServiceDetail(param string) types.ResponseQuery {
	app.logger.Infof("GetServiceDetail, Parameter: %s", param)
	var funcParam GetServiceDetailParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	value, _ := app.state.Get([]byte(key), true)
	if value == nil {
		value = []byte("{}")
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal(value, &service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	returnValue, err := json.Marshal(service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) updateNode(param string, nodeID string) types.ResponseDeliverTx {
	app.logger.Infof("UpdateNode, Parameter: %s", param)
	var funcParam UpdateNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	key := nodeIDKeyPrefix + keySeparator + nodeID
	value, _ := app.state.Get([]byte(key), false)
	if value == nil {
		return app.ReturnDeliverTxLog(code.NodeIDNotFound, "Node ID not found", "")
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(value), &nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.UnmarshalError, err.Error(), "")
	}
	// update MasterPublicKey
	if funcParam.MasterPublicKey != "" {
		nodeDetail.MasterPublicKey = funcParam.MasterPublicKey
	}
	// update PublicKey
	if funcParam.PublicKey != "" {
		nodeDetail.PublicKey = funcParam.PublicKey
	}
	// update SupportedRequestMessageDataUrlTypeList and Role of node ID is IdP
	if funcParam.SupportedRequestMessageDataUrlTypeList != nil && string(app.getRoleFromNodeID(nodeID)) == "IdP" {
		nodeDetail.SupportedRequestMessageDataUrlTypeList = funcParam.SupportedRequestMessageDataUrlTypeList
	}
	nodeDetailValue, err := utils.ProtoDeterministicMarshal(&nodeDetail)
	if err != nil {
		return app.ReturnDeliverTxLog(code.MarshalError, err.Error(), "")
	}
	app.state.Set([]byte(key), []byte(nodeDetailValue))
	return app.ReturnDeliverTxLog(code.OK, "success", "")
}

func (app *ABCIApplication) checkExistingIdentity(param string) types.ResponseQuery {
	app.logger.Infof("CheckExistingIdentity, Parameter: %s", param)
	var funcParam CheckExistingIdentityParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result CheckExistingIdentityResult
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "Found reference group code and identity detail in parameter", app.state.Height)
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, _ := app.state.Get([]byte(identityToRefCodeKey), true)
		if refGroupCodeFromDB == nil {
			returnValue, err := json.Marshal(result)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			return app.ReturnQuery(returnValue, "success", app.state.Height)
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, _ := app.state.Get([]byte(refGroupKey), true)
	if refGroupValue == nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "success", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "success", app.state.Height)
	}
	result.Exist = true
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getAccessorKey(param string) types.ResponseQuery {
	app.logger.Infof("GetAccessorKey, Parameter: %s", param)
	var funcParam GetAccessorKeyParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetAccessorKeyResult
	result.AccessorPublicKey = ""
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, _ := app.state.Get([]byte(accessorToRefCodeKey), true)
	if refGroupCodeFromDB == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCodeFromDB)
	refGroupValue, _ := app.state.Get([]byte(refGroupKey), true)
	if refGroupValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		for _, accessor := range idp.Accessors {
			if accessor.AccessorId == funcParam.AccessorID {
				result.AccessorPublicKey = accessor.AccessorPublicKey
				result.Active = accessor.Active
				break
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getServiceList(param string) types.ResponseQuery {
	app.logger.Infof("GetServiceList, Parameter: %s", param)
	key := "AllService"
	value, _ := app.state.Get([]byte(key), true)
	if value == nil {
		result := make([]ServiceDetail, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	result := make([]*data.ServiceDetail, 0)
	// filter flag==true
	var services data.ServiceDetailList
	err := proto.Unmarshal([]byte(value), &services)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for _, service := range services.Services {
		if service.Active {
			result = append(result, service)
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getServiceNameByServiceID(serviceID string) string {
	key := serviceKeyPrefix + keySeparator + serviceID
	value, _ := app.state.Get([]byte(key), true)
	if value == nil {
		return ""
	}
	var result ServiceDetail
	err := json.Unmarshal([]byte(value), &result)
	if err != nil {
		return ""
	}
	return result.ServiceName
}

func (app *ABCIApplication) checkExistingAccessorID(param string) types.ResponseQuery {
	app.logger.Infof("CheckExistingAccessorID, Parameter: %s", param)
	var funcParam CheckExistingAccessorIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result CheckExistingResult
	result.Exist = false
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, _ := app.state.Get([]byte(accessorToRefCodeKey), true)
	if refGroupCodeFromDB == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCodeFromDB)
	refGroupValue, _ := app.state.Get([]byte(refGroupKey), true)
	if refGroupValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		for _, accessor := range idp.Accessors {
			if accessor.AccessorId == funcParam.AccessorID {
				result.Exist = true
				break
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getNodeInfo(param string) types.ResponseQuery {
	app.logger.Infof("GetNodeInfo, Parameter: %s", param)
	var funcParam GetNodeInfoParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.NodeID
	nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
	if nodeDetailValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}

	// If node behind proxy
	if nodeDetail.ProxyNodeId != "" {
		proxyNodeID := nodeDetail.ProxyNodeId
		// Get proxy node detail
		proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
		proxyNodeDetailValue, _ := app.state.Get([]byte(proxyNodeDetailKey), true)
		if proxyNodeDetailValue == nil {
			return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
		}
		var proxyNode data.NodeDetail
		err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if nodeDetail.Role == "IdP" {
			var result GetNodeInfoResultIdPandASBehindProxy
			result.PublicKey = nodeDetail.PublicKey
			result.MasterPublicKey = nodeDetail.MasterPublicKey
			result.NodeName = nodeDetail.NodeName
			result.Role = nodeDetail.Role
			result.MaxIal = nodeDetail.MaxIal
			result.MaxAal = nodeDetail.MaxAal
			result.SupportedRequestMessageDataUrlTypeList = append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...)
			result.Proxy.NodeID = string(proxyNodeID)
			result.Proxy.NodeName = proxyNode.NodeName
			result.Proxy.PublicKey = proxyNode.PublicKey
			result.Proxy.MasterPublicKey = proxyNode.MasterPublicKey
			if proxyNode.Mq != nil {
				for _, mq := range proxyNode.Mq {
					var msq MsqAddress
					msq.IP = mq.Ip
					msq.Port = mq.Port
					result.Proxy.Mq = append(result.Proxy.Mq, msq)
				}
			}
			result.Proxy.Config = nodeDetail.ProxyConfig
			result.Active = nodeDetail.Active
			value, err := json.Marshal(result)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			return app.ReturnQuery(value, "success", app.state.Height)
		}
		var result GetNodeInfoResultRPandASBehindProxy
		result.PublicKey = nodeDetail.PublicKey
		result.MasterPublicKey = nodeDetail.MasterPublicKey
		result.NodeName = nodeDetail.NodeName
		result.Role = nodeDetail.Role
		result.Proxy.NodeID = string(proxyNodeID)
		result.Proxy.NodeName = proxyNode.NodeName
		result.Proxy.PublicKey = proxyNode.PublicKey
		result.Proxy.MasterPublicKey = proxyNode.MasterPublicKey
		if proxyNode.Mq != nil {
			for _, mq := range proxyNode.Mq {
				var msq MsqAddress
				msq.IP = mq.Ip
				msq.Port = mq.Port
				result.Proxy.Mq = append(result.Proxy.Mq, msq)
			}
		}
		result.Proxy.Config = nodeDetail.ProxyConfig
		result.Active = nodeDetail.Active
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "success", app.state.Height)
	}
	if nodeDetail.Role == "IdP" {
		var result GetNodeInfoIdPResult
		result.PublicKey = nodeDetail.PublicKey
		result.MasterPublicKey = nodeDetail.MasterPublicKey
		result.NodeName = nodeDetail.NodeName
		result.Role = nodeDetail.Role
		result.MaxIal = nodeDetail.MaxIal
		result.MaxAal = nodeDetail.MaxAal
		result.SupportedRequestMessageDataUrlTypeList = append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...)
		if nodeDetail.Mq != nil {
			for _, mq := range nodeDetail.Mq {
				var msq MsqAddress
				msq.IP = mq.Ip
				msq.Port = mq.Port
				result.Mq = append(result.Mq, msq)
			}
		}
		result.Active = nodeDetail.Active
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "success", app.state.Height)
	}
	var result GetNodeInfoResult
	result.PublicKey = nodeDetail.PublicKey
	result.MasterPublicKey = nodeDetail.MasterPublicKey
	result.NodeName = nodeDetail.NodeName
	result.Role = nodeDetail.Role
	if nodeDetail.Mq != nil {
		for _, mq := range nodeDetail.Mq {
			var msq MsqAddress
			msq.IP = mq.Ip
			msq.Port = mq.Port
			result.Mq = append(result.Mq, msq)
		}
	}
	result.Active = nodeDetail.Active
	value, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(value, "success", app.state.Height)
}

func (app *ABCIApplication) getIdentityInfo(param string) types.ResponseQuery {
	app.logger.Infof("GetIdentityInfo, Parameter: %s", param)
	var funcParam GetIdentityInfoParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetIdentityInfoResult
	if funcParam.ReferenceGroupCode != "" && funcParam.IdentityNamespace != "" && funcParam.IdentityIdentifierHash != "" {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "Found reference group code and identity detail in parameter", app.state.Height)
	}
	refGroupCode := ""
	if funcParam.ReferenceGroupCode != "" {
		refGroupCode = funcParam.ReferenceGroupCode
	} else {
		identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
		refGroupCodeFromDB, _ := app.state.Get([]byte(identityToRefCodeKey), true)
		if refGroupCodeFromDB == nil {
			returnValue, err := json.Marshal(result)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			return app.ReturnQuery(returnValue, "Reference group not found", app.state.Height)
		}
		refGroupCode = string(refGroupCodeFromDB)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
	refGroupValue, _ := app.state.Get([]byte(refGroupKey), true)
	if refGroupValue == nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "Reference group not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		returnValue, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(returnValue, "Reference group not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		if funcParam.NodeID == idp.NodeId && idp.Active {
			result.Ial = idp.Ial
			result.ModeList = idp.Mode
			break
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if result.Ial <= 0.0 {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getDataSignature(param string) types.ResponseQuery {
	app.logger.Infof("GetDataSignature, Parameter: %s", param)
	var funcParam GetDataSignatureParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	signDataKey := dataSignatureKeyPrefix + keySeparator + funcParam.NodeID + keySeparator + funcParam.ServiceID + keySeparator + funcParam.RequestID
	signDataValue, _ := app.state.Get([]byte(signDataKey), true)
	if signDataValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var result GetDataSignatureResult
	result.Signature = string(signDataValue)
	returnValue, err := json.Marshal(result)
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getServicesByAsID(param string) types.ResponseQuery {
	app.logger.Infof("GetServicesByAsID, Parameter: %s", param)
	var funcParam GetServicesByAsIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetServicesByAsIDResult
	result.Services = make([]Service, 0)
	provideServiceKey := providedServicesKeyPrefix + keySeparator + funcParam.AsID
	provideServiceValue, _ := app.state.Get([]byte(provideServiceKey), true)
	if provideServiceValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	var services data.ServiceList
	err = proto.Unmarshal([]byte(provideServiceValue), &services)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	nodeDetailKey := nodeIDKeyPrefix + keySeparator + funcParam.AsID
	nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
	if nodeDetailValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	var nodeDetail data.NodeDetail
	err = proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for index, provideService := range services.Services {
		serviceKey := serviceKeyPrefix + keySeparator + provideService.ServiceId
		serviceValue, _ := app.state.Get([]byte(serviceKey), true)
		if serviceValue == nil {
			continue
		}
		var service data.ServiceDetail
		err = proto.Unmarshal([]byte(serviceValue), &service)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		if nodeDetail.Active && service.Active {
			// Set suspended from NDID
			approveServiceKey := approvedServiceKeyPrefix + keySeparator + provideService.ServiceId + keySeparator + funcParam.AsID
			approveServiceJSON, _ := app.state.Get([]byte(approveServiceKey), true)
			if approveServiceJSON == nil {
				continue
			}
			var approveService data.ApproveService
			err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
			if err == nil {
				services.Services[index].Suspended = !approveService.Active
			}
			var newRow Service
			newRow.Active = services.Services[index].Active
			newRow.MinAal = services.Services[index].MinAal
			newRow.MinIal = services.Services[index].MinIal
			newRow.ServiceID = services.Services[index].ServiceId
			newRow.Suspended = services.Services[index].Suspended
			newRow.SupportedNamespaceList = services.Services[index].SupportedNamespaceList
			result.Services = append(result.Services, newRow)
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result.Services) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getIdpNodesInfo(param string) types.ResponseQuery {
	app.logger.Infof("GetIdpNodesInfo, Parameter: %s", param)
	var funcParam GetIdpNodesParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var returnNodes GetIdpNodesInfoResult
	returnNodes.Node = make([]interface{}, 0)
	if funcParam.ReferenceGroupCode == "" && funcParam.IdentityNamespace == "" && funcParam.IdentityIdentifierHash == "" {
		idpsValue, _ := app.state.Get(idpListKeyBytes, true)
		var idpsList data.IdPList
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, idp := range idpsList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp
				nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
				if nodeDetailValue == nil {
					continue
				}
				var nodeDetail data.NodeDetail
				err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
				if err != nil {
					continue
				}
				// check node is active
				if !nodeDetail.Active {
					continue
				}
				// check Max IAL && AAL
				if !(nodeDetail.MaxIal >= funcParam.MinIal &&
					nodeDetail.MaxAal >= funcParam.MinAal) {
					continue
				}
				// Filter by node_id_list
				if len(funcParam.NodeIDList) > 0 {
					if !contains(idp, funcParam.NodeIDList) {
						continue
					}
				}
				// Filter by supported_request_message_data_url_type_list
				if len(funcParam.SupportedRequestMessageDataUrlTypeList) > 0 {
					// foundSupported := false
					supportedCount := 0
					for _, supportedType := range nodeDetail.SupportedRequestMessageDataUrlTypeList {
						if contains(supportedType, funcParam.SupportedRequestMessageDataUrlTypeList) {
							supportedCount++
						}
					}
					if supportedCount < len(funcParam.SupportedRequestMessageDataUrlTypeList) {
						continue
					}
				}
				// If node is behind proxy
				if nodeDetail.ProxyNodeId != "" {
					proxyNodeID := nodeDetail.ProxyNodeId
					// Get proxy node detail
					proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
					proxyNodeDetailValue, _ := app.state.Get([]byte(proxyNodeDetailKey), true)
					if proxyNodeDetailValue == nil {
						return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
					}
					var proxyNode data.NodeDetail
					err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
					if err != nil {
						return app.ReturnQuery(nil, err.Error(), app.state.Height)
					}
					// Check proxy node is active
					if !proxyNode.Active {
						continue
					}
					var msqDesNode IdpNodeBehindProxy
					msqDesNode.NodeID = idp
					msqDesNode.Name = nodeDetail.NodeName
					msqDesNode.MaxIal = nodeDetail.MaxIal
					msqDesNode.MaxAal = nodeDetail.MaxAal
					msqDesNode.PublicKey = nodeDetail.PublicKey
					msqDesNode.SupportedRequestMessageDataUrlTypeList = append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...)
					msqDesNode.Proxy.NodeID = string(proxyNodeID)
					msqDesNode.Proxy.PublicKey = proxyNode.PublicKey
					if proxyNode.Mq != nil {
						for _, mq := range proxyNode.Mq {
							var msq MsqAddress
							msq.IP = mq.Ip
							msq.Port = mq.Port
							msqDesNode.Proxy.Mq = append(msqDesNode.Proxy.Mq, msq)
						}
					}
					msqDesNode.Proxy.Config = nodeDetail.ProxyConfig
					returnNodes.Node = append(returnNodes.Node, msqDesNode)
				} else {
					var msq []MsqAddress
					for _, mq := range nodeDetail.Mq {
						var msqAddress MsqAddress
						msqAddress.IP = mq.Ip
						msqAddress.Port = mq.Port
						msq = append(msq, msqAddress)
					}
					var msqDesNode IdpNode
					msqDesNode.NodeID = idp
					msqDesNode.Name = nodeDetail.NodeName
					msqDesNode.MaxIal = nodeDetail.MaxIal
					msqDesNode.MaxAal = nodeDetail.MaxAal
					msqDesNode.PublicKey = nodeDetail.PublicKey
					msqDesNode.SupportedRequestMessageDataUrlTypeList = append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...)
					msqDesNode.Mq = msq
					returnNodes.Node = append(returnNodes.Node, msqDesNode)
				}
			}
		}
	} else {
		refGroupCode := ""
		if funcParam.ReferenceGroupCode != "" {
			refGroupCode = funcParam.ReferenceGroupCode
		} else {
			identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
			refGroupCodeFromDB, _ := app.state.Get([]byte(identityToRefCodeKey), true)
			if refGroupCodeFromDB == nil {
				return app.ReturnQuery(nil, "not found", app.state.Height)
			}
			refGroupCode = string(refGroupCodeFromDB)
		}
		refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCode)
		refGroupValue, _ := app.state.Get([]byte(refGroupKey), true)
		if refGroupValue == nil {
			return app.ReturnQuery(nil, "not found", app.state.Height)
		}
		var refGroup data.ReferenceGroup
		err := proto.Unmarshal(refGroupValue, &refGroup)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		for _, idp := range refGroup.Idps {
			nodeDetailKey := nodeIDKeyPrefix + keySeparator + idp.NodeId
			nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
			if nodeDetailValue == nil {
				continue
			}
			var nodeDetail data.NodeDetail
			err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
			if err != nil {
				continue
			}
			// check node is active
			if !nodeDetail.Active {
				continue
			}
			// check Max IAL && AAL
			if !(nodeDetail.MaxIal >= funcParam.MinIal &&
				nodeDetail.MaxAal >= funcParam.MinAal) {
				continue
			}
			// check IdP has Association with Identity
			if !idp.Active {
				continue
			}
			// check Ial > min ial
			if idp.Ial < funcParam.MinIal {
				continue
			}
			// Filter by node_id_list
			if len(funcParam.NodeIDList) > 0 {
				if !contains(idp.NodeId, funcParam.NodeIDList) {
					continue
				}
			}
			// Filter by supported_request_message_data_url_type_list
			if len(funcParam.SupportedRequestMessageDataUrlTypeList) > 0 {
				// foundSupported := false
				supportedCount := 0
				for _, supportedType := range nodeDetail.SupportedRequestMessageDataUrlTypeList {
					if contains(supportedType, funcParam.SupportedRequestMessageDataUrlTypeList) {
						supportedCount++
					}
				}
				if supportedCount < len(funcParam.SupportedRequestMessageDataUrlTypeList) {
					continue
				}
			}
			// Filter by mode_list
			if len(funcParam.ModeList) > 0 {
				supportedModeCount := 0
				for _, mode := range idp.Mode {
					if containsInt32(mode, funcParam.ModeList) {
						supportedModeCount++
					}
				}
				if supportedModeCount < len(funcParam.ModeList) {
					continue
				}
			}
			// If node is behind proxy
			if nodeDetail.ProxyNodeId != "" {
				proxyNodeID := nodeDetail.ProxyNodeId
				// Get proxy node detail
				proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
				proxyNodeDetailValue, _ := app.state.Get([]byte(proxyNodeDetailKey), true)
				if proxyNodeDetailValue == nil {
					return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
				}
				var proxyNode data.NodeDetail
				err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
				if err != nil {
					return app.ReturnQuery(nil, err.Error(), app.state.Height)
				}
				// Check proxy node is active
				if !proxyNode.Active {
					continue
				}
				var msqDesNode IdpNodeBehindProxyWithModeList
				msqDesNode.NodeID = idp.NodeId
				msqDesNode.Name = nodeDetail.NodeName
				msqDesNode.MaxIal = nodeDetail.MaxIal
				msqDesNode.MaxAal = nodeDetail.MaxAal
				msqDesNode.PublicKey = nodeDetail.PublicKey
				msqDesNode.SupportedRequestMessageDataUrlTypeList = append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...)
				msqDesNode.Proxy.NodeID = string(proxyNodeID)
				msqDesNode.Proxy.PublicKey = proxyNode.PublicKey
				if proxyNode.Mq != nil {
					for _, mq := range proxyNode.Mq {
						var msq MsqAddress
						msq.IP = mq.Ip
						msq.Port = mq.Port
						msqDesNode.Proxy.Mq = append(msqDesNode.Proxy.Mq, msq)
					}
				}
				msqDesNode.Proxy.Config = nodeDetail.ProxyConfig
				msqDesNode.ModeList = idp.Mode
				returnNodes.Node = append(returnNodes.Node, msqDesNode)
			} else {
				var msq []MsqAddress
				for _, mq := range nodeDetail.Mq {
					var msqAddress MsqAddress
					msqAddress.IP = mq.Ip
					msqAddress.Port = mq.Port
					msq = append(msq, msqAddress)
				}
				var msqDesNode IdpNodeWithModeList
				msqDesNode.NodeID = idp.NodeId
				msqDesNode.Name = nodeDetail.NodeName
				msqDesNode.MaxIal = nodeDetail.MaxIal
				msqDesNode.MaxAal = nodeDetail.MaxAal
				msqDesNode.PublicKey = nodeDetail.PublicKey
				msqDesNode.SupportedRequestMessageDataUrlTypeList = append(make([]string, 0), nodeDetail.SupportedRequestMessageDataUrlTypeList...)
				msqDesNode.Mq = msq
				msqDesNode.Ial = idp.Ial
				msqDesNode.ModeList = idp.Mode
				returnNodes.Node = append(returnNodes.Node, msqDesNode)
			}
		}
	}
	value, err := json.Marshal(returnNodes)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(returnNodes.Node) == 0 {
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	return app.ReturnQuery(value, "success", app.state.Height)
}

func (app *ABCIApplication) getAsNodesInfoByServiceId(param string) types.ResponseQuery {
	app.logger.Infof("GetAsNodesInfoByServiceId, Parameter: %s", param)
	var funcParam GetAsNodesByServiceIdParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	key := serviceDestinationKeyPrefix + keySeparator + funcParam.ServiceID
	value, _ := app.state.Get([]byte(key), true)
	if value == nil {
		var result GetAsNodesInfoByServiceIdResult
		result.Node = make([]interface{}, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	// filter serive is active
	serviceKey := serviceKeyPrefix + keySeparator + funcParam.ServiceID
	serviceValue, _ := app.state.Get([]byte(serviceKey), true)
	if serviceValue == nil {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "not found", app.state.Height)
	}
	var service data.ServiceDetail
	err = proto.Unmarshal([]byte(serviceValue), &service)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if service.Active == false {
		var result GetAsNodesByServiceIdResult
		result.Node = make([]ASNode, 0)
		value, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(value, "service is not active", app.state.Height)
	}
	var storedData data.ServiceDesList
	err = proto.Unmarshal([]byte(value), &storedData)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	// Make mapping
	mapNodeIDList := map[string]bool{}
	for _, nodeID := range funcParam.NodeIDList {
		mapNodeIDList[nodeID] = true
	}
	var result GetAsNodesInfoByServiceIdResult
	result.Node = make([]interface{}, 0)
	for index := range storedData.Node {
		// filter from node_id_list
		if len(mapNodeIDList) > 0 {
			if mapNodeIDList[storedData.Node[index].NodeId] == false {
				continue
			}
		}
		// filter service destination is Active
		if !storedData.Node[index].Active {
			continue
		}
		// Filter approve from NDID
		approveServiceKey := approvedServiceKeyPrefix + keySeparator + funcParam.ServiceID + keySeparator + storedData.Node[index].NodeId
		approveServiceJSON, _ := app.state.Get([]byte(approveServiceKey), true)
		if approveServiceJSON == nil {
			continue
		}
		var approveService data.ApproveService
		err = proto.Unmarshal([]byte(approveServiceJSON), &approveService)
		if err != nil {
			continue
		}
		if !approveService.Active {
			continue
		}
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + storedData.Node[index].NodeId
		nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
		if nodeDetailValue == nil {
			continue
		}
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal(nodeDetailValue, &nodeDetail)
		if err != nil {
			continue
		}
		// filter node is active
		if !nodeDetail.Active {
			continue
		}
		// If node is behind proxy
		if nodeDetail.ProxyNodeId != "" {
			proxyNodeID := nodeDetail.ProxyNodeId
			// Get proxy node detail
			proxyNodeDetailKey := nodeIDKeyPrefix + keySeparator + string(proxyNodeID)
			proxyNodeDetailValue, _ := app.state.Get([]byte(proxyNodeDetailKey), true)
			if proxyNodeDetailValue == nil {
				return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
			}
			var proxyNode data.NodeDetail
			err = proto.Unmarshal([]byte(proxyNodeDetailValue), &proxyNode)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			// Check proxy node is active
			if !proxyNode.Active {
				continue
			}
			var as ASWithMqNodeBehindProxy
			as.NodeID = storedData.Node[index].NodeId
			as.Name = nodeDetail.NodeName
			as.MinIal = storedData.Node[index].MinIal
			as.MinAal = storedData.Node[index].MinAal
			as.PublicKey = nodeDetail.PublicKey
			as.SupportedNamespaceList = storedData.Node[index].SupportedNamespaceList
			as.Proxy.NodeID = string(proxyNodeID)
			as.Proxy.PublicKey = proxyNode.PublicKey
			if proxyNode.Mq != nil {
				for _, mq := range proxyNode.Mq {
					var msq MsqAddress
					msq.IP = mq.Ip
					msq.Port = mq.Port
					as.Proxy.Mq = append(as.Proxy.Mq, msq)
				}
			}
			as.Proxy.Config = nodeDetail.ProxyConfig
			result.Node = append(result.Node, as)
		} else {
			var msqAddress []MsqAddress
			for _, mq := range nodeDetail.Mq {
				var msq MsqAddress
				msq.IP = mq.Ip
				msq.Port = mq.Port
				msqAddress = append(msqAddress, msq)
			}
			var newRow = ASWithMqNode{
				storedData.Node[index].NodeId,
				nodeDetail.NodeName,
				storedData.Node[index].MinIal,
				storedData.Node[index].MinAal,
				nodeDetail.PublicKey,
				msqAddress,
				storedData.Node[index].SupportedNamespaceList,
			}
			result.Node = append(result.Node, newRow)
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getNodesBehindProxyNode(param string) types.ResponseQuery {
	app.logger.Infof("GetNodesBehindProxyNode, Parameter: %s", param)
	var funcParam GetNodesBehindProxyNodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetNodesBehindProxyNodeResult
	result.Nodes = make([]interface{}, 0)
	behindProxyNodeKey := "BehindProxyNode" + keySeparator + funcParam.ProxyNodeID
	behindProxyNodeValue, _ := app.state.Get([]byte(behindProxyNodeKey), true)
	if behindProxyNodeValue == nil {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return app.ReturnQuery(nil, err.Error(), app.state.Height)
		}
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	var nodes data.BehindNodeList
	nodes.Nodes = make([]string, 0)
	err = proto.Unmarshal([]byte(behindProxyNodeValue), &nodes)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	for _, node := range nodes.Nodes {
		nodeDetailKey := nodeIDKeyPrefix + keySeparator + node
		nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
		if nodeDetailValue == nil {
			continue
		}
		var nodeDetail data.NodeDetail
		err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
		if err != nil {
			continue
		}

		// Check node has proxy node ID
		if nodeDetail.ProxyNodeId == "" {
			continue
		}

		if nodeDetail.Role == "IdP" {
			var row IdPBehindProxy
			row.NodeID = node
			row.NodeName = nodeDetail.NodeName
			row.Role = nodeDetail.Role
			row.PublicKey = nodeDetail.PublicKey
			row.MasterPublicKey = nodeDetail.MasterPublicKey
			row.MaxIal = nodeDetail.MaxIal
			row.MaxAal = nodeDetail.MaxAal
			row.Config = nodeDetail.ProxyConfig
			row.SupportedRequestMessageDataUrlTypeList = nodeDetail.SupportedRequestMessageDataUrlTypeList
			result.Nodes = append(result.Nodes, row)
		} else {
			var row ASorRPBehindProxy
			row.NodeID = node
			row.NodeName = nodeDetail.NodeName
			row.Role = nodeDetail.Role
			row.PublicKey = nodeDetail.PublicKey
			row.MasterPublicKey = nodeDetail.MasterPublicKey
			row.Config = nodeDetail.ProxyConfig
			result.Nodes = append(result.Nodes, row)
		}

	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result.Nodes) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getNodeIDList(param string) types.ResponseQuery {
	app.logger.Infof("GetNodeIDList, Parameter: %s", param)
	var funcParam GetNodeIDListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetNodeIDListResult
	result.NodeIDList = make([]string, 0)
	if strings.ToLower(funcParam.Role) == "rp" {
		var rpsList data.RPList
		rpsKey := "rpList"
		rpsValue, _ := app.state.Get([]byte(rpsKey), true)
		if rpsValue != nil {
			err := proto.Unmarshal(rpsValue, &rpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, nodeID := range rpsList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
				nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
				if nodeDetailValue != nil {
					var nodeDetail data.NodeDetail
					err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
					if err != nil {
						continue
					}
					if nodeDetail.Active {
						result.NodeIDList = append(result.NodeIDList, nodeID)
					}
				}
			}
		}
	} else if strings.ToLower(funcParam.Role) == "idp" {
		var idpsList data.IdPList
		idpsValue, _ := app.state.Get(idpListKeyBytes, true)
		if idpsValue != nil {
			err := proto.Unmarshal(idpsValue, &idpsList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, nodeID := range idpsList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
				nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
				if nodeDetailValue != nil {
					var nodeDetail data.NodeDetail
					err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
					if err != nil {
						continue
					}
					if nodeDetail.Active {
						result.NodeIDList = append(result.NodeIDList, nodeID)
					}
				}
			}
		}
	} else if strings.ToLower(funcParam.Role) == "as" {
		var asList data.ASList
		asKey := "asList"
		asValue, _ := app.state.Get([]byte(asKey), true)
		if asValue != nil {
			err := proto.Unmarshal(asValue, &asList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, nodeID := range asList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
				nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
				if nodeDetailValue != nil {
					var nodeDetail data.NodeDetail
					err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
					if err != nil {
						continue
					}
					if nodeDetail.Active {
						result.NodeIDList = append(result.NodeIDList, nodeID)
					}
				}
			}
		}
	} else {
		var allList data.AllList
		allKey := "allList"
		allValue, _ := app.state.Get([]byte(allKey), true)
		if allValue != nil {
			err := proto.Unmarshal(allValue, &allList)
			if err != nil {
				return app.ReturnQuery(nil, err.Error(), app.state.Height)
			}
			for _, nodeID := range allList.NodeId {
				nodeDetailKey := nodeIDKeyPrefix + keySeparator + nodeID
				nodeDetailValue, _ := app.state.Get([]byte(nodeDetailKey), true)
				if nodeDetailValue != nil {
					var nodeDetail data.NodeDetail
					err := proto.Unmarshal([]byte(nodeDetailValue), &nodeDetail)
					if err != nil {
						continue
					}
					if nodeDetail.Active {
						result.NodeIDList = append(result.NodeIDList, nodeID)
					}
				}
			}
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if len(result.NodeIDList) == 0 {
		return app.ReturnQuery(resultJSON, "not found", app.state.Height)
	}
	return app.ReturnQuery(resultJSON, "success", app.state.Height)
}

func (app *ABCIApplication) getAccessorOwner(param string) types.ResponseQuery {
	app.logger.Infof("GetAccessorOwner, Parameter: %s", param)
	var funcParam GetAccessorOwnerParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetAccessorOwnerResult
	result.NodeID = ""
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, _ := app.state.Get([]byte(accessorToRefCodeKey), true)
	if refGroupCodeFromDB == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	refGroupKey := refGroupCodeKeyPrefix + keySeparator + string(refGroupCodeFromDB)
	refGroupValue, _ := app.state.Get([]byte(refGroupKey), true)
	if refGroupValue == nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	var refGroup data.ReferenceGroup
	err = proto.Unmarshal(refGroupValue, &refGroup)
	if err != nil {
		return app.ReturnQuery([]byte("{}"), "not found", app.state.Height)
	}
	for _, idp := range refGroup.Idps {
		for _, accessor := range idp.Accessors {
			if accessor.AccessorId == funcParam.AccessorID {
				result.NodeID = idp.NodeId
				break
			}
		}
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) isInitEnded(param string) types.ResponseQuery {
	app.logger.Infof("IsInitEnded, Parameter: %s", param)
	var result IsInitEndedResult
	result.InitEnded = false
	value, _ := app.state.Get(initStateKeyBytes, true)
	if string(value) == "false" {
		result.InitEnded = true
	}
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) getChainHistory(param string) types.ResponseQuery {
	app.logger.Infof("GetChainHistory, Parameter: %s", param)
	chainHistoryInfoKey := "ChainHistoryInfo"
	value, _ := app.state.Get([]byte(chainHistoryInfoKey), true)
	return app.ReturnQuery(value, "success", app.state.Height)
}

func contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func containsInt32(a int32, list []int32) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (app *ABCIApplication) GetReferenceGroupCode(param string) types.ResponseQuery {
	app.logger.Infof("GetReferenceGroupCode, Parameter: %s", param)
	var funcParam GetReferenceGroupCodeParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	identityToRefCodeKey := identityToRefCodeKeyPrefix + keySeparator + funcParam.IdentityNamespace + keySeparator + funcParam.IdentityIdentifierHash
	refGroupCodeFromDB, _ := app.state.Get([]byte(identityToRefCodeKey), true)
	if refGroupCodeFromDB == nil {
		refGroupCodeFromDB = []byte("")
	}
	var result GetReferenceGroupCodeResult
	result.ReferenceGroupCode = string(refGroupCodeFromDB)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	if string(refGroupCodeFromDB) == "" {
		return app.ReturnQuery(returnValue, "not found", app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetReferenceGroupCodeByAccessorID(param string) types.ResponseQuery {
	app.logger.Infof("GetReferenceGroupCodeByAccessorID, Parameter: %s", param)
	var funcParam GetReferenceGroupCodeByAccessorIDParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	accessorToRefCodeKey := accessorToRefCodeKeyPrefix + keySeparator + funcParam.AccessorID
	refGroupCodeFromDB, _ := app.state.Get([]byte(accessorToRefCodeKey), true)
	if refGroupCodeFromDB == nil {
		refGroupCodeFromDB = []byte("")
	}
	var result GetReferenceGroupCodeResult
	result.ReferenceGroupCode = string(refGroupCodeFromDB)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetAllowedModeList(param string) types.ResponseQuery {
	app.logger.Infof("GetAllowedModeList, Parameter: %s", param)
	var funcParam GetAllowedModeListParam
	err := json.Unmarshal([]byte(param), &funcParam)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	var result GetAllowedModeListResult
	result.AllowedModeList = app.GetAllowedModeFromStateDB(funcParam.Purpose, true)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetAllowedModeFromStateDB(purpose string, committedState bool) (result []int32) {
	allowedModeKey := "AllowedModeList" + keySeparator + purpose
	var allowedModeList data.AllowedModeList
	allowedModeValue, _ := app.state.Get([]byte(allowedModeKey), committedState)
	if allowedModeValue == nil {
		// return default value
		if !modeFunctionMap[purpose] {
			result = append(result, 1)
		}
		result = append(result, 2)
		result = append(result, 3)
		return result
	}
	err := proto.Unmarshal(allowedModeValue, &allowedModeList)
	if err != nil {
		return result
	}
	result = allowedModeList.Mode
	return result
}

func (app *ABCIApplication) GetNamespaceMap(committedState bool) (result map[string]bool) {
	result = make(map[string]bool, 0)
	allNamespaceValue, _ := app.state.Get(allNamespaceKeyBytes, committedState)
	if allNamespaceValue == nil {
		return result
	}
	var namespaces data.NamespaceList
	err := proto.Unmarshal([]byte(allNamespaceValue), &namespaces)
	if err != nil {
		return result
	}
	for _, namespace := range namespaces.Namespaces {
		if namespace.Active {
			result[namespace.Namespace] = true
		}
	}
	return result
}

func (app *ABCIApplication) GetNamespaceAllowedIdentifierCountMap(committedState bool) (result map[string]int) {
	result = make(map[string]int, 0)
	allNamespaceValue, _ := app.state.Get(allNamespaceKeyBytes, committedState)
	if allNamespaceValue == nil {
		return result
	}
	var namespaces data.NamespaceList
	err := proto.Unmarshal([]byte(allNamespaceValue), &namespaces)
	if err != nil {
		return result
	}
	for _, namespace := range namespaces.Namespaces {
		if namespace.Active {
			if namespace.AllowedIdentifierCountInReferenceGroup == -1 {
				result[namespace.Namespace] = 0
			} else {
				result[namespace.Namespace] = int(namespace.AllowedIdentifierCountInReferenceGroup)
			}
		}
	}
	return result
}

func (app *ABCIApplication) GetAllowedMinIalForRegisterIdentityAtFirstIdp(param string) types.ResponseQuery {
	app.logger.Infof("GetAllowedMinIalForRegisterIdentityAtFirstIdp, Parameter: %s", param)
	var result GetAllowedMinIalForRegisterIdentityAtFirstIdpResult
	result.MinIal = app.GetAllowedMinIalForRegisterIdentityAtFirstIdpFromStateDB(true)
	returnValue, err := json.Marshal(result)
	if err != nil {
		return app.ReturnQuery(nil, err.Error(), app.state.Height)
	}
	return app.ReturnQuery(returnValue, "success", app.state.Height)
}

func (app *ABCIApplication) GetAllowedMinIalForRegisterIdentityAtFirstIdpFromStateDB(committedState bool) float64 {
	allowedMinIalKey := "AllowedMinIalForRegisterIdentityAtFirstIdp"
	var allowedMinIal data.AllowedMinIalForRegisterIdentityAtFirstIdp
	allowedMinIalValue, _ := app.state.Get([]byte(allowedMinIalKey), committedState)
	if allowedMinIalValue == nil {
		return 0
	}
	err := proto.Unmarshal(allowedMinIalValue, &allowedMinIal)
	if err != nil {
		return 0
	}
	return allowedMinIal.MinIal
}
