#include <errno.h>
#include "wrapper.h"

ssize_t e2sm_encode_ric_event_trigger_definition(void *buffer, size_t buf_size, size_t event_trigger_count, long *RT_periods) {
	E2SM_KPM_EventTriggerDefinition_t *eventTriggerDef = (E2SM_KPM_EventTriggerDefinition_t *)calloc(1, sizeof(E2SM_KPM_EventTriggerDefinition_t));
	if(!eventTriggerDef) {
		fprintf(stderr, "alloc EventTriggerDefinition failed\n");
		return -1;
	}

	E2SM_KPM_EventTriggerDefinition_Format1_t *innerDef = (E2SM_KPM_EventTriggerDefinition_Format1_t *)calloc(1, sizeof(E2SM_KPM_EventTriggerDefinition_Format1_t));
	if(!innerDef) {
		fprintf(stderr, "alloc EventTriggerDefinition Format1 failed\n");
		ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_EventTriggerDefinition, eventTriggerDef);
		return -1;
	}

	eventTriggerDef->present = E2SM_KPM_EventTriggerDefinition_PR_eventDefinition_Format1;
	eventTriggerDef->choice.eventDefinition_Format1 = innerDef;

	struct E2SM_KPM_EventTriggerDefinition_Format1__policyTest_List *policyTestList = (struct E2SM_KPM_EventTriggerDefinition_Format1__policyTest_List *)calloc(1, sizeof(struct E2SM_KPM_EventTriggerDefinition_Format1__policyTest_List));
	innerDef->policyTest_List = policyTestList;
	
	int index = 0;
	while(index < event_trigger_count) {
		Trigger_ConditionIE_Item_t *triggerCondition = (Trigger_ConditionIE_Item_t *)calloc(1, sizeof(Trigger_ConditionIE_Item_t));
		assert(triggerCondition != 0);
		triggerCondition->report_Period_IE = RT_periods[index];

		ASN_SEQUENCE_ADD(&policyTestList->list, triggerCondition);
		index++;
	}

	asn_enc_rval_t encode_result;
    encode_result = aper_encode_to_buffer(&asn_DEF_E2SM_KPM_EventTriggerDefinition, NULL, eventTriggerDef, buffer, buf_size);
    ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_EventTriggerDefinition, eventTriggerDef);
    if(encode_result.encoded == -1) {
        fprintf(stderr, "Cannot encode %s: %s\n", encode_result.failed_type->name, strerror(errno));
        return -1;
    } else {
	    return encode_result.encoded;
	}
}

ssize_t e2sm_encode_ric_action_definition(void *buffer, size_t buf_size, long ric_style_type) {
	E2SM_KPM_ActionDefinition_t *actionDef = (E2SM_KPM_ActionDefinition_t *)calloc(1, sizeof(E2SM_KPM_ActionDefinition_t));
	if(!actionDef) {
		fprintf(stderr, "alloc RIC ActionDefinition failed\n");
		return -1;
	}

	actionDef->ric_Style_Type = ric_style_type;

	asn_enc_rval_t encode_result;
    encode_result = aper_encode_to_buffer(&asn_DEF_E2SM_KPM_ActionDefinition, NULL, actionDef, buffer, buf_size);
    ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_ActionDefinition, actionDef);
	if(encode_result.encoded == -1) {
	    fprintf(stderr, "Cannot encode %s: %s\n", encode_result.failed_type->name, strerror(errno));
	    return -1;
	} else {
    	return encode_result.encoded;
    }
}

E2SM_KPM_IndicationHeader_t* e2sm_decode_ric_indication_header(void *buffer, size_t buf_size) {
	asn_dec_rval_t decode_result;
    E2SM_KPM_IndicationHeader_t *indHdr = 0;
    decode_result = aper_decode_complete(NULL, &asn_DEF_E2SM_KPM_IndicationHeader, (void **)&indHdr, buffer, buf_size);
    if(decode_result.code == RC_OK) {
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
    decode_result = aper_decode_complete(NULL, &asn_DEF_E2SM_KPM_IndicationMessage, (void **)&indMsg, buffer, buf_size);
    if(decode_result.code == RC_OK) {
    	return indMsg;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationMessage, indMsg);
        return NULL;
    }
}

void e2sm_free_ric_indication_message(E2SM_KPM_IndicationMessage_t* indMsg) {
	ASN_STRUCT_FREE(asn_DEF_E2SM_KPM_IndicationMessage, indMsg);
}