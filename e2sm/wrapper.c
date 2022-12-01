#include <errno.h>
#include "wrapper.h"
#include <stdio.h>


// int main() {

// 	long id[] = {1,2,3,4};

// 	void *buffer[8] = {} ;
// 	unsigned int sizeBuffer = 8;
// 	long actionDefCount = 8;
// 	long measId = 1;
// 	long ricStyleType = 1;
// 	long periods = 1;
// 	void *plm[8] = {} ;
// 	unsigned int sizePlm = 8;
// 	void *nRc[8] = {} ;
// 	unsigned int sizenRc = 8;

// 	int a = e2sm_encode_ric_action_definition(buffer, sizeBuffer, actionDefCount, measId, ricStyleType, periods, plm, nRc);

// 	printf("%d", a);

// 	return 0;
// }

ssize_t e2sm_encode_ric_event_trigger_definition(void *buffer, size_t buf_size, size_t event_trigger_count, long *RT_periods) {
	E2SM_KPM_EventTriggerDefinition_t *eventTriggerDef = (E2SM_KPM_EventTriggerDefinition_t *)calloc(1, sizeof(E2SM_KPM_EventTriggerDefinition_t));
	if(!eventTriggerDef) {
		fprintf(stderr, "Alloc EventTriggerDefinition failed\n");
		return -1;
	}

	E2SM_KPM_EventTriggerDefinition_Format1_t *innerDef = (E2SM_KPM_EventTriggerDefinition_Format1_t *)calloc(1, sizeof(E2SM_KPM_EventTriggerDefinition_Format1_t));
	if(!innerDef) {
		fprintf(stderr, "alloc EventTriggerDefinition Format1 failed\n");
		ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_EventTriggerDefinition, eventTriggerDef);
		return -1;
	}

	eventTriggerDef->eventDefinition_formats.present = E2SM_KPM_EventTriggerDefinition__eventDefinition_formats_PR_eventDefinition_Format1;
	eventTriggerDef->eventDefinition_formats.choice.eventDefinition_Format1 = innerDef;

	// struct E2SM_KPM_EventTriggerDefinition_Format1__policyTest_List *policyTestList = (struct E2SM_KPM_EventTriggerDefinition_Format1__policyTest_List *)calloc(1, sizeof(struct E2SM_KPM_EventTriggerDefinition_Format1__policyTest_List));
	// innerDef->policyTest_List = policyTestList;
	
	// int index = 0;
	// while(index < event_trigger_count) {
	// 	Trigger_ConditionIE_Item *triggerCondition = (Trigger_ConditionIE_Item *)calloc(1, sizeof(Trigger_ConditionIE_Item));
	// 	assert(triggerCondition != 0);
	// 	triggerCondition->report_Period_IE = RT_periods[index];
	// 	ASN_SEQUENCE_ADD(&policyTestList->list, triggerCondition);
	// 	index++;
	// }

	asn_enc_rval_t encode_result;
    encode_result = uper_encode_to_buffer(&asn_DEF_E2SM_KPM_EventTriggerDefinition, NULL, eventTriggerDef, buffer, buf_size);
    ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_EventTriggerDefinition, eventTriggerDef);
    if(encode_result.encoded == -1) {
        fprintf(stderr, "Cannot encode %s: %s\n", encode_result.failed_type->name, strerror(errno));
        return -1;
    } else {
	    return encode_result.encoded;
	}
}
// int main() {

// 	long id[] = {1,2,3,4};

// 	void *buffer[8] = {} ;
// 	unsigned int sizeBuffer = 8;
// 	long actionDefCount = 8;
// 	long measId = 1;
// 	long ricStyleType = 1;
// 	long periods = 1;
// 	void *plm[8] = {} ;
// 	unsigned int sizePlm = 8;
// 	void *nRc[8] = {} ;
// 	unsigned int sizenRc = 8;

// 	int a = e2sm_encode_ric_action_definition(buffer, sizeBuffer, actionDefCount, measId, ricStyleType, periods, plm, nRc);

