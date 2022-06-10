package control

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/clientmodel"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	influxdb2 "github.com/influxdata/influxdb-client-go"
)

type Control struct {
	RMR    chan *xapp.RMRParams //channel for receiving rmr message
	client influxdb2.Client     //client for influxdb
}

var (
	timeToWait           = "w10ms"
	subsequentActionType = "continue"
	actionType           = "report"
	actionId             = int64(1)
	seqId                = int64(1)
	funcId               = int64(0)
	hPort                = int64(8080)
	rPort                = int64(4560)
	clientEndpoint       = clientmodel.SubscriptionParamsClientEndpoint{Host: "service-ricxapp-xappkpimon-http.ricxapp", HTTPPort: &hPort, RMRPort: &rPort}
)

func (c Control) Consume(msg *xapp.RMRParams) error {
	id := xapp.Rmr.GetRicMessageName(msg.Mtype)
	xapp.Logger.Info(
		"Message received: name=%s meid=%s subId=%d txid=%s len=%d",
		id,
		msg.Meid.RanName,
		msg.SubId,
		msg.Xid,
		msg.PayloadLen,
	)
	c.RMR <- msg
	return nil
}

func NewControl() Control {
	xapp.Logger.Info("In new control\n")
	create_db()
	return Control{
		make(chan *xapp.RMRParams),
		influxdb2.NewClient("http://ricplt-influxdb.ricplt:8086", "client"),
	}
}
func create_db() {
	//Create a database named kpimon in influxDB
	_, err := http.Post("http://ricplt-influxdb.ricplt:8086/query?q=create%20database%20kpimon", "", nil)
	if err != nil {
		xapp.Logger.Error("Create database failed!")
	}
}

func (c Control) getEnbList() ([]*xapp.RNIBNbIdentity, error) {
	enbs, err := xapp.Rnib.GetListEnbIds()
	if err != nil {
		xapp.Logger.Error("err: %s", err)
		return nil, err
	}

	xapp.Logger.Info("List for connected eNBs :")
	for index, enb := range enbs {
		xapp.Logger.Info("%d. enbid: %s", index+1, enb.InventoryName)
	}
	return enbs, nil
}

func (c *Control) getGnbList() ([]*xapp.RNIBNbIdentity, error) {
	gnbs, err := xapp.Rnib.GetListGnbIds()

	if err != nil {
		xapp.Logger.Error("err: %s", err)
		return nil, err
	}
	xapp.Logger.Info("List of connected gNBs :")
	for index, gnb := range gnbs {
		xapp.Logger.Info("%d. gnbid : %s", index+1, gnb.InventoryName)
	}
	return gnbs, nil
}

func (c *Control) getnbList() []*xapp.RNIBNbIdentity {
	//Get all GnodeB and EnodeB connected to RIC
	var nbs []*xapp.RNIBNbIdentity

	if enbs, err := c.getEnbList(); err == nil {
		nbs = append(nbs, enbs...)
	}

	if gnbs, err := c.getGnbList(); err == nil {
		nbs = append(nbs, gnbs...)
	}
	return nbs
}

func (c Control) sendSubscription(meid string) {
	//Create Subscription message and send it to RIC platform
	xapp.Logger.Info("Sending subscription request for MEID: %v", meid)

	subscritionParams := clientmodel.SubscriptionParams{
		ClientEndpoint: &clientEndpoint,
		Meid:           &meid,
		RANFunctionID:  &funcId,
		SubscriptionDetails: clientmodel.SubscriptionDetailsList{
			&clientmodel.SubscriptionDetail{
				EventTriggers: clientmodel.EventTriggerDefinition{
					1234,
				},
				XappEventInstanceID: &seqId,
				ActionToBeSetupList: clientmodel.ActionsToBeSetup{
					&clientmodel.ActionToBeSetup{
						ActionID:   &actionId,
						ActionType: &actionType,
						ActionDefinition: clientmodel.ActionDefinition{
							5678,
						},
						SubsequentAction: &clientmodel.SubsequentAction{
							SubsequentActionType: &subsequentActionType,
							TimeToWait:           &timeToWait,
						},
					},
				},
			},
		},
	}

	b, err := json.MarshalIndent(subscritionParams, "", " ")
	if err != nil {
		xapp.Logger.Error("Json marshaling failed: %v", err)
	}
	xapp.Logger.Info("*****body: %s", string(b))

	resp, err := xapp.Subscription.Subscribe(&subscritionParams)
	if err != nil {
		xapp.Logger.Error("Subscription (%s) failed  with error: %s", meid, err)
		return
	}
	xapp.Logger.Info("Successfully subscription done (%s), subscriptrion id: %s", meid, *resp.SubscriptionID)
}

