package control

/*
#include <e2sm/wrapper.h>
#cgo LDFLAGS: -lm  -le2smwrapper
#cgo CFLAGS:  -I/usr/local/include/e2sm
*/
import "C"
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/clientmodel"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"

	//	"bytes"
	//	"encoding/binary"

	//"encoding/base64"
	//"strings"

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
	funcId               = int64(2)
	hPort                = int64(8080)
	rPort                = int64(4560)
	clientEndpoint       = clientmodel.SubscriptionParamsClientEndpoint{Host: "service-ricxapp-kpimon-go-http.ricxapp", HTTPPort: &hPort, RMRPort: &rPort}
)
var Glob_cell = make(map[string]bool)
var Glob_Ran_cell = make(map[string][]string)
var Glob_cell_Plmn = make(map[string]string)
var ranUeKpi = make(map[string][]string)
var ranCellKpi = make(map[string][]string)

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
	xapp.Logger.Info("In new control")
	create_db()
	xapp.Logger.Info("returning control")
	return Control{
		make(chan *xapp.RMRParams),
		influxdb2.NewClient("http://ricplt-influxdb.ricplt:8086", "client"),
	}
}
func create_db() {
	//Create a database named kpimon in influxDB
	xapp.Logger.Info("In create_db")
	_, err := http.Post("http://ricplt-influxdb.ricplt:8086/query?q=create%20database%20kpimon", "", nil)
	if err != nil {
		xapp.Logger.Error("Create database failed!")
	}
	xapp.Logger.Info("exiting create_db")
}

func (c Control) getEnbList() ([]*xapp.RNIBNbIdentity, error) {
	enbs, err := xapp.Rnib.GetListEnbIds()
	if err != nil {
		xapp.Logger.Error("err: %s", err)
		return nil, err
	}

	var connected_enb_names []string
	for _, enb := range enbs {
		connected_enb_names = append(connected_enb_names, enb.InventoryName)
	}
	xapp.Logger.Info("List for connected eNBs:", connected_enb_names)
	
	return enbs, nil
}

func (c *Control) getGnbList() ([]*xapp.RNIBNbIdentity, error) {
	gnbs, err := xapp.Rnib.GetListGnbIds()

	if err != nil {
		xapp.Logger.Error("err: %s", err)
		return nil, err
	}

	var connected_gnb_names []string
	for _, gnb := range gnbs {
		connected_gnb_names = append(connected_gnb_names, gnb.InventoryName)
	}
	xapp.Logger.Info("List for connected gNBs:", connected_gnb_names)

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

func cellid_to_list_of_int(str string) []int64 {
	l := len(str)
	var ans []int64
	for i := 0; i < l; i += 2 {
		output, err := strconv.ParseInt(str[i:i+2], 16, 64)
		if err != nil {
			fmt.Println(err)
			return ans
		}
		ans = append(ans, output)
	}
	return ans
}
func plmnid_to_list_of_int(str string) []int64 {
	l := len(str)
	var ans []int64
	for i := 0; i < l; i += 2 {
		output, err := strconv.ParseInt(str[i:i+2], 16, 64)
		if err != nil {
			fmt.Println(err)
			return ans
		}
		ans = append(ans, output)
	}
	return ans
}
func encode_action_format1(plmn string, cellid string, meid string) clientmodel.ActionDefinition {
	lol1 := plmnid_to_list_of_int(plmn)
	lol2 := cellid_to_list_of_int(cellid)
	var format1 []int64
	//format1=[]int64{0,1,1,8,0,19,32,0,3,1,32,0,0,32,0,4,1,32,0,0,32,0,78,1,32,0,0,32,0,79,1,32,0,0,32,0,8,1,32,0,0,32,0,7,1,32,0,0,32,0,11,1,32,0,0,32,0,12,1,32,0,0,32,0,80,1,32,0,0,32,0,81,1,32,0,0,32,0,13,1,32,0,0,32,0,14,1,32,0,0,32,0,40,1,32,0,0,32,0,41,1,32,0,0,32,0,42,1,32,0,0,32,0,82,1,32,0,0,32,0,83,1,32,0,0,32,0,84,1,32,0,0,32,0,85,1,32,0,0,32,0,86,1,32,0,0,64,39,15,0,0,31,1,18,52,92,0,16}
	//0(nr) 1(eutra)  0,31,1(plmn)18,52,92,0,16(cellid)
	link := "http://service-ricplt-e2mgr-http.ricplt.svc.cluster.local:3800/v1/nodeb/"

	link = link + meid
	tmpr, err := http.Get(link)
	if err != nil {
		log.Fatalln(err)
		l := []int64{0}
		return l
	}
	defer tmpr.Body.Close()
	var resp E2mgrResponse

	err = json.NewDecoder(tmpr.Body).Decode(&resp)
	if err != nil {
		log.Fatalln(err)
		l := []int64{0}
		return l

	}
	counter := 0
	//RanFunctionId=2 for kpm in viavi
	for i := 0; i < len(resp.Gnb.RanFunctions); i++ {
		if resp.Gnb.RanFunctions[i].RanFunctionId == 2 {
			counter = i
			break
		}
	}
	cString := C.CString(resp.Gnb.RanFunctions[counter].RanFunctionDefinition)
	defer C.free(unsafe.Pointer(cString)) // Free the allocated C string when done
	// Call the C function
	determine := 2 //2 for format1 by name
	result := C.encode_action_Definition(cString, C.int(determine))

	for i := 0; i < int(result.length); i++ {
		value := int64(*(*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(result.array)) + uintptr(i*4))))
		format1 = append(format1, value)
	}
	format1 = append(format1, lol1...) //appending plmn
	format1 = append(format1, lol2...) //appending cellid
	return format1
}