// 	printf("%d", a);

// 	return 0;
// }

/*
ssize_t e2sm_encode_ric_action_definition(void *buffer, size_t buf_size, long action_Def_count, long meas_ID, long ric_style_type, long GR_periods, Plm plm, Nrc nRc) {
	E2SM_KPM_ActionDefinition_t *actionDef = (E2SM_KPM_ActionDefinition_t *)calloc(1, sizeof(E2SM_KPM_ActionDefinition_t));
	if(!actionDef) {
		fprintf(stderr, "alloc RIC ActionDefinition failed\n");
		return -1;
	}

	E2SM_KPM_ActionDefinition_Format1_t *actionDefFor1 = (E2SM_KPM_ActionDefinition_Format1_t *)calloc(1, sizeof(E2SM_KPM_ActionDefinition_Format1_t));
	if(!actionDefFor1) {
		ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_ActionDefinition, actionDef);
		fprintf(stderr, "alloc RIC ActionDefinition failed\n");
		return -1;
	}
	CGI_t *cellGlobal = (CGI_t *)calloc(1, sizeof(CGI_t));
	if(!cellGlobal) {
		fprintf(stderr, "alloc RIC ActionDefinition failed\n");
		return -1;
	}

	NR_CGI_t *nrCGIs = (NR_CGI_t *)calloc(1, sizeof(NR_CGI_t));
	if(!nrCGIs) {
		fprintf(stderr, "alloc RIC ActionDefinition failed\n");
		return -1;
	}

	actionDef->ric_Style_Type = ric_style_type;
	actionDef->actionDefinition_formats.present = E2SM_KPM_ActionDefinition__actionDefinition_formats_PR_actionDefinition_Format1;
	actionDef->actionDefinition_formats.choice.actionDefinition_Format1 = actionDefFor1;
	actionDefFor1->granulPeriod = GR_periods;
	actionDefFor1->cellGlobalID = cellGlobal;
	cellGlobal->present = CGI_PR_nR_CGI;
	cellGlobal->choice.nR_CGI = nrCGIs;


	int plmSize = plm.size;
	if(plmSize != 0) {

	}

    PLMNIdentity_t *plms = &nrCGIs->pLMNIdentity;
    plms->buf = (uint8_t *)calloc(1, plmSize);
    if(!plms->buf) {
        fprintf(stderr, "alloc plms failed\n");
	    ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_ActionDefinition, actionDef);
        return -1;
    }

    memcpy(plms->buf, plm.plms, plmSize);
    plms->size = plmSize;
	

	int nRcSize = nRc.size;

	NRCellIdentity_t *nRCs = &nrCGIs->nRCellIdentity;
    nRCs->buf = (uint8_t *)calloc(1, nRcSize);
    if(!nRCs->buf) {
        fprintf(stderr, "alloc nRCs failed\n");
	    ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_ActionDefinition, actionDef);
        return -1;
    }

    memcpy(nRCs->buf, nRc.nRcs, nRcSize);
    nRCs->size = nRcSize; 


	MeasurementInfoList_t *infoList = &actionDefFor1->measInfoList;
	
	int index = 0;
	while(index < action_Def_count) {
		MeasurementInfoItem_t *infoItem = (MeasurementInfoItem_t *)calloc(1, sizeof(MeasurementInfoItem_t));
		// if(!infoItem) {
		// 	fprintf(stderr, "alloc Measurement Info Item failed\n");
		//     ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_ActionDefinition, actionDef);
		// return -1;
		// }
		assert(infoItem != 0);

	
		MeasurementType_t *measureType = &infoItem->measType;

		measureType->present = MeasurementType_PR_measID;
		measureType->choice.measID = meas_ID;


		LabelInfoList_t *lableList = &infoItem->labelInfoList;

		ASN_SEQUENCE_ADD(&infoList->list, infoItem);
        index++;
	}

	asn_enc_rval_t encode_result;
    encode_result = uper_encode_to_buffer(&asn_DEF_E2SM_KPM_ActionDefinition, NULL,actionDef, buffer, buf_size);
    ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_ActionDefinition, actionDef);

	if(encode_result.encoded == -1) {
	    fprintf(stderr, "Cannot encode %s: %s\n", encode_result.failed_type->name, strerror(errno));
	    return -1;
	} else {
		fprintf(stderr, "successfully\n");
    	return encode_result.encoded;
    }
}
*/
E2SM_KPM_IndicationHeader_t* e2sm_decode_ric_indication_header(void *buffer, size_t buf_size) {
	asn_dec_rval_t decode_result;
    E2SM_KPM_IndicationHeader_t *indHdr = 0;
     //fprintf(stderr, "\nbuffer= %p",buffer);
	//fprintf(stdout, "\nbuf_size=%ld",buf_size);
	//fprintf(stdout, "\nE2SM_KPM_IndicationHeader_t_size=%ld",sizeof(E2SM_KPM_IndicationHeader_t));
   //decode_result = uper_decode_complete(NULL, &asn_DEF_E2SM_KPM_IndicationHeader, (void **)&indHdr, (char*)buffer, buf_size);
   //ATS_ALIGNED_CANONICAL_PER
    decode_result =asn_decode(0,ATS_ALIGNED_BASIC_PER,&asn_DEF_E2SM_KPM_IndicationHeader,(void **)&indHdr,buffer,buf_size);
    if(decode_result.code == RC_OK) {
	// xer_fprint(stderr, &asn_DEF_E2SM_KPM_IndicationHeader, indHdr);
        return indHdr;
    }
     else if (decode_result.code ==RC_WMORE ) {
        //xer_fprint(stderr, &asn_DEF_E2SM_KPM_IndicationHeader, indHdr);
         //fprintf(stderr, "\n decode_result.consumed= %ld \n",decode_result.consumed);
        fprintf(stderr, "\nheader RC_WMORE ");
	return indHdr;
        }

    else {
        ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationHeader, indHdr);
        return NULL;
    }
}