func (c *Control) controlLoop() {
	//Handle receiving message based on message type
	for {
		msg := <-c.RMR
		xapp.Logger.Debug("Received message type: %d", msg.Mtype)
		switch msg.Mtype {
		case xapp.RIC_INDICATION:
			go c.handleIndication(msg)
		default:
			xapp.Logger.Error("Unknown Message Type '%d', discarding", msg.Mtype)
		}
	}
}
func (c *Control) handleIndication(params *xapp.RMRParams) (err error) {
	var e2ap *E2ap
	var e2sm *E2sm

	//Decode message and put it into log
	indicationMsg, err := e2ap.GetIndicationMessage(params.Payload)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication message: %v", err)
		log.Printf("Failed to decode RIC Indication message: %v", err)
		return
	}

	log.Printf("RIC Indication message from {%s} received", params.Meid.RanName)
	log.Printf("RequestID: %d", indicationMsg.RequestID)
	log.Printf("RequestSequenceNumber: %d", indicationMsg.RequestSequenceNumber)
	log.Printf("FunctionID: %d", indicationMsg.FuncID)
	log.Printf("ActionID: %d", indicationMsg.ActionID)
	log.Printf("IndicationSN: %d", indicationMsg.IndSN)
	log.Printf("IndicationType: %d", indicationMsg.IndType)
	log.Printf("IndicationHeader: %x", indicationMsg.IndHeader)
	log.Printf("IndicationMessage: %x", indicationMsg.IndMessage)
	log.Printf("CallProcessID: %x", indicationMsg.CallProcessID)

	indicationHdr, err := e2sm.GetIndicationHeader(indicationMsg.IndHeader)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication Header: %v", err)
		log.Printf("Failed to decode RIC Indication Header: %v", err)
		return
	}

	var cellIDHdr string
	var plmnIDHdr string
	var sliceIDHdr int32
	var fiveQIHdr int64

	//Decoding Ric Indication Header
	log.Printf("-----------RIC Indication Header-----------")
	if indicationHdr.IndHdrType == 1 {
		log.Printf("RIC Indication Header Format: %d", indicationHdr.IndHdrType)
		indHdrFormat1 := indicationHdr.IndHdr.(*IndicationHeaderFormat1)

		log.Printf("GlobalKPMnodeIDType: %d", indHdrFormat1.GlobalKPMnodeIDType)

		if indHdrFormat1.GlobalKPMnodeIDType == 1 {
			globalKPMnodegNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodegNBIDType)

			globalgNBID := globalKPMnodegNBID.GlobalgNBID

			log.Printf("PlmnID: %x", globalgNBID.PlmnID.Buf)
			log.Printf("gNB ID Type: %d", globalgNBID.GnbIDType)
			if globalgNBID.GnbIDType == 1 {
				gNBID := globalgNBID.GnbID.(*GNBID)
				log.Printf("gNB ID ID: %x, Unused: %d", gNBID.Buf, gNBID.BitsUnused)
			}

			if globalKPMnodegNBID.GnbCUUPID != nil {
				log.Printf("gNB-CU-UP ID: %x", globalKPMnodegNBID.GnbCUUPID.Buf)
			}

			if globalKPMnodegNBID.GnbDUID != nil {
				log.Printf("gNB-DU ID: %x", globalKPMnodegNBID.GnbDUID.Buf)
			}
		} else if indHdrFormat1.GlobalKPMnodeIDType == 2 {
			globalKPMnodeengNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodeengNBIDType)

			log.Printf("PlmnID: %x", globalKPMnodeengNBID.PlmnID.Buf)
			log.Printf("en-gNB ID Type: %d", globalKPMnodeengNBID.GnbIDType)
			if globalKPMnodeengNBID.GnbIDType == 1 {
				engNBID := globalKPMnodeengNBID.GnbID.(*ENGNBID)
				log.Printf("en-gNB ID ID: %x, Unused: %d", engNBID.Buf, engNBID.BitsUnused)
			}
		} else if indHdrFormat1.GlobalKPMnodeIDType == 3 {
			globalKPMnodengeNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodengeNBIDType)

			log.Printf("PlmnID: %x", globalKPMnodengeNBID.PlmnID.Buf)
			log.Printf("ng-eNB ID Type: %d", globalKPMnodengeNBID.EnbIDType)
			if globalKPMnodengeNBID.EnbIDType == 1 {
				ngeNBID := globalKPMnodengeNBID.EnbID.(*NGENBID_Macro)
				log.Printf("ng-eNB ID ID: %x, Unused: %d", ngeNBID.Buf, ngeNBID.BitsUnused)
			} else if globalKPMnodengeNBID.EnbIDType == 2 {
				ngeNBID := globalKPMnodengeNBID.EnbID.(*NGENBID_ShortMacro)
				log.Printf("ng-eNB ID ID: %x, Unused: %d", ngeNBID.Buf, ngeNBID.BitsUnused)
			} else if globalKPMnodengeNBID.EnbIDType == 3 {
				ngeNBID := globalKPMnodengeNBID.EnbID.(*NGENBID_LongMacro)
				log.Printf("ng-eNB ID ID: %x, Unused: %d", ngeNBID.Buf, ngeNBID.BitsUnused)
			}
		} else if indHdrFormat1.GlobalKPMnodeIDType == 4 {
			globalKPMnodeeNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodeeNBIDType)

			log.Printf("PlmnID: %x", globalKPMnodeeNBID.PlmnID.Buf)
			log.Printf("eNB ID Type: %d", globalKPMnodeeNBID.EnbIDType)
			if globalKPMnodeeNBID.EnbIDType == 1 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_Macro)
				log.Printf("eNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			} else if globalKPMnodeeNBID.EnbIDType == 2 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_Home)
				log.Printf("eNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			} else if globalKPMnodeeNBID.EnbIDType == 3 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_ShortMacro)
				log.Printf("eNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			} else if globalKPMnodeeNBID.EnbIDType == 4 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_LongMacro)
				log.Printf("eNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			}

		}

		if indHdrFormat1.NRCGI != nil {

			log.Printf("nRCGI.PlmnID: %x", indHdrFormat1.NRCGI.PlmnID.Buf)
			log.Printf("nRCGI.NRCellID ID: %x, Unused: %d", indHdrFormat1.NRCGI.NRCellID.Buf, indHdrFormat1.NRCGI.NRCellID.BitsUnused)

			cellIDHdr, err = e2sm.ParseNRCGI(*indHdrFormat1.NRCGI)
			if err != nil {
				xapp.Logger.Error("Failed to parse NRCGI in RIC Indication Header: %v", err)
				log.Printf("Failed to parse NRCGI in RIC Indication Header: %v", err)
				return
			}
		} else {
			cellIDHdr = ""
		}

		if indHdrFormat1.PlmnID != nil {
			log.Printf("PlmnID: %x", indHdrFormat1.PlmnID.Buf)

			plmnIDHdr, err = e2sm.ParsePLMNIdentity(indHdrFormat1.PlmnID.Buf, indHdrFormat1.PlmnID.Size)
			if err != nil {
				xapp.Logger.Error("Failed to parse PlmnID in RIC Indication Header: %v", err)
				log.Printf("Failed to parse PlmnID in RIC Indication Header: %v", err)
				return
			}
		} else {
			plmnIDHdr = ""
		}

		if indHdrFormat1.SliceID != nil {
			log.Printf("SST: %x", indHdrFormat1.SliceID.SST.Buf)

			if indHdrFormat1.SliceID.SD != nil {
				log.Printf("SD: %x", indHdrFormat1.SliceID.SD.Buf)
			}

			sliceIDHdr, err = e2sm.ParseSliceID(*indHdrFormat1.SliceID)
			if err != nil {
				xapp.Logger.Error("Failed to parse SliceID in RIC Indication Header: %v", err)
				log.Printf("Failed to parse SliceID in RIC Indication Header: %v", err)
				return
			}
		} else {
			sliceIDHdr = -1
		}

		if indHdrFormat1.FiveQI != -1 {
			log.Printf("5QI: %d", indHdrFormat1.FiveQI)
		}
		fiveQIHdr = indHdrFormat1.FiveQI

		if indHdrFormat1.Qci != -1 {
			log.Printf("QCI: %d", indHdrFormat1.Qci)
		}

		if indHdrFormat1.UeMessageType != -1 {
			log.Printf("Ue Report type: %d", indHdrFormat1.UeMessageType)
		}

		if indHdrFormat1.GnbDUID != nil {
			log.Printf("gNB-DU-ID: %x", indHdrFormat1.GnbDUID.Buf)
		}

		if indHdrFormat1.GnbNameType == 1 {
			log.Printf("gNB-DU-Name: %x", (indHdrFormat1.GnbName.(*GNB_DU_Name)).Buf)
		} else if indHdrFormat1.GnbNameType == 2 {
			log.Printf("gNB-CU-CP-Name: %x", (indHdrFormat1.GnbName.(*GNB_CU_CP_Name)).Buf)
		} else if indHdrFormat1.GnbNameType == 3 {
			log.Printf("gNB-CU-UP-Name: %x", (indHdrFormat1.GnbName.(*GNB_CU_UP_Name)).Buf)
		}

		if indHdrFormat1.GlobalgNBID != nil {
			log.Printf("PlmnID: %x", indHdrFormat1.GlobalgNBID.PlmnID.Buf)
			log.Printf("gNB ID Type: %d", indHdrFormat1.GlobalgNBID.GnbIDType)
			if indHdrFormat1.GlobalgNBID.GnbIDType == 1 {
				gNBID := indHdrFormat1.GlobalgNBID.GnbID.(*GNBID)
				log.Printf("gNB ID ID: %x, Unused: %d", gNBID.Buf, gNBID.BitsUnused)
			}
		}

	} else {
		xapp.Logger.Error("Unknown RIC Indication Header Format: %d", indicationHdr.IndHdrType)
		log.Printf("Unknown RIC Indication Header Format: %d", indicationHdr.IndHdrType)
		return
	}

	indMsg, err := e2sm.GetIndicationMessage(indicationMsg.IndMessage)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication Message: %v", err)
		log.Printf("Failed to decode RIC Indication Message: %v", err)
		return
	}

	var flag bool
	var containerType int32
	var timestampPDCPBytes *Timestamp
	var dlPDCPBytes int64
	var ulPDCPBytes int64
	var timestampPRB *Timestamp
	var availPRBDL int64
	var availPRBUL int64
	//Decoding RIC Indication Message
	log.Printf("-----------RIC Indication Message-----------")
	log.Printf("StyleType: %d", indMsg.StyleType)
	if indMsg.IndMsgType == 1 {
		log.Printf("RIC Indication Message Format: %d", indMsg.IndMsgType)

		indMsgFormat1 := indMsg.IndMsg.(*IndicationMessageFormat1)

		log.Printf("PMContainerCount: %d", indMsgFormat1.PMContainerCount)

		for PMContainerCounter := 0; PMContainerCounter < indMsgFormat1.PMContainerCount; PMContainerCounter++ {
			flag = false
			timestampPDCPBytes = nil
			dlPDCPBytes = -1
			ulPDCPBytes = -1
			timestampPRB = nil
			availPRBDL = -1
			availPRBUL = -1

			log.Printf("PMContainer[%d]: ", PMContainerCounter)

			pmContainer := indMsgFormat1.PMContainers[PMContainerCounter]

			if pmContainer.PFContainer != nil {
				containerType = pmContainer.PFContainer.ContainerType

				log.Printf("PFContainerType: %d", containerType)

				if containerType == 1 {
					log.Printf("oDU PF Container: ")

					oDU := pmContainer.PFContainer.Container.(*ODUPFContainerType)

					cellResourceReportCount := oDU.CellResourceReportCount
					log.Printf("CellResourceReportCount: %d", cellResourceReportCount)

					for cellResourceReportCounter := 0; cellResourceReportCounter < cellResourceReportCount; cellResourceReportCounter++ {
						log.Printf("CellResourceReport[%d]: ", cellResourceReportCounter)

						cellResourceReport := oDU.CellResourceReports[cellResourceReportCounter]

						log.Printf("nRCGI.PlmnID: %x", cellResourceReport.NRCGI.PlmnID.Buf)
						log.Printf("nRCGI.nRCellID: %x", cellResourceReport.NRCGI.NRCellID.Buf)

						cellID, err := e2sm.ParseNRCGI(cellResourceReport.NRCGI)
						if err != nil {
							xapp.Logger.Error("Failed to parse CellID in DU PF Container: %v", err)
							log.Printf("Failed to parse CellID in DU PF Container: %v", err)
							continue
						}
						if cellID == cellIDHdr {
							flag = true
						}

						log.Printf("TotalofAvailablePRBsDL: %d", cellResourceReport.TotalofAvailablePRBs.DL)
						log.Printf("TotalofAvailablePRBsUL: %d", cellResourceReport.TotalofAvailablePRBs.UL)

						if flag {
							availPRBDL = cellResourceReport.TotalofAvailablePRBs.DL
							availPRBUL = cellResourceReport.TotalofAvailablePRBs.UL
						}

						servedPlmnPerCellCount := cellResourceReport.ServedPlmnPerCellCount
						log.Printf("ServedPlmnPerCellCount: %d", servedPlmnPerCellCount)

						for servedPlmnPerCellCounter := 0; servedPlmnPerCellCounter < servedPlmnPerCellCount; servedPlmnPerCellCounter++ {
							log.Printf("ServedPlmnPerCell[%d]: ", servedPlmnPerCellCounter)

							servedPlmnPerCell := cellResourceReport.ServedPlmnPerCells[servedPlmnPerCellCounter]

							log.Printf("PlmnID: %x", servedPlmnPerCell.PlmnID.Buf)

							if servedPlmnPerCell.DUPM5GC != nil {
								slicePerPlmnPerCellCount := servedPlmnPerCell.DUPM5GC.SlicePerPlmnPerCellCount
								log.Printf("SlicePerPlmnPerCellCount: %d", slicePerPlmnPerCellCount)

								for slicePerPlmnPerCellCounter := 0; slicePerPlmnPerCellCounter < slicePerPlmnPerCellCount; slicePerPlmnPerCellCounter++ {
									log.Printf("SlicePerPlmnPerCell[%d]: ", slicePerPlmnPerCellCounter)

									slicePerPlmnPerCell := servedPlmnPerCell.DUPM5GC.SlicePerPlmnPerCells[slicePerPlmnPerCellCounter]

									log.Printf("SliceID.sST: %x", slicePerPlmnPerCell.SliceID.SST.Buf)
									if slicePerPlmnPerCell.SliceID.SD != nil {
										log.Printf("SliceID.sD: %x", slicePerPlmnPerCell.SliceID.SD.Buf)
									}

									fQIPERSlicesPerPlmnPerCellCount := slicePerPlmnPerCell.FQIPERSlicesPerPlmnPerCellCount
									log.Printf("5QIPerSlicesPerPlmnPerCellCount: %d", fQIPERSlicesPerPlmnPerCellCount)

									for fQIPERSlicesPerPlmnPerCellCounter := 0; fQIPERSlicesPerPlmnPerCellCounter < fQIPERSlicesPerPlmnPerCellCount; fQIPERSlicesPerPlmnPerCellCounter++ {
										log.Printf("5QIPerSlicesPerPlmnPerCell[%d]: ", fQIPERSlicesPerPlmnPerCellCounter)

										fQIPERSlicesPerPlmnPerCell := slicePerPlmnPerCell.FQIPERSlicesPerPlmnPerCells[fQIPERSlicesPerPlmnPerCellCounter]

										log.Printf("5QI: %d", fQIPERSlicesPerPlmnPerCell.FiveQI)
										log.Printf("PrbUsageDL: %d", fQIPERSlicesPerPlmnPerCell.PrbUsage.DL)
										log.Printf("PrbUsageUL: %d", fQIPERSlicesPerPlmnPerCell.PrbUsage.UL)
									}
								}
							}

							if servedPlmnPerCell.DUPMEPC != nil {
								perQCIReportCount := servedPlmnPerCell.DUPMEPC.PerQCIReportCount
								log.Printf("PerQCIReportCount: %d", perQCIReportCount)

								for perQCIReportCounter := 0; perQCIReportCounter < perQCIReportCount; perQCIReportCounter++ {
									log.Printf("PerQCIReports[%d]: ", perQCIReportCounter)

									perQCIReport := servedPlmnPerCell.DUPMEPC.PerQCIReports[perQCIReportCounter]

									log.Printf("QCI: %d", perQCIReport.QCI)
									log.Printf("PrbUsageDL: %d", perQCIReport.PrbUsage.DL)
									log.Printf("PrbUsageUL: %d", perQCIReport.PrbUsage.UL)
								}
							}
						}
					}
				} else if containerType == 2 {
					log.Printf("oCU-CP PF Container: ")

					oCUCP := pmContainer.PFContainer.Container.(*OCUCPPFContainerType)

					if oCUCP.GNBCUCPName != nil {
						log.Printf("gNB-CU-CP Name: %x", oCUCP.GNBCUCPName.Buf)
					}

					log.Printf("NumberOfActiveUEs: %d", oCUCP.CUCPResourceStatus.NumberOfActiveUEs)
				} else if containerType == 3 {
					log.Printf("oCU-UP PF Container: ")

					oCUUP := pmContainer.PFContainer.Container.(*OCUUPPFContainerType)

					if oCUUP.GNBCUUPName != nil {
						log.Printf("gNB-CU-UP Name: %x", oCUUP.GNBCUUPName.Buf)
					}

					cuUPPFContainerItemCount := oCUUP.CUUPPFContainerItemCount
					log.Printf("CU-UP PF Container Item Count: %d", cuUPPFContainerItemCount)

					for cuUPPFContainerItemCounter := 0; cuUPPFContainerItemCounter < cuUPPFContainerItemCount; cuUPPFContainerItemCounter++ {
						log.Printf("CU-UP PF Container Item [%d]: ", cuUPPFContainerItemCounter)

						cuUPPFContainerItem := oCUUP.CUUPPFContainerItems[cuUPPFContainerItemCounter]

						log.Printf("InterfaceType: %d", cuUPPFContainerItem.InterfaceType)

						cuUPPlmnCount := cuUPPFContainerItem.OCUUPPMContainer.CUUPPlmnCount
						log.Printf("CU-UP Plmn Count: %d", cuUPPlmnCount)

						for cuUPPlmnCounter := 0; cuUPPlmnCounter < cuUPPlmnCount; cuUPPlmnCounter++ {
							log.Printf("CU-UP Plmn [%d]: ", cuUPPlmnCounter)

							cuUPPlmn := cuUPPFContainerItem.OCUUPPMContainer.CUUPPlmns[cuUPPlmnCounter]

							log.Printf("PlmnID: %x", cuUPPlmn.PlmnID.Buf)

							plmnID, err := e2sm.ParsePLMNIdentity(cuUPPlmn.PlmnID.Buf, cuUPPlmn.PlmnID.Size)
							if err != nil {
								xapp.Logger.Error("Failed to parse PlmnID in CU-UP PF Container: %v", err)
								log.Printf("Failed to parse PlmnID in CU-UP PF Container: %v", err)
								continue
							}

							if cuUPPlmn.CUUPPM5GC != nil {
								sliceToReportCount := cuUPPlmn.CUUPPM5GC.SliceToReportCount
								log.Printf("SliceToReportCount: %d", sliceToReportCount)

								for sliceToReportCounter := 0; sliceToReportCounter < sliceToReportCount; sliceToReportCounter++ {
									log.Printf("SliceToReport[%d]: ", sliceToReportCounter)

									sliceToReport := cuUPPlmn.CUUPPM5GC.SliceToReports[sliceToReportCounter]

									log.Printf("SliceID.sST: %x", sliceToReport.SliceID.SST.Buf)
									if sliceToReport.SliceID.SD != nil {
										log.Printf("SliceID.sD: %x", sliceToReport.SliceID.SD.Buf)
									}

									sliceID, err := e2sm.ParseSliceID(sliceToReport.SliceID)
									if err != nil {
										xapp.Logger.Error("Failed to parse sliceID in CU-UP PF Container with PlmnID [%s]: %v", plmnID, err)
										log.Printf("Failed to parse sliceID in CU-UP PF Container with PlmnID [%s]: %v", plmnID, err)
										continue
									}

									fQIPERSlicesPerPlmnCount := sliceToReport.FQIPERSlicesPerPlmnCount
									log.Printf("5QIPerSlicesPerPlmnCount: %d", fQIPERSlicesPerPlmnCount)

									for fQIPERSlicesPerPlmnCounter := 0; fQIPERSlicesPerPlmnCounter < fQIPERSlicesPerPlmnCount; fQIPERSlicesPerPlmnCounter++ {
										log.Printf("5QIPerSlicesPerPlmn[%d]: ", fQIPERSlicesPerPlmnCounter)

										fQIPERSlicesPerPlmn := sliceToReport.FQIPERSlicesPerPlmns[fQIPERSlicesPerPlmnCounter]

										fiveQI := fQIPERSlicesPerPlmn.FiveQI
										log.Printf("5QI: %d", fiveQI)

										if plmnID == plmnIDHdr && sliceID == sliceIDHdr && fiveQI == fiveQIHdr {
											flag = true
										}

										if fQIPERSlicesPerPlmn.PDCPBytesDL != nil {
											log.Printf("PDCPBytesDL: %x", fQIPERSlicesPerPlmn.PDCPBytesDL.Buf)

											if flag {
												dlPDCPBytes, err = e2sm.ParseInteger(fQIPERSlicesPerPlmn.PDCPBytesDL.Buf, fQIPERSlicesPerPlmn.PDCPBytesDL.Size)
												if err != nil {
													xapp.Logger.Error("Failed to parse PDCPBytesDL in CU-UP PF Container with PlmnID [%s], SliceID [%d], 5QI [%d]: %v", plmnID, sliceID, fiveQI, err)
													log.Printf("Failed to parse PDCPBytesDL in CU-UP PF Container with PlmnID [%s], SliceID [%d], 5QI [%d]: %v", plmnID, sliceID, fiveQI, err)
													continue
												}
											}
										}

										if fQIPERSlicesPerPlmn.PDCPBytesUL != nil {
											log.Printf("PDCPBytesUL: %x", fQIPERSlicesPerPlmn.PDCPBytesUL.Buf)

											if flag {
												ulPDCPBytes, err = e2sm.ParseInteger(fQIPERSlicesPerPlmn.PDCPBytesUL.Buf, fQIPERSlicesPerPlmn.PDCPBytesUL.Size)
												if err != nil {
													xapp.Logger.Error("Failed to parse PDCPBytesUL in CU-UP PF Container with PlmnID [%s], SliceID [%d], 5QI [%d]: %v", plmnID, sliceID, fiveQI, err)
													log.Printf("Failed to parse PDCPBytesUL in CU-UP PF Container with PlmnID [%s], SliceID [%d], 5QI [%d]: %v", plmnID, sliceID, fiveQI, err)
													continue
												}
											}
										}
									}
								}
							}

							if cuUPPlmn.CUUPPMEPC != nil {
								cuUPPMEPCPerQCIReportCount := cuUPPlmn.CUUPPMEPC.CUUPPMEPCPerQCIReportCount
								log.Printf("PerQCIReportCount: %d", cuUPPMEPCPerQCIReportCount)

								for cuUPPMEPCPerQCIReportCounter := 0; cuUPPMEPCPerQCIReportCounter < cuUPPMEPCPerQCIReportCount; cuUPPMEPCPerQCIReportCounter++ {
									log.Printf("PerQCIReport[%d]: ", cuUPPMEPCPerQCIReportCounter)

									cuUPPMEPCPerQCIReport := cuUPPlmn.CUUPPMEPC.CUUPPMEPCPerQCIReports[cuUPPMEPCPerQCIReportCounter]

									log.Printf("QCI: %d", cuUPPMEPCPerQCIReport.QCI)

									if cuUPPMEPCPerQCIReport.PDCPBytesDL != nil {
										log.Printf("PDCPBytesDL: %x", cuUPPMEPCPerQCIReport.PDCPBytesDL.Buf)
									}
									if cuUPPMEPCPerQCIReport.PDCPBytesUL != nil {
										log.Printf("PDCPBytesUL: %x", cuUPPMEPCPerQCIReport.PDCPBytesUL.Buf)
									}
								}
							}
						}
					}
				} else {
					xapp.Logger.Error("Unknown PF Container type: %d", containerType)
					log.Printf("Unknown PF Container type: %d", containerType)
					continue
				}
			}

			if pmContainer.RANContainer != nil {
				log.Printf("RANContainer: %x", pmContainer.RANContainer.Timestamp.Buf)

				timestamp, _ := e2sm.ParseTimestamp(pmContainer.RANContainer.Timestamp.Buf, pmContainer.RANContainer.Timestamp.Size)
				log.Printf("Timestamp=[sec: %d, nsec: %d]", timestamp.TVsec, timestamp.TVnsec)

				containerType = pmContainer.RANContainer.ContainerType
				if containerType == 1 {
					log.Printf("DU Usage Report: ")

					oDUUE := pmContainer.RANContainer.Container.(*DUUsageReportType)

					for CellResourceReportItemCounter := 0; CellResourceReportItemCounter < oDUUE.CellResourceReportItemCount; CellResourceReportItemCounter++ {
						cellResourceReportItem := oDUUE.CellResourceReportItems[CellResourceReportItemCounter]

						log.Printf("nRCGI.PlmnID: %x", cellResourceReportItem.NRCGI.PlmnID.Buf)
						log.Printf("nRCGI.NRCellID: %x, Unused: %d", cellResourceReportItem.NRCGI.NRCellID.Buf, cellResourceReportItem.NRCGI.NRCellID.BitsUnused)

						servingCellID, err := e2sm.ParseNRCGI(cellResourceReportItem.NRCGI)
						if err != nil {
							xapp.Logger.Error("Failed to parse NRCGI in DU Usage Report: %v", err)
							log.Printf("Failed to parse NRCGI in DU Usage Report: %v", err)
							continue
						}

						for UeResourceReportItemCounter := 0; UeResourceReportItemCounter < cellResourceReportItem.UeResourceReportItemCount; UeResourceReportItemCounter++ {
							ueResourceReportItem := cellResourceReportItem.UeResourceReportItems[UeResourceReportItemCounter]

							log.Printf("C-RNTI: %x", ueResourceReportItem.CRNTI.Buf)

							ueID, err := e2sm.ParseInteger(ueResourceReportItem.CRNTI.Buf, ueResourceReportItem.CRNTI.Size)
							if err != nil {
								xapp.Logger.Error("Failed to parse C-RNTI in DU Usage Report with Serving Cell ID [%s]: %v", servingCellID, err)
								log.Printf("Failed to parse C-RNTI in DU Usage Report with Serving Cell ID [%s]: %v", servingCellID, err)
								continue
							}

							var ueMetrics UeMetricsEntry

							ueMetrics.UeID = ueID
							log.Printf("UeID: %d", ueMetrics.UeID)
							ueMetrics.ServingCellID = servingCellID
							log.Printf("ServingCellID: %s", ueMetrics.ServingCellID)
							ueMetrics.MeasPeriodRF = 20

							if flag {
								timestampPRB = timestamp
							}

							ueMetrics.MeasTimestampPRB.TVsec = timestamp.TVsec
							ueMetrics.MeasTimestampPRB.TVnsec = timestamp.TVnsec

							if ueResourceReportItem.PRBUsageDL != -1 {
								ueMetrics.PRBUsageDL = ueResourceReportItem.PRBUsageDL
								log.Printf("PRBUsageDL: %d", ueMetrics.PRBUsageDL)
							}

							if ueResourceReportItem.PRBUsageUL != -1 {
								ueMetrics.PRBUsageUL = ueResourceReportItem.PRBUsageUL
								log.Printf("PRBUsageUL: %d", ueMetrics.PRBUsageUL)
							}
							c.writeUeMetrics_db(ueMetrics)
						}
					}
				} else if containerType == 2 {
					log.Printf("CU-CP Usage Report: ")

					oCUCPUE := pmContainer.RANContainer.Container.(*CUCPUsageReportType)

					for CellResourceReportItemCounter := 0; CellResourceReportItemCounter < oCUCPUE.CellResourceReportItemCount; CellResourceReportItemCounter++ {
						cellResourceReportItem := oCUCPUE.CellResourceReportItems[CellResourceReportItemCounter]

						log.Printf("nRCGI.PlmnID: %x", cellResourceReportItem.NRCGI.PlmnID.Buf)
						log.Printf("nRCGI.NRCellID: %x, Unused: %d", cellResourceReportItem.NRCGI.NRCellID.Buf, cellResourceReportItem.NRCGI.NRCellID.BitsUnused)

						servingCellID, err := e2sm.ParseNRCGI(cellResourceReportItem.NRCGI)
						if err != nil {
							xapp.Logger.Error("Failed to parse NRCGI in CU-CP Usage Report: %v", err)
							log.Printf("Failed to parse NRCGI in CU-CP Usage Report: %v", err)
							continue
						}

						for UeResourceReportItemCounter := 0; UeResourceReportItemCounter < cellResourceReportItem.UeResourceReportItemCount; UeResourceReportItemCounter++ {
							ueResourceReportItem := cellResourceReportItem.UeResourceReportItems[UeResourceReportItemCounter]

							log.Printf("C-RNTI: %x", ueResourceReportItem.CRNTI.Buf)

							ueID, err := e2sm.ParseInteger(ueResourceReportItem.CRNTI.Buf, ueResourceReportItem.CRNTI.Size)
							if err != nil {
								xapp.Logger.Error("Failed to parse C-RNTI in CU-CP Usage Report with Serving Cell ID [%s]: %v", err)
								log.Printf("Failed to parse C-RNTI in CU-CP Usage Report with Serving Cell ID [%s]: %v", err)
								continue
							}

							var ueMetrics UeMetricsEntry

							ueMetrics.UeID = ueID
							log.Printf("UeID: %d", ueMetrics.UeID)
							ueMetrics.ServingCellID = servingCellID
							log.Printf("ServingCellID: %s", ueMetrics.ServingCellID)

							ueMetrics.MeasTimeRF.TVsec = timestamp.TVsec
							ueMetrics.MeasTimeRF.TVnsec = timestamp.TVnsec

							ueMetrics.MeasPeriodPDCP = 20
							ueMetrics.MeasPeriodPRB = 20

							if ueResourceReportItem.ServingCellRF != nil {
								err = json.Unmarshal(ueResourceReportItem.ServingCellRF.Buf, &ueMetrics.ServingCellRF)
								log.Printf("ueMetrics.ServingCellRF: %+v", ueMetrics.ServingCellRF)
								if err != nil {
									xapp.Logger.Error("Failed to Unmarshal ServingCellRF in CU-CP Usage Report with UE ID [%d]: %v", ueID, err)
									log.Printf("Failed to Unmarshal ServingCellRF in CU-CP Usage Report with UE ID [%d]: %v", ueID, err)
									log.Printf("ServingCellRF raw data: %x", ueResourceReportItem.ServingCellRF.Buf)
									continue
								}
							}

							if ueResourceReportItem.NeighborCellRF != nil {
								err = json.Unmarshal(ueResourceReportItem.NeighborCellRF.Buf, &ueMetrics.NeighborCellsRF)
								log.Printf("ueMetrics.NeighborCellsRF: %+v", ueMetrics.NeighborCellsRF)
								if err != nil {
									xapp.Logger.Error("Failed to Unmarshal NeighborCellRF in CU-CP Usage Report with UE ID [%d]: %v", ueID, err)
									log.Printf("Failed to Unmarshal NeighborCellRF in CU-CP Usage Report with UE ID [%d]: %v", ueID, err)
									log.Printf("NeighborCellRF raw data: %x", ueResourceReportItem.NeighborCellRF.Buf)
									continue
								}
							}
							c.writeUeMetrics_db(ueMetrics)
						}
					}
				} else if containerType == 3 {
					log.Printf("CU-UP Usage Report: ")

					oCUUPUE := pmContainer.RANContainer.Container.(*CUUPUsageReportType)

					for CellResourceReportItemCounter := 0; CellResourceReportItemCounter < oCUUPUE.CellResourceReportItemCount; CellResourceReportItemCounter++ {
						cellResourceReportItem := oCUUPUE.CellResourceReportItems[CellResourceReportItemCounter]

						log.Printf("nRCGI.PlmnID: %x", cellResourceReportItem.NRCGI.PlmnID.Buf)
						log.Printf("nRCGI.NRCellID: %x, Unused: %d", cellResourceReportItem.NRCGI.NRCellID.Buf, cellResourceReportItem.NRCGI.NRCellID.BitsUnused)

						servingCellID, err := e2sm.ParseNRCGI(cellResourceReportItem.NRCGI)
						if err != nil {
							xapp.Logger.Error("Failed to parse NRCGI in CU-UP Usage Report: %v", err)
							log.Printf("Failed to parse NRCGI in CU-UP Usage Report: %v", err)
							continue
						}

						for UeResourceReportItemCounter := 0; UeResourceReportItemCounter < cellResourceReportItem.UeResourceReportItemCount; UeResourceReportItemCounter++ {
							ueResourceReportItem := cellResourceReportItem.UeResourceReportItems[UeResourceReportItemCounter]

							log.Printf("C-RNTI: %x", ueResourceReportItem.CRNTI.Buf)

							ueID, err := e2sm.ParseInteger(ueResourceReportItem.CRNTI.Buf, ueResourceReportItem.CRNTI.Size)
							if err != nil {
								xapp.Logger.Error("Failed to parse C-RNTI in CU-UP Usage Report Serving Cell ID [%s]: %v", servingCellID, err)
								log.Printf("Failed to parse C-RNTI in CU-UP Usage Report Serving Cell ID [%s]: %v", servingCellID, err)
								continue
							}

							var ueMetrics UeMetricsEntry

							ueMetrics.UeID = ueID
							log.Printf("UeID: %d", ueMetrics.UeID)
							ueMetrics.ServingCellID = servingCellID
							log.Printf("ServingCellID: %s", ueMetrics.ServingCellID)

							if flag {
								timestampPDCPBytes = timestamp
							}

							ueMetrics.MeasTimestampPDCPBytes.TVsec = timestamp.TVsec
							ueMetrics.MeasTimestampPDCPBytes.TVnsec = timestamp.TVnsec

							if ueResourceReportItem.PDCPBytesDL != nil {
								ueMetrics.PDCPBytesDL, err = e2sm.ParseInteger(ueResourceReportItem.PDCPBytesDL.Buf, ueResourceReportItem.PDCPBytesDL.Size)
								if err != nil {
									xapp.Logger.Error("Failed to parse PDCPBytesDL in CU-UP Usage Report with UE ID [%d]: %v", ueID, err)
									log.Printf("Failed to parse PDCPBytesDL in CU-UP Usage Report with UE ID [%d]: %v", ueID, err)
									continue
								}
							}

							if ueResourceReportItem.PDCPBytesUL != nil {
								ueMetrics.PDCPBytesUL, err = e2sm.ParseInteger(ueResourceReportItem.PDCPBytesUL.Buf, ueResourceReportItem.PDCPBytesUL.Size)
								if err != nil {
									xapp.Logger.Error("Failed to parse PDCPBytesUL in CU-UP Usage Report with UE ID [%d]: %v", ueID, err)
									log.Printf("Failed to parse PDCPBytesUL in CU-UP Usage Report with UE ID [%d]: %v", ueID, err)
									continue
								}
							}

							c.writeUeMetrics_db(ueMetrics)
						}
					}
				} else {
					xapp.Logger.Error("Unknown PF Container Type: %d", containerType)
					log.Printf("Unknown PF Container Type: %d", containerType)
					continue
				}
			}

			if flag {
				var cellMetrics CellMetricsEntry

				cellMetrics.MeasPeriodPDCP = 20
				cellMetrics.MeasPeriodPRB = 20
				cellMetrics.CellID = cellIDHdr

				if timestampPDCPBytes != nil {
					cellMetrics.MeasTimestampPDCPBytes.TVsec = timestampPDCPBytes.TVsec
					cellMetrics.MeasTimestampPDCPBytes.TVnsec = timestampPDCPBytes.TVnsec
				}
				if dlPDCPBytes != -1 {
					cellMetrics.PDCPBytesDL = dlPDCPBytes
				}
				if ulPDCPBytes != -1 {
					cellMetrics.PDCPBytesUL = ulPDCPBytes
				}
				if timestampPRB != nil {
					cellMetrics.MeasTimestampPRB.TVsec = timestampPRB.TVsec
					cellMetrics.MeasTimestampPRB.TVnsec = timestampPRB.TVnsec
				}
				if availPRBDL != -1 {
					cellMetrics.AvailPRBDL = availPRBDL
				}
				if availPRBUL != -1 {
					cellMetrics.AvailPRBUL = availPRBUL
				}
				c.writeCellMetrics_db(cellMetrics)
			}
		}
	} else {
		xapp.Logger.Error("Unknown RIC Indication Message Format: %d", indMsg.IndMsgType)
		log.Printf("Unkonw RIC Indication Message Format: %d", indMsg.IndMsgType)
		return
	}

	return nil
}