func encode_action_format2() clientmodel.ActionDefinition {
	var format2 []int64
	format2 = []int64{0, 1, 0, 0, 0, 20, 0, 160, 68, 82, 66, 46, 85, 69, 84, 104, 112, 68, 108, 1, 0, 0, 0, 1, 64, 68, 82, 66, 46, 85, 69, 84, 104, 112, 85, 108, 1, 0, 0, 0, 1, 0, 71, 78, 66, 45, 68, 85, 45, 73, 68, 1, 0, 0, 0, 0, 160, 78, 82, 45, 67, 71, 73, 1, 0, 0, 0, 0, 160, 78, 82, 45, 80, 67, 73, 1, 0, 0, 0, 2, 192, 81, 111, 115, 70, 108, 111, 119, 46, 80, 100, 99, 112, 80, 100, 117, 86, 111, 108, 117, 109, 101, 68, 108, 1, 0, 0, 0, 2, 192, 81, 111, 115, 70, 108, 111, 119, 46, 80, 100, 99, 112, 80, 100, 117, 86, 111, 108, 117, 109, 101, 85, 108, 1, 0, 0, 0, 1, 64, 82, 82, 67, 46, 67, 111, 110, 110, 77, 97, 120, 1, 0, 0, 0, 1, 96, 82, 82, 67, 46, 67, 111, 110, 110, 77, 101, 97, 110, 1, 0, 0, 0, 1, 160, 82, 82, 85, 46, 80, 114, 98, 65, 118, 97, 105, 108, 68, 108, 1, 0, 0, 0, 1, 160, 82, 82, 85, 46, 80, 114, 98, 65, 118, 97, 105, 108, 85, 108, 1, 0, 0, 0, 1, 32, 82, 82, 85, 46, 80, 114, 98, 84, 111, 116, 1, 0, 0, 0, 1, 96, 82, 82, 85, 46, 80, 114, 98, 84, 111, 116, 68, 108, 1, 0, 0, 0, 1, 96, 82, 82, 85, 46, 80, 114, 98, 84, 111, 116, 85, 108, 1, 0, 0, 0, 1, 128, 82, 82, 85, 46, 80, 114, 98, 85, 115, 101, 100, 68, 108, 1, 0, 0, 0, 1, 128, 82, 82, 85, 46, 80, 114, 98, 85, 115, 101, 100, 85, 108, 1, 0, 0, 0, 1, 64, 86, 105, 97, 118, 105, 46, 71, 101, 111, 46, 120, 1, 0, 0, 0, 1, 64, 86, 105, 97, 118, 105, 46, 71, 101, 111, 46, 121, 1, 0, 0, 0, 1, 64, 86, 105, 97, 118, 105, 46, 71, 101, 111, 46, 122, 1, 0, 0, 0, 2, 0, 86, 105, 97, 118, 105, 46, 82, 97, 100, 105, 111, 46, 112, 111, 119, 101, 114, 1, 0, 0, 0, 2, 64, 86, 105, 97, 118, 105, 46, 82, 97, 100, 105, 111, 46, 115, 101, 99, 116, 111, 114, 115, 1, 0, 0, 0, 0, 0}
	//encode the variable part and append it to our array.
	format2 = append(format2, 89) //appending variable part if necessory
	return format2
}

