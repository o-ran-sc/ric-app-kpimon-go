/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2AP-IEs"
 * 	found in "e2ap-ied-v03.01.asn"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -findirect-choice -fno-include-deps -no-gen-example -D /tmp/workspace/tuan1/`
 */

#ifndef	_E2nodeComponentInterfaceF1_H_
#define	_E2nodeComponentInterfaceF1_H_


#include "asn_application.h"

/* Including external dependencies */
#include "GNB-DU-ID.h"
#include "constr_SEQUENCE.h"

#ifdef __cplusplus
extern "C" {
#endif

/* E2nodeComponentInterfaceF1 */
typedef struct E2nodeComponentInterfaceF1 {
	GNB_DU_ID_t	 gNB_DU_ID;
	/*
	 * This type is extensible,
	 * possible extensions are below.
	 */
	
	/* Context for parsing across buffer boundaries */
	asn_struct_ctx_t _asn_ctx;
} E2nodeComponentInterfaceF1_t;

/* Implementation */
extern asn_TYPE_descriptor_t asn_DEF_E2nodeComponentInterfaceF1;
extern asn_SEQUENCE_specifics_t asn_SPC_E2nodeComponentInterfaceF1_specs_1;
extern asn_TYPE_member_t asn_MBR_E2nodeComponentInterfaceF1_1[1];

#ifdef __cplusplus
}
#endif

#endif	/* _E2nodeComponentInterfaceF1_H_ */
#include "asn_internal.h"
