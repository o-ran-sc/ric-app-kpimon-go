/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2SM-COMMON-IEs"
 * 	found in "e2sm-v03.01.asn"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -findirect-choice -fno-include-deps -no-gen-example -D /tmp/workspace/tuan/`
 */

#ifndef	_RRCclass_LTE_H_
#define	_RRCclass_LTE_H_


#include "asn_application.h"

/* Including external dependencies */
#include "NativeEnumerated.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Dependencies */
typedef enum RRCclass_LTE {
	RRCclass_LTE_bCCH_BCH	= 0,
	RRCclass_LTE_bCCH_BCH_MBMS	= 1,
	RRCclass_LTE_bCCH_DL_SCH	= 2,
	RRCclass_LTE_bCCH_DL_SCH_BR	= 3,
	RRCclass_LTE_bCCH_DL_SCH_MBMS	= 4,
	RRCclass_LTE_mCCH	= 5,
	RRCclass_LTE_pCCH	= 6,
	RRCclass_LTE_dL_CCCH	= 7,
	RRCclass_LTE_dL_DCCH	= 8,
	RRCclass_LTE_uL_CCCH	= 9,
	RRCclass_LTE_uL_DCCH	= 10,
	RRCclass_LTE_sC_MCCH	= 11
	/*
	 * Enumeration is extensible
	 */
} e_RRCclass_LTE;

/* RRCclass-LTE */
typedef long	 RRCclass_LTE_t;

/* Implementation */
extern asn_per_constraints_t asn_PER_type_RRCclass_LTE_constr_1;
extern asn_TYPE_descriptor_t asn_DEF_RRCclass_LTE;
extern const asn_INTEGER_specifics_t asn_SPC_RRCclass_LTE_specs_1;
asn_struct_free_f RRCclass_LTE_free;
asn_struct_print_f RRCclass_LTE_print;
asn_constr_check_f RRCclass_LTE_constraint;
ber_type_decoder_f RRCclass_LTE_decode_ber;
der_type_encoder_f RRCclass_LTE_encode_der;
xer_type_decoder_f RRCclass_LTE_decode_xer;
xer_type_encoder_f RRCclass_LTE_encode_xer;
jer_type_encoder_f RRCclass_LTE_encode_jer;
oer_type_decoder_f RRCclass_LTE_decode_oer;
oer_type_encoder_f RRCclass_LTE_encode_oer;
per_type_decoder_f RRCclass_LTE_decode_uper;
per_type_encoder_f RRCclass_LTE_encode_uper;
per_type_decoder_f RRCclass_LTE_decode_aper;
per_type_encoder_f RRCclass_LTE_encode_aper;

#ifdef __cplusplus
}
#endif

#endif	/* _RRCclass_LTE_H_ */
#include "asn_internal.h"