func encode_action_format3(meid string) clientmodel.ActionDefinition {
	var format3 []int64

	//for simulation-by measId(not supported in Viavi 1.4)
	//format3=[]int64{0,1,3,64,0,29,32,0,90,0,0,16,0,0,16,0,91,0,0,16,0,0,16,0,3,0,0,16,0,0,16,0,4,0,0,16,0,0,16,0,80,0,0,16,0,0,16,0,13,0,0,16,0,0,16,0,14,0,0,16,0,0,16,0,92,0,0,16,0,0,16,0,93,0,0,16,0,0,16,0,44,0,0,16,0,0,16,0,40,0,0,16,0,0,16,0,41,0,0,16,0,0,16,0,42,0,0,16,0,0,16,0,94,0,0,16,0,0,16,0,95,0,0,16,0,0,16,0,96,0,0,16,0,0,16,0,97,0,0,16,0,0,16,0,45,0,0,16,0,0,16,0,98,0,0,16,0,0,16,0,99,0,0,16,0,0,16,0,100,0,0,16,0,0,16,0,101,0,0,16,0,0,16,0,102,0,0,16,0,0,16,0,103,0,0,16,0,0,16,0,104,0,0,16,0,0,16,0,43,0,0,16,0,0,16,0,105,0,0,16,0,0,16,0,106,0,0,16,0,0,16,0,107,0,0,16,0,0,16,0,108,0,0,16,0,0,32,39,15}
	//variable part is not presetnt in action def format 3

	link := "http://service-ricplt-e2mgr-http.ricplt.svc.cluster.local:3800/v1/nodeb/"
	link = link + meid
	tmpr, err := http.Get(link)
	if err != nil {
		log.Fatalln(err)
		l := []int64{0}
		return l

	}
	defer tmpr.Body.Close()
	var resp E2mgrResponse

	err = json.NewDecoder(tmpr.Body).Decode(&resp)
	if err != nil {
		log.Fatalln(err)
		l := []int64{0}
		return l

	}
	counter := 0
	//RanFunctionId=2 for kpm in viavi
	for i := 0; i < len(resp.Gnb.RanFunctions); i++ {
		if resp.Gnb.RanFunctions[i].RanFunctionId == 2 {
			counter = i
			break
		}
	}
	cString := C.CString(resp.Gnb.RanFunctions[counter].RanFunctionDefinition)
	defer C.free(unsafe.Pointer(cString)) // Free the allocated C string when done
	// Call the C function
	determine := 4 //2 for format1 by name

	result := C.encode_action_Definition(cString, C.int(determine))
	for i := 0; i < int(result.length); i++ {
		value := int64(*(*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(result.array)) + uintptr(i*4))))
		format3 = append(format3, value)

	}

	//variable part is not presetnt in action def format 3
	return format3
}
func encode_actionsToBeSetup(meid string) clientmodel.ActionsToBeSetup {
	var l clientmodel.ActionsToBeSetup
	/*
			link:="http://service-ricplt-e2mgr-http.ricplt.svc.cluster.local:3800/v1/nodeb/"
			link=link+meid
			tmpr,err := http.Get(link)
			if err != nil {
		      		log.Fatalln(err)
		      		return l
			}
			defer tmpr.Body.Close()
			var resp E2mgrResponse

			err=json.NewDecoder(tmpr.Body).Decode(&resp)
			if err != nil {
		                log.Fatalln(err)
		                return l
		        }


			counter := 0
			for i := 0; i < len(resp.Gnb.NodeConfigs); i++ {
				if resp.Gnb.NodeConfigs[i].E2nodeComponentInterfaceType == "f1" {
					counter = i
					break
				}
			}
			tm := resp.Gnb.NodeConfigs[counter].E2nodeComponentRequestPart
			base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(tm)))
			nl, _ := base64.StdEncoding.Decode(base64Text, []byte(tm))
			message := string(base64Text[:nl])

			counter = 0
			for i := 0; i < len(meid); i++ {
				if meid[i] == '_' {
					counter++
				}
				if counter == 3 {
					counter = i + 1
					break
				}
			}

			ans := strings.ToUpper(meid[counter:len(meid)])
			l1 := int64(len(message))
			l2 := int64(len(ans))
			var cells []string
			for i := int64(0); i <= l1-l2; i++ {
				if strings.Contains(message[i:i+l2], ans) {
					Glob_cell[message[i:i+10]]=true
					cells = append(cells, message[i:i+10])
					fmt.Println(message[i : i+10])
				}
			}
	*/

	var n int64 = 1
	cells := Glob_Ran_cell[meid]
	//fmt.Println("len of Tonga cells= ",len(cells))
	//fmt.Println("Tonga cells = ", cells)
	//fmt.Println("Tonga  = ", Glob_Ran_cell)

	var tempCells []string
	for _, ele := range cells {
		if ele != "" {
			tempCells = append(tempCells, ele)
		}
	}
	//fmt.Println("len of Tonga tmep cells= ",len(tempCells))
	//fmt.Println("Tonga temp cells = ", tempCells)

	//var ue int64 = 1 //get no of ue connected to du(if required)

	//for action def 1
	for n <= int64(len(tempCells)) {
		//fix double length issue/empty cell sting  in Glob_Ran_cell map
		var tmp int64 = n
		var lol *int64 = &tmp
		s := clientmodel.ActionToBeSetup{
			ActionID:         lol,
			ActionType:       &actionType,
			ActionDefinition: encode_action_format1(Glob_cell_Plmn[tempCells[n-1]], tempCells[n-1], meid),
			SubsequentAction: &clientmodel.SubsequentAction{
				SubsequentActionType: &subsequentActionType,
				TimeToWait:           &timeToWait,
			},
		}
		l = append(l, &s)

		n = n + 1
	}
	var tmp_act_id int64 = n
	/*
		n = 1
		// for action def 2
		for n <= ue {
			var tmp int64 = tmp_act_id
			var lol *int64 = &tmp
			s := clientmodel.ActionToBeSetup{
				ActionID:         lol,
				ActionType:       &actionType,
				ActionDefinition: encode_action_format2(),
				SubsequentAction:  &clientmodel.SubsequentAction{
					SubsequentActionType: &subsequentActionType,
					TimeToWait:           &timeToWait,
				},
			}
			l = append(l, &s)
			tmp_act_id = tmp_act_id + 1
			n = n + 1
		}
	*/

	//for action def 3
	var tmp int64 = tmp_act_id
	var lol *int64 = &tmp
	s := clientmodel.ActionToBeSetup{
		ActionID:         lol,
		ActionType:       &actionType,
		ActionDefinition: encode_action_format3(meid),
		SubsequentAction: &clientmodel.SubsequentAction{
			SubsequentActionType: &subsequentActionType,
			TimeToWait:           &timeToWait,
		},
	}
	l = append(l, &s)

	return l

}
func (c Control) sendSubscription(meid string) {
	//Create Subscription message and send it to RIC platform
	xapp.Logger.Info("Sending subscription request for MEID: %v", meid)
	/*
		subscritionParams := clientmodel.SubscriptionParams{
			ClientEndpoint: &clientEndpoint,
			Meid:           &meid,
			RANFunctionID:  &funcId,
			SubscriptionDetails: clientmodel.SubscriptionDetailsList{
				&clientmodel.SubscriptionDetail{
					EventTriggers: clientmodel.EventTriggerDefinition{
						8,39,15,
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
	*/

	//8,39,15, for 10000 ms reporting period
	subscritionParams := clientmodel.SubscriptionParams{
		ClientEndpoint: &clientEndpoint,
		Meid:           &meid,
		RANFunctionID:  &funcId,
		SubscriptionDetails: clientmodel.SubscriptionDetailsList{
			&clientmodel.SubscriptionDetail{
				EventTriggers: clientmodel.EventTriggerDefinition{
					8, 39, 15,
				},
				XappEventInstanceID: &seqId,
				ActionToBeSetupList: encode_actionsToBeSetup(meid),
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

func Test() (err error) {
	var e2ap *E2ap
	//var e2sm *E2sm

	//indicationMsg, err := e2ap.GetIndicationMessage(params.Payload)
	indicationMsg, err := e2ap.GetIndicationM()

	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication message: %v", err)
		return
	}

	//log.Printf("RIC Indication message from {%s} received", params.Meid.RanName)
	/*
	   indicationHdr, err := e2sm.GetIndicationHeader(indicationMsg.IndHeader)
	   if err != nil {
	           xapp.Logger.Error("Failed to decode RIC Indication Header: %v", err)
	           return
	   }
	*/

	//Decoding message and put information into log
	log.Printf("-----------RIC Indication Header-----------")
	log.Printf("indicationMsg.IndHeader= %x", indicationMsg.IndHeader)
	/*
		buf := new(bytes.Buffer) //create my buffer
		binary.Write(buf, binary.LittleEndian, indicationMsg.IndHeader)
		log.Printf("binary Write buf= %x",buf )
		b := buf.Bytes()
		//str := buf.String()
		//log.Printf(" buf Strin()= %s",str )
		//cptr1:= unsafe.Pointer(C.CString(str))
		cptr1:= unsafe.Pointer(&b[0])
		defer C.free(cptr1)
	*/
	cptr1 := unsafe.Pointer(&indicationMsg.IndHeader[0])
	decodedHdr := C.e2sm_decode_ric_indication_header(cptr1, C.size_t(len(indicationMsg.IndHeader)))
	//decodedHdr := C.e2sm_decode_ric_indication_header(cptr1, C.size_t(len(str)))
	//decodedHdr := C.e2sm_decode_ric_indication_header(cptr1, C.size_t(buf.Len()))
	if decodedHdr == nil {
		return errors.New("e2sm wrapper is unable to get IndicationHeader due to wrong or invalid input")
	}
	defer C.e2sm_free_ric_indication_header(decodedHdr)
	IndHdrType := int32(decodedHdr.indicationHeader_formats.present)
	if IndHdrType == 0 {
		log.Printf("No Indication Header present")
	}
	if IndHdrType == 1 {
		log.Printf("Indication Header format = %d", IndHdrType)
		indHdrFormat1_C := *(**C.E2SM_KPM_IndicationHeader_Format1_t)(unsafe.Pointer(&decodedHdr.indicationHeader_formats.choice[0]))
		//senderName_C := (*C.PrintableString_t)(unsafe.Pointer(indHdrFormat1_C.senderName))
		senderName_C := indHdrFormat1_C.senderName
		var senderName []byte
		senderName = C.GoBytes(unsafe.Pointer(senderName_C.buf), C.int(senderName_C.size))
		log.Printf("Sender Name = %x", senderName)

		//senderType_C := (*C.PrintableString_t)(unsafe.Pointer(indHdrFormat1_C.senderType))
		senderType_C := indHdrFormat1_C.senderType
		//senderType []byte
		senderType := C.GoBytes(unsafe.Pointer(senderType_C.buf), C.int(senderType_C.size))
		log.Printf("Sender Type = %x", senderType)

		//vendorName_C := (*C.PrintableString_t)(unsafe.Pointer(indHdrFormat1_C.vendorName))
		vendorName_C := indHdrFormat1_C.vendorName
		//vendorName  []byte
		vendorName := C.GoBytes(unsafe.Pointer(vendorName_C.buf), C.int(vendorName_C.size))
		log.Printf("Vendor Name = %x", vendorName)

	}

	/*
	   indMsg, err := e2sm.GetIndicationMessage(indicationMsg.IndMessage)
	   if err != nil {
	           xapp.Logger.Error("Failed to decode RIC Indication Message: %v", err)
	           return
	   }
	*/
	log.Printf("-----------RIC Indication Message-----------")
	log.Printf("indicationMsg.IndMessage= %x", indicationMsg.IndMessage)
	cptr2 := unsafe.Pointer(&indicationMsg.IndMessage[0])
	indicationmessage := C.e2sm_decode_ric_indication_message(cptr2, C.size_t(len(indicationMsg.IndMessage)))
	if indicationmessage == nil {
		return errors.New("e2sm wrapper is unable to get IndicationMessage due to wrong or invalid input")
	}
	defer C.e2sm_free_ric_indication_message(indicationmessage)
	IndMsgType := int32(indicationmessage.indicationMessage_formats.present)
	if IndMsgType == 1 { //parsing cell metrics
		fmt.Printf(" parsing for cell metrics\n")
		indMsgFormat1_C := *(**C.E2SM_KPM_IndicationMessage_Format1_t)(unsafe.Pointer(&indicationmessage.indicationMessage_formats.choice[0]))
		no_of_cell := int32(indMsgFormat1_C.measData.list.count)
		fmt.Printf(" \n No of cell = %d\n", no_of_cell)
		//fmt.Println(no_of_cell)
		for n := int32(0); n < no_of_cell; n++ {
			var sizeof_MeasurementDataItem_t *C.MeasurementDataItem_t
			MeasurementDataItem_C := *(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat1_C.measData.list.array)) + (uintptr)(int(n))*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))
			no_of_cell_metrics := int32(MeasurementDataItem_C.measRecord.list.count)
			var CellM CellMetricsEntry
			v := reflect.ValueOf(CellM)
			fmt.Printf(" \n No of cell metrics = %d\n", no_of_cell_metrics)
			values := make(map[string]interface{}, v.NumField())
			//assert no_of_cell_metrics == v.NumField()   they both should be equal.
			for i := int32(0); i < no_of_cell_metrics; i++ {
				//fmt.Println(i)
				if v.Field(int(i)).CanInterface() {
					var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
					MeasurementRecordItem_C := *(**C.MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasurementDataItem_C.measRecord.list.array)) + (uintptr)(int(i))*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))
					type_var := int(MeasurementRecordItem_C.present)
					if type_var == 1 {
						var cast_integer *C.long = (*C.long)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
						values[v.Type().Field(int(i)).Name] = int32(*cast_integer)
					} else if type_var == 2 {
						var cast_float *C.double = (*C.double)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
						values[v.Type().Field(int(i)).Name] = float64(*cast_float)
					} else {
						fmt.Printf("Wrong Data Type")
					}

				} else {
					fmt.Printf("sorry you have a unexported field (lower case) value you are trying to sneak past. Can not allow it: %v\n", v.Type().Field(int(i)).Name)
				}
			} //end of inner for loop

			fmt.Println(values)
			fmt.Printf("Parsing Cell Metric Done")
			//c.writeCellMetrics_db(&values)//push cellmetrics map entry to database.
		} //end of outer for loop
		//end of if IndMsgType==1 , parsing cell metrics done

	} else if IndMsgType == 2 { //parsing ue metrics

		fmt.Printf(" parsing for UE metrics")
		indMsgFormat2_C := *(**C.E2SM_KPM_IndicationMessage_Format2_t)(unsafe.Pointer(&indicationmessage.indicationMessage_formats.choice[0]))
		no_of_ue_metrics := int32(indMsgFormat2_C.measData.list.count)
		fmt.Printf(" \n No of ue metrics = %d\n", no_of_ue_metrics)

		var sizeof_MeasurementDataItem_t *C.MeasurementDataItem_t
		MeasurementDataItem_C := *(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(0)*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))

		no_of_ue := int32(MeasurementDataItem_C.measRecord.list.count)
		fmt.Printf(" \n No of ue= %d\n", no_of_ue)
		for n := int32(0); n < no_of_ue; n++ {
			var UeM UeMetricsEntry
			v := reflect.ValueOf(UeM)
			values := make(map[string]interface{}, v.NumField())
			//assert no_of_ue_metrics == v.NumField()   they both should be equal.
			for i := int32(0); i < no_of_ue_metrics; i++ {
				//fmt.Println(i)
				if v.Field(int(i)).CanInterface() {

					var sizeof_MeasurementDataItem_t *C.MeasurementDataItem_t
					MeasurementDataItem_C := *(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(i)*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))
					var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
					MeasurementRecordItem_C := *(**C.MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasurementDataItem_C.measRecord.list.array)) + (uintptr)(n)*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))

					type_var := int(MeasurementRecordItem_C.present)
					if type_var == 1 {
						var cast_integer *C.long = (*C.long)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
						values[v.Type().Field(int(i)).Name] = int32(*cast_integer)
					} else if type_var == 2 {
						var cast_float *C.double = (*C.double)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
						values[v.Type().Field(int(i)).Name] = float64(*cast_float)

					} else {
						fmt.Printf("Wrong Data Type")
					}

				} else {
					fmt.Printf("sorry you have a unexported field (lower case) value you are trying to sneak past. Can not allow it: %v\n", v.Type().Field(int(i)).Name)
				}

			} //end of inner for loop
			fmt.Println(values)
			fmt.Printf("Parsing UE Metric Done")
			//c.writeUeMetrics_db(&values)//push UEmetrics map entry to database.

		} // end of outer for loop
		//parsing ue metrics done
	} else {
		fmt.Printf(" Invalid Indication message format")

	}

	return nil
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
	//var e2sm *E2sm

	indicationMsg, err := e2ap.GetIndicationMessage(params.Payload)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication message: %v", err)
		return
	}

	log.Printf("RIC Indication message from {%s} received", params.Meid.RanName)
	/*
		indicationHdr, err := e2sm.GetIndicationHeader(indicationMsg.IndHeader)
		if err != nil {
			xapp.Logger.Error("Failed to decode RIC Indication Header: %v", err)
			return
		}
	*/

	//Decoding message and put information into log
	//log.Printf("-----------RIC Indication Header-----------")
	//log.Printf("indicationMsg.IndHeader= %x", indicationMsg.IndHeader)
	/*
	   buf := new(bytes.Buffer) //create my buffer
	   binary.Write(buf, binary.LittleEndian, indicationMsg.IndHeader)
	   log.Printf("binary Write buf= %x",buf )
	   b := buf.Bytes()
	   //str := buf.String()
	   //log.Printf(" buf Strin()= %s",str )
	   //cptr1:= unsafe.Pointer(C.CString(str))
	   cptr1:= unsafe.Pointer(&b[0])
	   defer C.free(cptr1)
	*/

	if indicationMsg == nil || indicationMsg.IndHeader == nil || len(indicationMsg.IndHeader) == 0 ||
		indicationMsg.IndMessage == nil || len(indicationMsg.IndMessage) == 0 {
		return errors.New("Unable to get IndicationHeader or IndicationMessage due to invalid size")
	}

	cptr1 := unsafe.Pointer(&indicationMsg.IndHeader[0])
	decodedHdr := C.e2sm_decode_ric_indication_header(cptr1, C.size_t(len(indicationMsg.IndHeader)))
	//decodedHdr := C.e2sm_decode_ric_indication_header(cptr1, C.size_t(len(str)))
	//decodedHdr := C.e2sm_decode_ric_indication_header(cptr1, C.size_t(buf.Len()))
	if decodedHdr == nil {
		return errors.New("e2sm wrapper is unable to get IndicationHeader due to wrong or invalid input")
	}
	defer C.e2sm_free_ric_indication_header(decodedHdr)
	IndHdrType := int32(decodedHdr.indicationHeader_formats.present)
	if IndHdrType == 0 {
		log.Printf("No Indication Header present")
	}
	if IndHdrType == 1 {
		log.Printf("Indication Header format = %d", IndHdrType)
		/*
		   indHdrFormat1_C := *(**C.E2SM_KPM_IndicationHeader_Format1_t)(unsafe.Pointer(&decodedHdr.indicationHeader_formats.choice[0]))
		   //senderName_C := (*C.PrintableString_t)(unsafe.Pointer(indHdrFormat1_C.senderName))
		   senderName_C:=indHdrFormat1_C.senderName
		   var senderName []byte
		   senderName = C.GoBytes(unsafe.Pointer(senderName_C.buf), C.int(senderName_C.size))
		   //log.Printf("Sender Name = %x",senderName)

		   //senderType_C := (*C.PrintableString_t)(unsafe.Pointer(indHdrFormat1_C.senderType))
		   senderType_C :=indHdrFormat1_C.senderType
		   //senderType []byte
		   senderType := C.GoBytes(unsafe.Pointer(senderType_C.buf), C.int(senderType_C.size))
		   //log.Printf("Sender Type = %x",senderType)

		   //vendorName_C := (*C.PrintableString_t)(unsafe.Pointer(indHdrFormat1_C.vendorName))
		   vendorName_C :=indHdrFormat1_C.vendorName
		   //vendorName  []byte
		   vendorName := C.GoBytes(unsafe.Pointer(vendorName_C.buf), C.int(vendorName_C.size))
		   //log.Printf("Vendor Name = %x",vendorName)
		*/

	}

	/*
	   indMsg, err := e2sm.GetIndicationMessage(indicationMsg.IndMessage)
	   if err != nil {
	           xapp.Logger.Error("Failed to decode RIC Indication Message: %v", err)
	           return
	   }
	*/
	//log.Printf("-----------RIC Indication Message-----------")
	//log.Printf("indicationMsg.IndMessage= %x",indicationMsg.IndMessage)
	cptr2 := unsafe.Pointer(&indicationMsg.IndMessage[0])
	indicationmessage := C.e2sm_decode_ric_indication_message(cptr2, C.size_t(len(indicationMsg.IndMessage)))
	if indicationmessage == nil {
		return errors.New("e2sm wrapper is unable to get IndicationMessage due to wrong or invalid input")
	}
	defer C.e2sm_free_ric_indication_message(indicationmessage)
	IndMsgType := int32(indicationmessage.indicationMessage_formats.present)
	if IndMsgType == 1 { //parsing cell metrics
		fmt.Printf(" parsing for cell metrics\n")
		indMsgFormat1_C := *(**C.E2SM_KPM_IndicationMessage_Format1_t)(unsafe.Pointer(&indicationmessage.indicationMessage_formats.choice[0]))
		no_of_cell := int32(indMsgFormat1_C.measData.list.count)
		fmt.Printf(" \n No of cell = %d\n", no_of_cell)
		//fmt.Println(no_of_cell)
		for n := int32(0); n < no_of_cell; n++ {
			var sizeof_MeasurementDataItem_t *C.MeasurementDataItem_t
			MeasurementDataItem_C := *(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat1_C.measData.list.array)) + (uintptr)(int(n))*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))
			no_of_cell_metrics := int32(MeasurementDataItem_C.measRecord.list.count)
			//var CellM CellMetricsEntry
			//v := reflect.ValueOf(CellM)
			fmt.Printf(" \n No of cell metrics = %d\n", no_of_cell_metrics)
			//values := make(map[string]interface{}, v.NumField())
			values := make(map[string]interface{})
			CellKpi := ranCellKpi[params.Meid.RanName]
			var tmpCellKpi []string
			for _, ele := range CellKpi {
				if ele != "" {
					tmpCellKpi = append(tmpCellKpi, ele)
				}
			}

			//assert no_of_cell_metrics == v.NumField()   they both should be equal.
			if int(no_of_cell_metrics) != len(tmpCellKpi) {
				log.Printf("no_of_cell_metrics !=  len(tmpCellKpi)")
				return errors.New("no_of_cell_metrics  !=len(tmpCellKpi)")
			}
			for i := int32(0); i < no_of_cell_metrics; i++ {
				//fmt.Println(i)
				//if v.Field(int(i)).CanInterface() {
				var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
				MeasurementRecordItem_C := *(**C.MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasurementDataItem_C.measRecord.list.array)) + (uintptr)(int(i))*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))
				type_var := int(MeasurementRecordItem_C.present)
				if type_var == 1 {
					var cast_integer *C.long = (*C.long)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
					//values[v.Type().Field(int(i)).Name]=int32(*cast_integer)
					values[tmpCellKpi[int(i)]] = int32(*cast_integer)
				} else if type_var == 2 {
					var cast_float *C.double = (*C.double)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
					//values[v.Type().Field(int(i)).Name]=float64(*cast_float)
					values[tmpCellKpi[int(i)]] = float64(*cast_float)
				} else {
					fmt.Printf("Wrong Data Type")
				}

				// }else {
				//fmt.Printf("sorry you have a unexported field (lower case) value you are trying to sneak past. Can not allow it: %v\n", v.Type().Field(int(i)).Name)
				//}
			} //end of inner for loop

			fmt.Println(values)
			fmt.Printf("Parsing Cell Metric Done")
			c.writeCellMetrics_db(&values) //push cellmetrics map entry to database.
		} //end of outer for loop
		//end of if IndMsgType==1 , parsing cell metrics done

	} else if IndMsgType == 2 { //parsing ue metrics

		fmt.Printf(" parsing for UE metrics")
		indMsgFormat2_C := *(**C.E2SM_KPM_IndicationMessage_Format2_t)(unsafe.Pointer(&indicationmessage.indicationMessage_formats.choice[0]))
		//no_of_ue_metrics:=int32(indMsgFormat2_C .measData.list.count)
		//fmt.Printf(" \n No of ue metrics = %d\n",no_of_ue_metrics)

		var sizeof_MeasurementDataItem_t *C.MeasurementDataItem_t
		MeasurementDataItem_C := *(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(0)*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))
		UeKpi := ranUeKpi[params.Meid.RanName]
		var tmpUeKpi []string
		for _, ele := range UeKpi {
			if ele != "" {
				tmpUeKpi = append(tmpUeKpi, ele)
			}
		}

		no_of_ue := int32(MeasurementDataItem_C.measRecord.list.count) / int32(len(tmpUeKpi))
		no_of_ue_metrics := int32(len(tmpUeKpi))
		fmt.Printf(" \n No of ue= %d\n", no_of_ue)

		for n := int32(0); n < no_of_ue; n++ {
			//var UeM UeMetricsEntry
			//v := reflect.ValueOf(UeM)
			//values := make(map[string]interface{}, v.NumField())
			values := make(map[string]interface{})
			//assert no_of_ue_metrics == v.NumField()   they both should be equal.
			//if (int(no_of_ue_metrics) != v.NumField()){
			//log.Printf("no_of_ue_metrics != v.NumField()")
			//return errors.New("no_of_ue_metrics != v.NumField()")
			//}
			for i := int32(0); i < no_of_ue_metrics; i++ {
				//fmt.Println(i)
				//if v.Field(int(i)).CanInterface() {

				//var sizeof_MeasurementDataItem_t  *C.MeasurementDataItem_t
				//MeasurementDataItem_C:=*(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(i)*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))

				var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
				MeasurementRecordItem_C := *(**C.MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasurementDataItem_C.measRecord.list.array)) + (uintptr)(n+(i*no_of_ue))*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))
				//var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
				//MeasurementRecordItem_C:=*(**C.MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasurementDataItem_C.measRecord.list.array)) + (uintptr)(n)*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))

				type_var := int(MeasurementRecordItem_C.present)
				if type_var == 1 {
					var cast_integer *C.long = (*C.long)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
					values[tmpUeKpi[int(i)]] = int32(*cast_integer)
				} else if type_var == 2 {
					var cast_float *C.double = (*C.double)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
					values[tmpUeKpi[int(i)]] = float64(*cast_float)

				} else {
					fmt.Printf("Wrong Data Type")
				}

				// }else {
				// fmt.Printf("sorry you have a unexported field (lower case) value you are trying to sneak past. Can not allow it: %v\n", v.Type().Field(int(i)).Name)
				// }

			} //end of inner for loop
			fmt.Println(values)
			fmt.Printf("Parsing UE Metric Done")
			c.writeUeMetrics_db(&values) //push UEmetrics map entry to database.

		} // end of outer for loop
		//parsing ue metrics done
	} else {
		fmt.Printf(" Invalid Indication message format")

	}

	return nil

}

