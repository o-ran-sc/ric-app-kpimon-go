#ifndef	_WRAPPER_H_
#define	_WRAPPER_H_

#include "E2SM-KPM-EventTriggerDefinition.h"
#include "E2SM-KPM-EventTriggerDefinition-Format1.h"
#include "Trigger-ConditionIE-Item.h"
#include "E2SM-KPM-ActionDefinition.h"
#include "E2SM-KPM-IndicationHeader.h"
#include "E2SM-KPM-IndicationHeader-Format1.h"
#include "GlobalKPMnode-ID.h"
#include "GlobalKPMnode-gNB-ID.h"
#include "GlobalKPMnode-en-gNB-ID.h"
#include "GlobalKPMnode-ng-eNB-ID.h"
#include "GlobalKPMnode-eNB-ID.h"
#include "PLMN-Identity.h"
#include "GNB-ID-Choice.h"
#include "GNB-CU-UP-ID.h"
#include "GNB-DU-ID.h"
#include "ENGNB-ID.h"
#include "ENB-ID-Choice.h"
#include "ENB-ID.h"
#include "NRCGI.h"
#include "SNSSAI.h"
#include "GNB-Name.h"
#include "E2SM-KPM-IndicationMessage.h"
#include "E2SM-KPM-IndicationMessage-Format1.h"
#include "PM-Containers-List.h"
#include "PF-Container.h"
#include "RAN-Container.h"
#include "ODU-PF-Container.h"
#include "CellResourceReportListItem.h"
#include "ServedPlmnPerCellListItem.h"
#include "FGC-DU-PM-Container.h"
#include "EPC-DU-PM-Container.h"
#include "SlicePerPlmnPerCellListItem.h"
#include "FQIPERSlicesPerPlmnPerCellListItem.h"
#include "PerQCIReportListItem.h"
#include "OCUCP-PF-Container.h"
#include "OCUUP-PF-Container.h"
#include "PF-ContainerListItem.h"
#include "PlmnID-List.h"
#include "FGC-CUUP-PM-Format.h"
#include "SliceToReportListItem.h"
#include "FQIPERSlicesPerPlmnListItem.h"
#include "EPC-CUUP-PM-Format.h"
#include "PerQCIReportListItemFormat.h"
#include "DU-Usage-Report-Per-UE.h"
#include "DU-Usage-Report-CellResourceReportItem.h"
#include "DU-Usage-Report-UeResourceReportItem.h"
#include "CU-CP-Usage-Report-Per-UE.h"
#include "CU-CP-Usage-Report-CellResourceReportItem.h"
#include "CU-CP-Usage-Report-UeResourceReportItem.h"
#include "CU-UP-Usage-Report-Per-UE.h"
#include "CU-UP-Usage-Report-CellResourceReportItem.h"
#include "CU-UP-Usage-Report-UeResourceReportItem.h"

ssize_t e2sm_encode_ric_event_trigger_definition(void *buffer, size_t buf_size, size_t event_trigger_count, long *RT_periods);
ssize_t e2sm_encode_ric_action_definition(void *buffer, size_t buf_size, long ric_style_type);
E2SM_KPM_IndicationHeader_t* e2sm_decode_ric_indication_header(void *buffer, size_t buf_size);
void e2sm_free_ric_indication_header(E2SM_KPM_IndicationHeader_t* indHdr);
E2SM_KPM_IndicationMessage_t* e2sm_decode_ric_indication_message(void *buffer, size_t buf_size);
void e2sm_free_ric_indication_message(E2SM_KPM_IndicationMessage_t* indMsg);

#endif /* _WRAPPER_H_ */
