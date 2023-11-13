/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2AP-PDU-Contents"
 * 	found in "e2ap-pdu-v03.01.asn"
 * 	`asn1c -pdu=auto -fincludes-quoted -fcompound-names -findirect-choice -fno-include-deps -no-gen-example -D /tmp/workspace/tuan1/`
 */

#include "RICaction-RequiredToBeModified-Item.h"

asn_TYPE_member_t asn_MBR_RICaction_RequiredToBeModified_Item_1[] = {
	{ ATF_NOFLAGS, 0, offsetof(struct RICaction_RequiredToBeModified_Item, ricActionID),
		(ASN_TAG_CLASS_CONTEXT | (0 << 2)),
		-1,	/* IMPLICIT tag at current level */
		&asn_DEF_RICactionID,
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
		"ricActionID"
		},
	{ ATF_NOFLAGS, 0, offsetof(struct RICaction_RequiredToBeModified_Item, ricTimeToWait),
		(ASN_TAG_CLASS_CONTEXT | (1 << 2)),
		-1,	/* IMPLICIT tag at current level */
		&asn_DEF_RICtimeToWait,
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
		"ricTimeToWait"
		},
};
static const ber_tlv_tag_t asn_DEF_RICaction_RequiredToBeModified_Item_tags_1[] = {
	(ASN_TAG_CLASS_UNIVERSAL | (16 << 2))
};
static const asn_TYPE_tag2member_t asn_MAP_RICaction_RequiredToBeModified_Item_tag2el_1[] = {
    { (ASN_TAG_CLASS_CONTEXT | (0 << 2)), 0, 0, 0 }, /* ricActionID */
    { (ASN_TAG_CLASS_CONTEXT | (1 << 2)), 1, 0, 0 } /* ricTimeToWait */
};
asn_SEQUENCE_specifics_t asn_SPC_RICaction_RequiredToBeModified_Item_specs_1 = {
	sizeof(struct RICaction_RequiredToBeModified_Item),
	offsetof(struct RICaction_RequiredToBeModified_Item, _asn_ctx),
	asn_MAP_RICaction_RequiredToBeModified_Item_tag2el_1,
	2,	/* Count of tags in the map */
	0, 0, 0,	/* Optional elements (not needed) */
	2,	/* First extension addition */
};
asn_TYPE_descriptor_t asn_DEF_RICaction_RequiredToBeModified_Item = {
	"RICaction-RequiredToBeModified-Item",
	"RICaction-RequiredToBeModified-Item",
	&asn_OP_SEQUENCE,
	asn_DEF_RICaction_RequiredToBeModified_Item_tags_1,
	sizeof(asn_DEF_RICaction_RequiredToBeModified_Item_tags_1)
		/sizeof(asn_DEF_RICaction_RequiredToBeModified_Item_tags_1[0]), /* 1 */
	asn_DEF_RICaction_RequiredToBeModified_Item_tags_1,	/* Same as above */
	sizeof(asn_DEF_RICaction_RequiredToBeModified_Item_tags_1)
		/sizeof(asn_DEF_RICaction_RequiredToBeModified_Item_tags_1[0]), /* 1 */
	{
#if !defined(ASN_DISABLE_OER_SUPPORT)
		0,
#endif  /* !defined(ASN_DISABLE_OER_SUPPORT) */
#if !defined(ASN_DISABLE_UPER_SUPPORT) || !defined(ASN_DISABLE_APER_SUPPORT)
		0,
#endif  /* !defined(ASN_DISABLE_UPER_SUPPORT) || !defined(ASN_DISABLE_APER_SUPPORT) */
		SEQUENCE_constraint
	},
	asn_MBR_RICaction_RequiredToBeModified_Item_1,
	2,	/* Elements count */
	&asn_SPC_RICaction_RequiredToBeModified_Item_specs_1	/* Additional specs */
};