// func (c *Control) writeUeMetrics_db(ueMetrics UeMetricsEntry) {
// 	//Write metric to InfluxDB using API
// 	writeAPI := c.client.WriteAPIBlocking("my-org", "kpimon")
// 	ueMetricsJSON, err := json.Marshal(ueMetrics)
// 	if err != nil {
// 		xapp.Logger.Error("Marshal UE Metrics failed!")
// 	}
// 	p := influxdb2.NewPointWithMeasurement("ricIndication_UeMetrics").
// 		AddField("UE Metrics", ueMetricsJSON).
// 		SetTime(time.Now())
// 	writeAPI.WritePoint(context.Background(), p)
// 	xapp.Logger.Info("Wrote UE Metrics to InfluxDB")
// }

// func (c *Control) writeCellMetrics_db(cellMetrics CellMetricsEntry) {
// 	writeAPI := c.client.WriteAPIBlocking("my-org", "kpimon")
// 	cellMetricsJSON, er := json.Marshal(cellMetrics)
// 	if er != nil {
// 		xapp.Logger.Error("Marshal Cell Metrics failed!")
// 	}
// 	p := influxdb2.NewPointWithMeasurement("ricIndication_cellMetrics").
// 		AddField("Cell Metrics", cellMetricsJSON).
// 		SetTime(time.Now())
// 	writeAPI.WritePoint(context.Background(), p)
// 	xapp.Logger.Info("Wrote Cell Metrics to InfluxDB")
// }