void e2sm_free_ric_indication_header(E2SM_KPM_IndicationHeader_t* indHdr) {
	ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationHeader, indHdr);
}

E2SM_KPM_IndicationMessage_t* e2sm_decode_ric_indication_message(void *buffer, size_t buf_size) {
	asn_dec_rval_t decode_result;
     E2SM_KPM_IndicationMessage_t *indMsg = 0;
     //fprintf(stdout, "\nbuffer= %p",buffer);
     //fprintf(stdout, "\nbuf_size=%ld",buf_size);
     //fprintf(stdout, "\nE2SM_KPM_IndicationMessage_t size=%ld",sizeof(E2SM_KPM_IndicationMessage_t));
    //decode_result = uper_decode_complete(NULL, &asn_DEF_E2SM_KPM_IndicationMessage, (void **)&indMsg, (char *)buffer, buf_size);
    //ATS_ALIGNED_CANONICAL_PER
   decode_result =asn_decode(0,ATS_ALIGNED_BASIC_PER,&asn_DEF_E2SM_KPM_IndicationMessage,(void **)&indMsg,buffer,buf_size);
    if(decode_result.code == RC_OK) {
	    //xer_fprint(stdout, &asn_DEF_E2SM_KPM_IndicationMessage, indMsg);
    	return indMsg;
    }
     else if (decode_result.code ==RC_WMORE ) {
         //xer_fprint(stderr, &asn_DEF_E2SM_KPM_IndicationMessage, indMsg);
         //fprintf(stderr, "\n MSG decode_result.consumed= %ld \n",decode_result.consumed);
             fprintf(stderr, "\n MSG RC_WMORE ");
	     return NULL;
     }

    else {
        ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationMessage, indMsg);
        return NULL;
    }
}

void e2sm_free_ric_indication_message(E2SM_KPM_IndicationMessage_t* indMsg) {
	ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationMessage, indMsg);
}