func (c *Control) writeUeMetrics_db(ueMetrics UeMetricsEntry) {
	//Write UE metrics to InfluxDB using API
	writeAPI := c.client.WriteAPIBlocking("my-org", "kpimon")
	ueMetricsJSON, err := json.Marshal(ueMetrics)
	if err != nil {
		xapp.Logger.Error("Marshal UE Metrics failed!")
	}
	p := influxdb2.NewPointWithMeasurement("ricIndication_UeMetrics").
		AddField("UE Metrics", ueMetricsJSON).
		SetTime(time.Now())
	writeAPI.WritePoint(context.Background(), p)
	xapp.Logger.Info("Wrote UE Metrics to InfluxDB")
}
func (c *Control) writeCellMetrics_db(cellMetrics CellMetricsEntry) {
	//Write cell metrics to InfluxDB using API
	writeAPI := c.client.WriteAPIBlocking("my-org", "kpimon")
	cellMetricsJSON, er := json.Marshal(cellMetrics)
	if er != nil {
		xapp.Logger.Error("Marshal Cell Metrics failed!")
	}
	p := influxdb2.NewPointWithMeasurement("ricIndication_cellMetrics").
		AddField("Cell Metrics", cellMetricsJSON).
		SetTime(time.Now())
	writeAPI.WritePoint(context.Background(), p)
	xapp.Logger.Info("Wrote Cell Metrics to InfluxDB")
}
func (c Control) xAppStartCB(d interface{}) {
	xapp.Logger.Info("In callback KPI monitor xApp ...")

	// Get eNodeB list
	nbList := c.getnbList()

	// Send subscription request to connected NodeB
	for _, nb := range nbList {
		if nb.ConnectionStatus == 1 {
			xapp.Logger.Info("Before send subscription request to %v", nb.InventoryName)
			c.sendSubscription(nb.InventoryName)
			xapp.Logger.Info("After send subscription request to %v", nb.InventoryName)
		}

	}
	//Call controlLoop to handle the message from e2Nodes
	go c.controlLoop()
	xapp.Logger.Info("End callback KPI monitor xApp ...")
}

func (c Control) Run() {
	// Setup level
	xapp.Logger.SetLevel(xapp.Config.GetInt("logger.level"))
	// Register callback
	xapp.SetReadyCB(c.xAppStartCB, true)
	// Start xApp
	xapp.Run(c)
}