/*
	func (c *Control) queryUEReports() {
		log.Printf("query UE")
		resp, err := http.Get("http://10.244.0.68/sba/influx/query?db=RIC-Test-static&q=select+*+from+UEReports")
		if err != nil {
			log.Fatalln(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		var results ViaviMessages
		er := json.Unmarshal(body, &results)
		if er != nil {
			panic(er)
		}
		// 	sb := string(body)
		//    log.Printf(sb)
		c.writeUeMetrics_db(results.results.series)
	}

	func (c *Control) queryCellReports() {
		log.Printf("query cell")
		resp, err := http.Get("http://10.244.0.68/sba/influx/query?db=RIC-Test-static&q=select+*+from+CellReports")
		if err != nil {
			log.Fatalln(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		var results ViaviMessages
		er := json.Unmarshal(body, &results)
		if er != nil {
			panic(er)
		}
		c.writeCellMetrics_db(results.results.series)
	}
*/
func (c *Control) writeUeMetrics_db(ueMetrics *map[string]interface{}) {
	writeAPI := c.client.WriteAPIBlocking("my-org", "kpimon")
	/*
		ueMetricsJSON, err := json.Marshal(ueMetrics)
		if err != nil {
			xapp.Logger.Error("Marshal UE Metrics failed!")
		}

		p := influxdb2.NewPointWithMeasurement("UeMetrics").
			AddField("UE Metrics", ueMetricsJSON).
			SetTime(time.Now())
	*/
	p := influxdb2.NewPointWithMeasurement("UeMetrics").SetTime(time.Now())
	for key, value := range *ueMetrics {
		p = p.AddField(key, value)
	}

	writeAPI.WritePoint(context.Background(), p)
	xapp.Logger.Info("Wrote UE Metrics to InfluxDB")
}

