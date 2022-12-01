package control
/*
#include <e2sm/wrapper.h>
#cgo LDFLAGS: -le2smwrapper -lm
#cgo CFLAGS: -I/usr/local/include/e2sm
*/
import "C"

import (
	"unsafe"
)

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
//	"bytes"
//	"encoding/binary"
	"strconv"
	"encoding/base64"
	"strings"
	"fmt"
	"reflect"
	"errors"
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
	funcId               = int64(2)
	hPort                = int64(8080)
	rPort                = int64(4560)
	clientEndpoint       = clientmodel.SubscriptionParamsClientEndpoint{Host: "service-ricxapp-kpimon-go-http.ricxapp", HTTPPort: &hPort, RMRPort: &rPort}
)
var Glob_cell = make(map[string]bool)
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
	xapp.Logger.Info("returning control\n")
	return Control{
		make(chan *xapp.RMRParams),
		influxdb2.NewClient("http://ricplt-influxdb.ricplt:8086", "client"),
	}
}
func create_db() {
	//Create a database named kpimon in influxDB
	xapp.Logger.Info("In create_db\n")
	_, err := http.Post("http://ricplt-influxdb.ricplt:8086/query?q=create%20database%20kpimon", "", nil)
	if err != nil {
		xapp.Logger.Error("Create database failed!")
	}
	xapp.Logger.Info("exiting create_db\n")
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
func encode_action_format1(plmn string, cellid string) clientmodel.ActionDefinition {
	lol1 := plmnid_to_list_of_int(plmn)
	lol2 := cellid_to_list_of_int(cellid)
	var format1 []int64
	//format1=[]int64{0,1,1,8,0,19,32,0,3,1,32,0,0,32,0,4,1,32,0,0,32,0,78,1,32,0,0,32,0,79,1,32,0,0,32,0,8,1,32,0,0,32,0,7,1,32,0,0,32,0,11,1,32,0,0,32,0,12,1,32,0,0,32,0,80,1,32,0,0,32,0,81,1,32,0,0,32,0,13,1,32,0,0,32,0,14,1,32,0,0,32,0,40,1,32,0,0,32,0,41,1,32,0,0,32,0,42,1,32,0,0,32,0,82,1,32,0,0,32,0,83,1,32,0,0,32,0,84,1,32,0,0,32,0,85,1,32,0,0,32,0,86,1,32,0,0,64,39,15,0,0,31,1,18,52,92,0,16}
	//0(nr) 1(eutra)  0,31,1(plmn)18,52,92,0,16(cellid)

	//for simulation-by measID (not supported in viavi 1.4)
	//format1=[]int64{0,1,1,8,0,22,32,0,3,1,32,0,0,32,0,4,1,32,0,0,32,0,78,1,32,0,0,32,0,79,1,32,0,0,32,0,80,1,32,0,0,32,0,81,1,32,0,0,32,0,8,1,32,0,0,32,0,7,1,32,0,0,32,0,11,1,32,0,0,32,0,12,1,32,0,0,32,0,82,1,32,0,0,32,0,83,1,32,0,0,32,0,13,1,32,0,0,32,0,14,1,32,0,0,32,0,40,1,32,0,0,32,0,41,1,32,0,0,32,0,42,1,32,0,0,32,0,84,1,32,0,0,32,0,85,1,32,0,0,32,0,86,1,32,0,0,32,0,87,1,32,0,0,32,0,88,1,32,0,0,32,0,89,1,32,0,0,64,39,15,0}//assuming nr cells

	//for simulation-by measName(supported in Viavi 1.4)
	format1=[]int64{0,1,1,8,0,22,0,160,68,82,66,46,85,69,84,104,112,68,108,1,32,0,0,0,160,68,82,66,46,85,69,84,104,112,85,108,1,32,0,0,0,176,80,69,69,46,65,118,103,80,111,119,101,114,1,32,0,0,0,144,80,69,69,46,69,110,101,114,103,121,1,32,0,0,1,144,81,111,115,70,108,111,119,46,84,111,116,80,100,99,112,80,100,117,86,111,108,117,109,101,68,108,1,32,0,0,1,144,81,111,115,70,108,111,119,46,84,111,116,80,100,99,112,80,100,117,86,111,108,117,109,101,85,108,1,32,0,0,0,160,82,82,67,46,67,111,110,110,77,97,120,1,32,0,0,0,176,82,82,67,46,67,111,110,110,77,101,97,110,1,32,0,0,0,208,82,82,85,46,80,114,98,65,118,97,105,108,68,108,1,32,0,0,0,208,82,82,85,46,80,114,98,65,118,97,105,108,85,108,1,32,0,0,0,176,82,82,85,46,80,114,98,84,111,116,68,108,1,32,0,0,0,176,82,82,85,46,80,114,98,84,111,116,85,108,1,32,0,0,0,192,82,82,85,46,80,114,98,85,115,101,100,68,108,1,32,0,0,0,192,82,82,85,46,80,114,98,85,115,101,100,85,108,1,32,0,0,0,160,86,105,97,118,105,46,71,101,111,46,120,1,32,0,0,0,160,86,105,97,118,105,46,71,101,111,46,121,1,32,0,0,0,160,86,105,97,118,105,46,71,101,111,46,122,1,32,0,0,0,192,86,105,97,118,105,46,71,110,98,68,117,73,100,1,32,0,0,0,160,86,105,97,118,105,46,78,114,67,103,105,1,32,0,0,0,160,86,105,97,118,105,46,78,114,80,99,105,1,32,0,0,1,96,86,105,97,118,105,46,82,97,100,105,111,46,97,110,116,101,110,110,97,84,121,112,101,1,32,0,0,1,32,86,105,97,118,105,46,82,97,100,105,111,46,97,122,105,109,117,116,104,1,32,0,0,1,0,86,105,97,118,105,46,82,97,100,105,111,46,112,111,119,101,114,1,32,0,0,64,39,15,0}//assuming nr cells
	//for e2test
	 // format1=[]int64{0,1,1,8,0,42,32,0,0,1,32,0,0,32,0,1,1,32,0,0,32,0,2,1,32,0,0,32,0,3,1,32,0,0,32,0,4,1,32,0,0,32,0,5,1,32,0,0,32,0,6,1,32,0,0,32,0,7,1,32,0,0,32,0,8,1,32,0,0,32,0,9,1,32,0,0,32,0,10,1,32,0,0,32,0,11,1,32,0,0,32,0,12,1,32,0,0,32,0,13,1,32,0,0,32,0,14,1,32,0,0,32,0,15,1,32,0,0,32,0,16,1,32,0,0,32,0,17,1,32,0,0,32,0,18,1,32,0,0,32,0,19,1,32,0,0,32,0,20,1,32,0,0,32,0,21,1,32,0,0,32,0,22,1,32,0,0,32,0,23,1,32,0,0,32,0,24,1,32,0,0,32,0,25,1,32,0,0,32,0,26,1,32,0,0,32,0,27,1,32,0,0,32,0,28,1,32,0,0,32,0,29,1,32,0,0,32,0,30,1,32,0,0,32,0,31,1,32,0,0,32,0,32,1,32,0,0,32,0,33,1,32,0,0,32,0,34,1,32,0,0,32,0,35,1,32,0,0,32,0,36,1,32,0,0,32,0,37,1,32,0,0,32,0,38,1,32,0,0,32,0,39,1,32,0,0,32,0,40,1,32,0,0,32,0,41,1,32,0,0,32,0,42,1,32,0,0,64,39,15,0}
	format1 = append(format1, lol1...) //appending plmn
        format1 = append(format1, lol2...) //appending cellid
	return format1
}

func encode_action_format2() clientmodel.ActionDefinition{
	var format2 []int64
	format2 = []int64{0, 1, 0, 0, 0, 20, 0, 160, 68, 82, 66, 46, 85, 69, 84, 104, 112, 68, 108, 1, 0, 0, 0, 1, 64, 68, 82, 66, 46, 85, 69, 84, 104, 112, 85, 108, 1, 0, 0, 0, 1, 0, 71, 78, 66, 45, 68, 85, 45, 73, 68, 1, 0, 0, 0, 0, 160, 78, 82, 45, 67, 71, 73, 1, 0, 0, 0, 0, 160, 78, 82, 45, 80, 67, 73, 1, 0, 0, 0, 2, 192, 81, 111, 115, 70, 108, 111, 119, 46, 80, 100, 99, 112, 80, 100, 117, 86, 111, 108, 117, 109, 101, 68, 108, 1, 0, 0, 0, 2, 192, 81, 111, 115, 70, 108, 111, 119, 46, 80, 100, 99, 112, 80, 100, 117, 86, 111, 108, 117, 109, 101, 85, 108, 1, 0, 0, 0, 1, 64, 82, 82, 67, 46, 67, 111, 110, 110, 77, 97, 120, 1, 0, 0, 0, 1, 96, 82, 82, 67, 46, 67, 111, 110, 110, 77, 101, 97, 110, 1, 0, 0, 0, 1, 160, 82, 82, 85, 46, 80, 114, 98, 65, 118, 97, 105, 108, 68, 108, 1, 0, 0, 0, 1, 160, 82, 82, 85, 46, 80, 114, 98, 65, 118, 97, 105, 108, 85, 108, 1, 0, 0, 0, 1, 32, 82, 82, 85, 46, 80, 114, 98, 84, 111, 116, 1, 0, 0, 0, 1, 96, 82, 82, 85, 46, 80, 114, 98, 84, 111, 116, 68, 108, 1, 0, 0, 0, 1, 96, 82, 82, 85, 46, 80, 114, 98, 84, 111, 116, 85, 108, 1, 0, 0, 0, 1, 128, 82, 82, 85, 46, 80, 114, 98, 85, 115, 101, 100, 68, 108, 1, 0, 0, 0, 1, 128, 82, 82, 85, 46, 80, 114, 98, 85, 115, 101, 100, 85, 108, 1, 0, 0, 0, 1, 64, 86, 105, 97, 118, 105, 46, 71, 101, 111, 46, 120, 1, 0, 0, 0, 1, 64, 86, 105, 97, 118, 105, 46, 71, 101, 111, 46, 121, 1, 0, 0, 0, 1, 64, 86, 105, 97, 118, 105, 46, 71, 101, 111, 46, 122, 1, 0, 0, 0, 2, 0, 86, 105, 97, 118, 105, 46, 82, 97, 100, 105, 111, 46, 112, 111, 119, 101, 114, 1, 0, 0, 0, 2, 64, 86, 105, 97, 118, 105, 46, 82, 97, 100, 105, 111, 46, 115, 101, 99, 116, 111, 114, 115, 1, 0, 0, 0, 0, 0}
	//encode the variable part and append it to our array.
	format2 = append(format2, 89) //appending variable part if necessory
	return format2
}

func encode_action_format3() clientmodel.ActionDefinition {
	var format3 []int64

	//for simulation-by measId(not supported in Viavi 1.4)
	//format3=[]int64{0,1,3,64,0,29,32,0,90,0,0,16,0,0,16,0,91,0,0,16,0,0,16,0,3,0,0,16,0,0,16,0,4,0,0,16,0,0,16,0,80,0,0,16,0,0,16,0,13,0,0,16,0,0,16,0,14,0,0,16,0,0,16,0,92,0,0,16,0,0,16,0,93,0,0,16,0,0,16,0,44,0,0,16,0,0,16,0,40,0,0,16,0,0,16,0,41,0,0,16,0,0,16,0,42,0,0,16,0,0,16,0,94,0,0,16,0,0,16,0,95,0,0,16,0,0,16,0,96,0,0,16,0,0,16,0,97,0,0,16,0,0,16,0,45,0,0,16,0,0,16,0,98,0,0,16,0,0,16,0,99,0,0,16,0,0,16,0,100,0,0,16,0,0,16,0,101,0,0,16,0,0,16,0,102,0,0,16,0,0,16,0,103,0,0,16,0,0,16,0,104,0,0,16,0,0,16,0,43,0,0,16,0,0,16,0,105,0,0,16,0,0,16,0,106,0,0,16,0,0,16,0,107,0,0,16,0,0,16,0,108,0,0,16,0,0,32,39,15}
	//variable part is not presetnt in action def format 3

	//for simulation-by measName(supported in Viavi 1.4)
	format3=[]int64{0,1,3,64,0,29,0,160,68,82,66,46,85,69,67,113,105,68,108,0,0,16,0,0,0,80,68,82,66,46,85,69,67,113,105,85,108,0,0,16,0,0,0,80,68,82,66,46,85,69,84,104,112,68,108,0,0,16,0,0,0,80,68,82,66,46,85,69,84,104,112,85,108,0,0,16,0,0,0,200,81,111,115,70,108,111,119,46,84,111,116,80,100,99,112,80,100,117,86,111,108,117,109,101,68,108,0,0,16,0,0,0,96,82,82,85,46,80,114,98,85,115,101,100,68,108,0,0,16,0,0,0,96,82,82,85,46,80,114,98,85,115,101,100,85,108,0,0,16,0,0,0,80,84,66,46,84,111,116,78,98,114,68,108,0,0,16,0,0,0,80,84,66,46,84,111,116,78,98,114,85,108,0,0,16,0,0,0,96,86,105,97,118,105,46,67,101,108,108,46,105,100,0,0,16,0,0,0,80,86,105,97,118,105,46,71,101,111,46,120,0,0,16,0,0,0,80,86,105,97,118,105,46,71,101,111,46,121,0,0,16,0,0,0,80,86,105,97,118,105,46,71,101,111,46,122,0,0,16,0,0,0,96,86,105,97,118,105,46,81,111,83,46,53,113,105,0,0,16,0,0,0,120,86,105,97,118,105,46,81,111,83,46,67,101,108,108,73,100,0,0,16,0,0,0,112,86,105,97,118,105,46,81,111,83,46,68,114,98,73,100,0,0,16,0,0,0,104,86,105,97,118,105,46,81,111,83,46,71,102,98,114,0,0,16,0,0,0,104,86,105,97,118,105,46,83,108,105,99,101,46,105,100,0,0,16,0,0,0,112,86,105,97,118,105,46,85,69,46,66,101,97,109,73,100,0,0,16,0,0,0,128,86,105,97,118,105,46,85,69,46,70,114,97,109,101,67,110,116,0,0,16,0,0,0,112,86,105,97,118,105,46,85,69,46,82,115,83,105,110,114,0,0,16,0,0,0,96,86,105,97,118,105,46,85,69,46,82,115,114,112,0,0,16,0,0,0,96,86,105,97,118,105,46,85,69,46,82,115,114,113,0,0,16,0,0,0,112,86,105,97,118,105,46,85,69,46,84,99,82,110,116,105,0,0,16,0,0,0,136,86,105,97,118,105,46,85,69,46,97,110,111,109,97,108,105,101,115,0,0,16,0,0,0,80,86,105,97,118,105,46,85,69,46,105,100,0,0,16,0,0,0,184,86,105,97,118,105,46,85,69,46,115,101,114,118,105,110,103,68,105,115,116,97,110,99,101,0,0,16,0,0,0,104,86,105,97,118,105,46,85,69,46,115,112,101,101,100,0,0,16,0,0,0,208,86,105,97,118,105,46,85,69,46,116,97,114,103,101,116,84,104,114,111,117,103,104,112,117,116,68,108,0,0,16,0,0,0,208,86,105,97,118,105,46,85,69,46,116,97,114,103,101,116,84,104,114,111,117,103,104,112,117,116,85,108,0,0,16,0,0,32,39,15}

	//variable part is not presetnt in action def format 3

	//for e2test
	//format3=[]int64{0,1,3,64,0,62,32,0,43,0,0,16,0,0,16,0,44,0,0,16,0,0,16,0,45,0,0,16,0,0,16,0,3,0,0,16,0,0,16,0,4,0,0,16,0,0,16,0,13,0,0,16,0,0,16,0,14,0,0,16,0,0,16,0,46,0,0,16,0,0,16,0,47,0,0,16,0,0,16,0,48,0,0,16,0,0,16,0,49,0,0,16,0,0,16,0,50,0,0,16,0,0,16,0,51,0,0,16,0,0,16,0,52,0,0,16,0,0,16,0,53,0,0,16,0,0,16,0,54,0,0,16,0,0,16,0,55,0,0,16,0,0,16,0,56,0,0,16,0,0,16,0,57,0,0,16,0,0,16,0,58,0,0,16,0,0,16,0,59,0,0,16,0,0,16,0,60,0,0,16,0,0,16,0,61,0,0,16,0,0,16,0,62,0,0,16,0,0,16,0,63,0,0,16,0,0,16,0,64,0,0,16,0,0,16,0,65,0,0,16,0,0,16,0,66,0,0,16,0,0,16,0,67,0,0,16,0,0,16,0,68,0,0,16,0,0,16,0,69,0,0,16,0,0,16,0,70,0,0,16,0,0,16,0,71,0,0,16,0,0,16,0,72,0,0,16,0,0,16,0,73,0,0,16,0,0,16,0,74,0,0,16,0,0,16,0,75,0,0,16,0,0,16,0,76,0,0,16,0,0,16,0,77,0,0,16,0,0,16,0,27,0,0,16,0,0,16,0,19,0,0,16,0,0,16,0,26,0,0,16,0,0,16,0,39,0,0,16,0,0,16,0,15,0,0,16,0,0,16,0,16,0,0,16,0,0,16,0,17,0,0,16,0,0,16,0,18,0,0,16,0,0,16,0,20,0,0,16,0,0,16,0,21,0,0,16,0,0,16,0,22,0,0,16,0,0,16,0,23,0,0,16,0,0,16,0,28,0,0,16,0,0,16,0,29,0,0,16,0,0,16,0,30,0,0,16,0,0,16,0,31,0,0,16,0,0,16,0,32,0,0,16,0,0,16,0,33,0,0,16,0,0,16,0,34,0,0,16,0,0,16,0,35,0,0,16,0,0,16,0,36,0,0,16,0,0,16,0,40,0,0,16,0,0,16,0,41,0,0,16,0,0,16,0,42,0,0,16,0,0,32,39,15}
	//variable part is not presetnt in action def format 3
	return format3
}
func encode_actionsToBeSetup(meid string) clientmodel.ActionsToBeSetup {
	var l clientmodel.ActionsToBeSetup
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

	var n int64 = 1
	//var ue int64 = 1 //get no of ue connected to du(if required)

	//for action def 1
	for n <= int64(len(cells)) {
		var tmp int64 = n
		var lol *int64 = &tmp
		s := clientmodel.ActionToBeSetup{
			ActionID:         lol,
			ActionType:       &actionType,
			ActionDefinition: encode_action_format1(resp.GlobalNbId.PlmnId, cells[n-1]),
			SubsequentAction:  &clientmodel.SubsequentAction{
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
		ActionDefinition: encode_action_format3(),
		SubsequentAction:  &clientmodel.SubsequentAction{
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
					8,39,15,
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
        if IndHdrType==0{
                log.Printf("No Indication Header present")
        }
        if IndHdrType==1{
                log.Printf("Indication Header format = %d",IndHdrType)
                indHdrFormat1_C := *(**C.E2SM_KPM_IndicationHeader_Format1_t)(unsafe.Pointer(&decodedHdr.indicationHeader_formats.choice[0]))
                //senderName_C := (*C.PrintableString_t)(unsafe.Pointer(indHdrFormat1_C.senderName))
		senderName_C:=indHdrFormat1_C.senderName
                var senderName []byte
                senderName = C.GoBytes(unsafe.Pointer(senderName_C.buf), C.int(senderName_C.size))
                log.Printf("Sender Name = %x",senderName)

                //senderType_C := (*C.PrintableString_t)(unsafe.Pointer(indHdrFormat1_C.senderType))
		senderType_C :=indHdrFormat1_C.senderType
                //senderType []byte
                senderType := C.GoBytes(unsafe.Pointer(senderType_C.buf), C.int(senderType_C.size))
                log.Printf("Sender Type = %x",senderType)

                //vendorName_C := (*C.PrintableString_t)(unsafe.Pointer(indHdrFormat1_C.vendorName))
		vendorName_C :=indHdrFormat1_C.vendorName
                //vendorName  []byte
                vendorName := C.GoBytes(unsafe.Pointer(vendorName_C.buf), C.int(vendorName_C.size))
                log.Printf("Vendor Name = %x",vendorName)


        }

        /*
                indMsg, err := e2sm.GetIndicationMessage(indicationMsg.IndMessage)
                if err != nil {
                        xapp.Logger.Error("Failed to decode RIC Indication Message: %v", err)
                        return
                }
        */
        log.Printf("-----------RIC Indication Message-----------")
	log.Printf("indicationMsg.IndMessage= %x",indicationMsg.IndMessage)
        cptr2 := unsafe.Pointer(&indicationMsg.IndMessage[0])
        indicationmessage := C.e2sm_decode_ric_indication_message(cptr2, C.size_t(len(indicationMsg.IndMessage)))
        if  indicationmessage == nil {
                return errors.New("e2sm wrapper is unable to get IndicationMessage due to wrong or invalid input")
        }
        defer C.e2sm_free_ric_indication_message(indicationmessage)
        IndMsgType := int32(indicationmessage.indicationMessage_formats.present)
        if IndMsgType==1  {//parsing cell metrics
                fmt.Printf(" parsing for cell metrics\n" )
                indMsgFormat1_C := *(**C.E2SM_KPM_IndicationMessage_Format1_t)(unsafe.Pointer(&indicationmessage.indicationMessage_formats.choice[0]))
                no_of_cell:=int32(indMsgFormat1_C .measData.list.count)
		fmt.Printf(" \n No of cell = %d\n",no_of_cell )
		//fmt.Println(no_of_cell)
                for n := int32(0); n < no_of_cell; n++ {
                                var sizeof_MeasurementDataItem_t  *C.MeasurementDataItem_t
                                MeasurementDataItem_C:=*(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat1_C.measData.list.array)) + (uintptr)(int(n))*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))
                                no_of_cell_metrics:=int32(MeasurementDataItem_C.measRecord.list.count)
                                var CellM CellMetricsEntry
                                v := reflect.ValueOf(CellM)
				fmt.Printf(" \n No of cell metrics = %d\n",no_of_cell_metrics)
                                values := make(map[string]interface{}, v.NumField())
                                //assert no_of_cell_metrics == v.NumField()   they both should be equal.
                                for i := int32(0); i < no_of_cell_metrics; i++ {
					//fmt.Println(i)
                                        if v.Field(int(i)).CanInterface() {
                                                        var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
                                                        MeasurementRecordItem_C:=*(**C. MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasurementDataItem_C.measRecord.list.array)) + (uintptr)(int(i))*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))
                                                        type_var:=int(MeasurementRecordItem_C.present)
                                                        if type_var==1{
                                                                var cast_integer *C.long = (*C.long)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
                                                                values[v.Type().Field(int(i)).Name]=int32(*cast_integer)
                                                                }else if type_var==2{
                                var cast_float *C.double = (*C.double)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
                                values[v.Type().Field(int(i)).Name]=float64(*cast_float)
                                                        }else{
                                                        fmt.Printf("Wrong Data Type")
                                                }

                                                }else {
                                                fmt.Printf("sorry you have a unexported field (lower case) value you are trying to sneak past. Can not allow it: %v\n", v.Type().Field(int(i)).Name)
                                                }
                                        }//end of inner for loop


				fmt.Println(values)
				fmt.Printf("Parsing Cell Metric Done")
				//c.writeCellMetrics_db(&values)//push cellmetrics map entry to database.
                        }//end of outer for loop
                        //end of if IndMsgType==1 , parsing cell metrics done

        }  else if IndMsgType==2  { //parsing ue metrics

                fmt.Printf(" parsing for UE metrics" )
                indMsgFormat2_C := *(**C.E2SM_KPM_IndicationMessage_Format2_t)(unsafe.Pointer(&indicationmessage.indicationMessage_formats.choice[0]))
                no_of_ue_metrics:=int32(indMsgFormat2_C .measData.list.count)
		fmt.Printf(" \n No of ue metrics = %d\n",no_of_ue_metrics)

                var sizeof_MeasurementDataItem_t  *C.MeasurementDataItem_t
                MeasurementDataItem_C:=*(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(0)*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))

                no_of_ue:=int32(MeasurementDataItem_C.measRecord.list.count)
		fmt.Printf(" \n No of ue= %d\n",no_of_ue)
                for n := int32(0); n < no_of_ue; n++ {
                                var UeM UeMetricsEntry
                                v := reflect.ValueOf(UeM)
                                values := make(map[string]interface{}, v.NumField())
                                //assert no_of_ue_metrics == v.NumField()   they both should be equal.
                                for i := int32(0); i < no_of_ue_metrics; i++ {
				//fmt.Println(i)
                                if v.Field(int(i)).CanInterface() {

                                        var sizeof_MeasurementDataItem_t  *C.MeasurementDataItem_t
                                        MeasurementDataItem_C:=*(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(i)*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))
                                        var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
                                        MeasurementRecordItem_C:=*(**C.MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasurementDataItem_C.measRecord.list.array)) + (uintptr)(n)*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))

                                        type_var:=int(MeasurementRecordItem_C.present)
                                if type_var==1{
                                        var cast_integer *C.long = (*C.long)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
                                        values[v.Type().Field(int(i)).Name]=int32(*cast_integer)
                                }else if type_var==2{
                                        var cast_float *C.double = (*C.double)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
                                        values[v.Type().Field(int(i)).Name]=float64(*cast_float)

                                        }else{
                                        fmt.Printf("Wrong Data Type")
                                }

                        }else {
                                fmt.Printf("sorry you have a unexported field (lower case) value you are trying to sneak past. Can not allow it: %v\n", v.Type().Field(int(i)).Name)
                                }


                                        }       //end of inner for loop
			fmt.Println(values)
			 fmt.Printf("Parsing UE Metric Done")
			 //c.writeUeMetrics_db(&values)//push UEmetrics map entry to database.

                        }// end of outer for loop
        //parsing ue metrics done
        }else{
                fmt.Printf(" Invalid Indication message format" )

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
        cptr1 := unsafe.Pointer(&indicationMsg.IndHeader[0])
        decodedHdr := C.e2sm_decode_ric_indication_header(cptr1, C.size_t(len(indicationMsg.IndHeader)))
        //decodedHdr := C.e2sm_decode_ric_indication_header(cptr1, C.size_t(len(str)))
        //decodedHdr := C.e2sm_decode_ric_indication_header(cptr1, C.size_t(buf.Len()))
        if decodedHdr == nil {
                return errors.New("e2sm wrapper is unable to get IndicationHeader due to wrong or invalid input")
        }
        defer C.e2sm_free_ric_indication_header(decodedHdr)
        IndHdrType := int32(decodedHdr.indicationHeader_formats.present)
        if IndHdrType==0{
                log.Printf("No Indication Header present")
        }
        if IndHdrType==1{
                log.Printf("Indication Header format = %d",IndHdrType)
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
        if  indicationmessage == nil {
                return errors.New("e2sm wrapper is unable to get IndicationMessage due to wrong or invalid input")
        }
        defer C.e2sm_free_ric_indication_message(indicationmessage)
        IndMsgType := int32(indicationmessage.indicationMessage_formats.present)
        if IndMsgType==1  {//parsing cell metrics
                fmt.Printf(" parsing for cell metrics\n" )
                indMsgFormat1_C := *(**C.E2SM_KPM_IndicationMessage_Format1_t)(unsafe.Pointer(&indicationmessage.indicationMessage_formats.choice[0]))
                no_of_cell:=int32(indMsgFormat1_C .measData.list.count)
                fmt.Printf(" \n No of cell = %d\n",no_of_cell )
                //fmt.Println(no_of_cell)
                for n := int32(0); n < no_of_cell; n++ {
                                var sizeof_MeasurementDataItem_t  *C.MeasurementDataItem_t
                                MeasurementDataItem_C:=*(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat1_C.measData.list.array)) + (uintptr)(int(n))*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))
                                no_of_cell_metrics:=int32(MeasurementDataItem_C.measRecord.list.count)
                                var CellM CellMetricsEntry
                                v := reflect.ValueOf(CellM)
                                fmt.Printf(" \n No of cell metrics = %d\n",no_of_cell_metrics)
                                values := make(map[string]interface{}, v.NumField())
                                //assert no_of_cell_metrics == v.NumField()   they both should be equal.
				if (int(no_of_cell_metrics) != v.NumField()){
			 log.Printf("no_of_cell_metrics != v.NumField()")
 			return errors.New("no_of_cell_metrics != v.NumField()")
				}
                                for i := int32(0); i < no_of_cell_metrics; i++ {
                                        //fmt.Println(i)
                                        if v.Field(int(i)).CanInterface() {
                                                        var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
                                                        MeasurementRecordItem_C:=*(**C. MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasurementDataItem_C.measRecord.list.array)) + (uintptr)(int(i))*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))
                                                        type_var:=int(MeasurementRecordItem_C.present)
                                                        if type_var==1{
                                                                var cast_integer *C.long = (*C.long)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
                                                                values[v.Type().Field(int(i)).Name]=int32(*cast_integer)
                                                                }else if type_var==2{
                                var cast_float *C.double = (*C.double)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
                                values[v.Type().Field(int(i)).Name]=float64(*cast_float)
                                                        }else{
                                                        fmt.Printf("Wrong Data Type")
                                                }

                                                }else {
                                                fmt.Printf("sorry you have a unexported field (lower case) value you are trying to sneak past. Can not allow it: %v\n", v.Type().Field(int(i)).Name)
                                                }
                                        }//end of inner for loop


                                fmt.Println(values)
                                fmt.Printf("Parsing Cell Metric Done")
                                c.writeCellMetrics_db(&values)//push cellmetrics map entry to database.
                        }//end of outer for loop
                        //end of if IndMsgType==1 , parsing cell metrics done

        }  else if IndMsgType==2  { //parsing ue metrics

                fmt.Printf(" parsing for UE metrics" )
                indMsgFormat2_C := *(**C.E2SM_KPM_IndicationMessage_Format2_t)(unsafe.Pointer(&indicationmessage.indicationMessage_formats.choice[0]))
                no_of_ue_metrics:=int32(indMsgFormat2_C .measData.list.count)
                fmt.Printf(" \n No of ue metrics = %d\n",no_of_ue_metrics)

                var sizeof_MeasurementDataItem_t  *C.MeasurementDataItem_t
                MeasurementDataItem_C:=*(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(0)*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))

                no_of_ue:=int32(MeasurementDataItem_C.measRecord.list.count)
                fmt.Printf(" \n No of ue= %d\n",no_of_ue)
                for n := int32(0); n < no_of_ue; n++ {
                                var UeM UeMetricsEntry
                                v := reflect.ValueOf(UeM)
                                values := make(map[string]interface{}, v.NumField())
                                //assert no_of_ue_metrics == v.NumField()   they both should be equal.
				if (int(no_of_ue_metrics) != v.NumField()){
 			 log.Printf("no_of_ue_metrics != v.NumField()")
			 return errors.New("no_of_ue_metrics != v.NumField()")
				}
                                for i := int32(0); i < no_of_ue_metrics; i++ {
                                //fmt.Println(i)
                                if v.Field(int(i)).CanInterface() {

                                        var sizeof_MeasurementDataItem_t  *C.MeasurementDataItem_t
                                        MeasurementDataItem_C:=*(**C.MeasurementDataItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(i)*unsafe.Sizeof(sizeof_MeasurementDataItem_t)))
                                        var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
                                        MeasurementRecordItem_C:=*(**C.MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasurementDataItem_C.measRecord.list.array)) + (uintptr)(n)*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))

                                        type_var:=int(MeasurementRecordItem_C.present)
                                if type_var==1{
                                        var cast_integer *C.long = (*C.long)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
                                        values[v.Type().Field(int(i)).Name]=int32(*cast_integer)
                                }else if type_var==2{
                                        var cast_float *C.double = (*C.double)(unsafe.Pointer(&MeasurementRecordItem_C.choice[0]))
                                        values[v.Type().Field(int(i)).Name]=float64(*cast_float)

                                        }else{
                                        fmt.Printf("Wrong Data Type")
                                }

                        }else {
                                fmt.Printf("sorry you have a unexported field (lower case) value you are trying to sneak past. Can not allow it: %v\n", v.Type().Field(int(i)).Name)
                                }


                                        }       //end of inner for loop
                        fmt.Println(values)
                         fmt.Printf("Parsing UE Metric Done")
                         c.writeUeMetrics_db(&values)//push UEmetrics map entry to database.

                        }// end of outer for loop
        //parsing ue metrics done
        }else{
                fmt.Printf(" Invalid Indication message format" )

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
	ueMetricsJSON, err := json.Marshal(ueMetrics)
	if err != nil {
		xapp.Logger.Error("Marshal UE Metrics failed!")
	}
	p := influxdb2.NewPointWithMeasurement("UeMetrics").
		AddField("UE Metrics", ueMetricsJSON).
		SetTime(time.Now())
	writeAPI.WritePoint(context.Background(), p)
	xapp.Logger.Info("Wrote UE Metrics to InfluxDB")
}

func (c *Control) writeCellMetrics_db(cellMetrics *map[string]interface{}) {
	writeAPI := c.client.WriteAPIBlocking("my-org", "kpimon")
	cellMetricsJSON, er := json.Marshal(cellMetrics)
	if er != nil {
		xapp.Logger.Error("Marshal Cell Metrics failed!")
	}
	p := influxdb2.NewPointWithMeasurement("cellMetrics").
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
	fmt.Println("len of Glob_cell= ",len(Glob_cell))
	fmt.Println("Glob_cell map = ", Glob_cell)

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
