/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2AP-PDU-Contents"
 * 	found in "e2ap-pdu-v03.01.asn"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -findirect-choice -fno-include-deps -no-gen-example -D /tmp/workspace/tuan1/`
 */

#include "RICsubscriptionResponse.h"

asn_TYPE_member_t asn_MBR_RICsubscriptionResponse_1[] = {
	{ ATF_NOFLAGS, 0, offsetof(struct RICsubscriptionResponse, protocolIEs),
		(ASN_TAG_CLASS_CONTEXT | (0 << 2)),
		-1,	/* IMPLICIT tag at current level */
		&asn_DEF_ProtocolIE_Container_85P1,
		0,
		{
#if !defined(ASN_DISABLE_OER_SUPPORT)
			0,
#endif  /* !defined(ASN_DISABLE_OER_SUPPORT) */
#if !defined(ASN_DISABLE_UPER_SUPPORT) || !defined(ASN_DISABLE_APER_SUPPORT)
			0,
#endif  /* !defined(ASN_DISABLE_UPER_SUPPORT) || !defined(ASN_DISABLE_APER_SUPPORT) */
			0
		},
		0, 0, /* No default value */
		"protocolIEs"
		},
};
static const ber_tlv_tag_t asn_DEF_RICsubscriptionResponse_tags_1[] = {
	(ASN_TAG_CLASS_UNIVERSAL | (16 << 2))
};
static const asn_TYPE_tag2member_t asn_MAP_RICsubscriptionResponse_tag2el_1[] = {
    { (ASN_TAG_CLASS_CONTEXT | (0 << 2)), 0, 0, 0 } /* protocolIEs */
};
asn_SEQUENCE_specifics_t asn_SPC_RICsubscriptionResponse_specs_1 = {
	sizeof(struct RICsubscriptionResponse),
	offsetof(struct RICsubscriptionResponse, _asn_ctx),
	asn_MAP_RICsubscriptionResponse_tag2el_1,
	1,	/* Count of tags in the map */
	0, 0, 0,	/* Optional elements (not needed) */
	1,	/* First extension addition */
};
asn_TYPE_descriptor_t asn_DEF_RICsubscriptionResponse = {
	"RICsubscriptionResponse",
	"RICsubscriptionResponse",
	&asn_OP_SEQUENCE,
	asn_DEF_RICsubscriptionResponse_tags_1,
	sizeof(asn_DEF_RICsubscriptionResponse_tags_1)
		/sizeof(asn_DEF_RICsubscriptionResponse_tags_1[0]), /* 1 */
	asn_DEF_RICsubscriptionResponse_tags_1,	/* Same as above */
	sizeof(asn_DEF_RICsubscriptionResponse_tags_1)
		/sizeof(asn_DEF_RICsubscriptionResponse_tags_1[0]), /* 1 */
	{
#if !defined(ASN_DISABLE_OER_SUPPORT)
		0,
#endif  /* !defined(ASN_DISABLE_OER_SUPPORT) */
#if !defined(ASN_DISABLE_UPER_SUPPORT) || !defined(ASN_DISABLE_APER_SUPPORT)
		0,
#endif  /* !defined(ASN_DISABLE_UPER_SUPPORT) || !defined(ASN_DISABLE_APER_SUPPORT) */
		SEQUENCE_constraint
	},
	asn_MBR_RICsubscriptionResponse_1,
	1,	/* Elements count */
	&asn_SPC_RICsubscriptionResponse_specs_1	/* Additional specs */
};