func (c *Control) writeCellMetrics_db(cellMetrics *map[string]interface{}) {
	writeAPI := c.client.WriteAPIBlocking("my-org", "kpimon")
	/*
		cellMetricsJSON, er := json.Marshal(cellMetrics)
		if er != nil {
			xapp.Logger.Error("Marshal Cell Metrics failed!")
		}

		p := influxdb2.NewPointWithMeasurement("cellMetrics").
			AddField("Cell Metrics", cellMetricsJSON).
			SetTime(time.Now())
	*/
	p := influxdb2.NewPointWithMeasurement("cellMetrics").SetTime(time.Now())
	for key, value := range *cellMetrics {
		p = p.AddField(key, value)
	}
	writeAPI.WritePoint(context.Background(), p)
	xapp.Logger.Info("Wrote Cell Metrics to InfluxDB")
}

func (c Control) xAppStartCB(d interface{}) {
	xapp.Logger.Info("In callback KPI monitor xApp ...")

	//ranUeKpi ranCellKpi
	// Get eNodeB list
	nbList := c.getnbList()

	for _, nb := range nbList {
		if nb.ConnectionStatus == 1 {
			xapp.Logger.Info("Building ranCellKp for %v", nb.InventoryName)
			link := "http://service-ricplt-e2mgr-http.ricplt.svc.cluster.local:3800/v1/nodeb/"
			link = link + nb.InventoryName
			tmpr, err := http.Get(link)
			if err != nil {
				log.Fatalln(err)
				panic("Unable to build ranCellKpi")
			}
			defer tmpr.Body.Close()
			var resp E2mgrResponse

			err = json.NewDecoder(tmpr.Body).Decode(&resp)
			if err != nil {
				log.Fatalln(err)
				panic("Unable to build ranCellKpi")

			}
			counter := 0
			//RanFunctionId=2 for kpm in viavi
			for i := 0; i < len(resp.Gnb.RanFunctions); i++ {
				if resp.Gnb.RanFunctions[i].RanFunctionId == 2 {
					counter = i
					break
				}
			}
			cString := C.CString(resp.Gnb.RanFunctions[counter].RanFunctionDefinition)
			defer C.free(unsafe.Pointer(cString)) // Free the allocated C string when done
			result := C.buildRanCellUeKpi(cString)

			ueSlice := make([]string, result.ueKpiSize)

			for _, v := range unsafe.Slice(result.ueKpi, result.ueKpiSize) {
				ueSlice = append(ueSlice, C.GoString(v))
			}
			ranUeKpi[nb.InventoryName] = ueSlice
			//fmt.Println("len of ranUeKpi= ",len(ranUeKpi))
			//fmt.Println("ranUeKpi map = ", ranUeKpi)

			cellSlice := make([]string, result.cellKpiSize)

			for _, v := range unsafe.Slice(result.cellKpi, result.cellKpiSize) {
				cellSlice = append(cellSlice, C.GoString(v))
			}
			ranCellKpi[nb.InventoryName] = cellSlice
			//fmt.Println("len of ranCellKpi= ",len(ranCellKpi))
			//fmt.Println("ranCellKpi map = ", ranCellKpi)
			/*
							counter = 0
				        		for i := 0; i < len(resp.Gnb.NodeConfigs); i++ {
				                		if resp.Gnb.NodeConfigs[i].E2nodeComponentInterfaceType == "f1" {
				                        		counter = i
				                        		break
				                		}
				        		}


				        		tm := resp.Gnb.NodeConfigs[counter].E2nodeComponentRequestPart
				       			base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(tm)))
				        		nl, _ := base64.StdEncoding.Decode(base64Text, []byte(tm))
				        		message := string(base64Text[:nl])
			*/
			/*
							cString2 := C.CString(message)
							fmt.Println("Ponga = ",tm)
							fmt.Println("Ponga = ",message)
							defer C.free(unsafe.Pointer(cString2)) // Free the allocated C string when done
							result2:=C.decodeF1apGetCellIds(cString2)
							cellList:=make([]string, result2.size)
							for _, v := range unsafe.Slice(result2.cellids, result2.size) {
				                                cellList = append(cellList, C.GoString(v))
				                        }
							plmnList:=make([]string, result2.size)
				                        for _, v := range unsafe.Slice(result2.plmn, result2.size) {
				                                plmnList = append(plmnList, C.GoString(v))
				                        }

			*/
			/*
				var f1ap *F1ap
				fmt.Println("Ponga = ",tm)
				cellList,plmnList:=f1ap.F1apGetCellIds(message)
				Glob_Ran_cell[nb.InventoryName]=cellList
				for i := 0; i < len(cellList); i++ {
					Glob_cell_Plmn[cellList[i]]=plmnList[i]
				}
			*/
			response, err4 := http.Get("http://localhost:8090/getdata?gnbid=" + nb.InventoryName)
			if err4 != nil {
				fmt.Println("Error:", err4)
				panic(err4)
			}
			defer response.Body.Close()

			var response2 APIResponse
			if err5 := json.NewDecoder(response.Body).Decode(&response2); err5 != nil {
				fmt.Println("Error decoding response:", err5)
				panic(err5)
			}

			fmt.Printf("f1apSever Response: %+v\n", response2)

			Glob_Ran_cell[nb.InventoryName] = response2.Cellids
			for i := 0; i < len(response2.Cellids); i++ {
				Glob_cell_Plmn[response2.Cellids[i]] = response2.PlmnIds[i]
			}

			//C.freeMemorydRanCellUeKpi(result)

		}

	}

	for {
		time.Sleep(5 * time.Second)
		if xapp.IsRegistered() {
			xapp.Logger.Info("App registration is done, ready to send subscription request.")
			break
		}
		xapp.Logger.Debug("App registration is not done yet, sleep 5s and check again")
	}
	// Send subscription request to connected NodeB
	for _, nb := range nbList {
		if nb.ConnectionStatus == 1 {
			xapp.Logger.Info("Before send subscription request to %v", nb.InventoryName)
			c.sendSubscription(nb.InventoryName)
			xapp.Logger.Info("After send subscription request to %v", nb.InventoryName)
		}

	}

	fmt.Println("len of ranUeKpi= ", len(ranUeKpi))
	fmt.Println("ranUeKpi map = ", ranUeKpi)

	fmt.Println("len of ranCellKpi= ", len(ranCellKpi))
	fmt.Println("ranCellKpi map = ", ranCellKpi)

	fmt.Println("len of Glob_cell_Plmn= ", len(Glob_cell_Plmn))
	fmt.Println("Glob_cell_Plmn map = ", Glob_cell_Plmn)

	go c.controlLoop()
	//go c.queryUEReports()
	//go c.queryCellReports()
	xapp.Logger.Info("End callback KPI monitor xApp ...")
}

func (c Control) Run() {
	// Setup level
	xapp.Logger.SetLevel(xapp.Config.GetInt("logger.level"))
	// Register callback
	xapp.Logger.Info("In Run() ...")
	xapp.SetReadyCB(c.xAppStartCB, true)
	// Start xApp
	xapp.Run(c)
}
