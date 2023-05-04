package eventsocket

//All constants corresponding to internal Freeswitch variables

/*
 * Variables
 */

const (
	AnswerState = iota + 1
	Application
	ApplicationData
	ApplicationResponse
	ApplicationUUID
	CallDirection
	CallerAni
	CallerOrigCallerIDNumber
	CallerDestinationNumber
	CallerChannelAnsweredTime
	CallerChannelBridgedTime
	CallerChannelCreatedTime
	CallerChannelHangupTime
	CallerNetworkAddr
	CallerChannelProgressTime
	CallerUniqueID
	ChannelCallState
	ChannelCallUUID
	ChannelName
	ChannelState
	ChannelStateNumber
	CoreUUID
	DtmfDigit
	CustomHeaders
	UserToUser
	EventName
	EventDateGmt
	EventDateTimeStamp
	EventSourceIpv4
	EventMsgBody
	EventSubClass
	HangupCause
	OtherLegUniqueID
	OtherLegDestinationNumber
	OtherLegChannelAnsweredTime
	OtherLegChannelName
	RecordFilePath
	UniqueID
	VariableCurrentApplication
	VariableSofiaProfileName
	VariableDomainName
	VariableSipCallID
	VariableSipFullFrom
	VariableSipFullTo
	VariableDetectSpeechResult
	VariableRecordSeconds
	VariableRecordStereo
	VariableSipInviteFailureStatus
	VariableDuration
	VariableSipHXInfo
	VariableDTMFResult
	VariableDTMFResultInvalid
	VariableBillSec
	VariablePlaybackTerminatorUsed
	VariableParentVerb
	VariableParentID
	VariableQueueID
	VariableOriginateSignalBond
	InstanceHashID

	// Conference
	Action
	ConferenceName
	ConferenceSize
	ConferenceUniqueId
	Floor
	Hear
	Hold
	MemberId
	MemberType
	MuteDetect
	Talking
	Speak
	ConfRecPath
	MillisecondsElapsed
)

var MapKeyIndex = map[string]int{
	"Answer-State":                       AnswerState,
	"Application":                        Application,
	"Application-Data":                   ApplicationData,
	"Application-Response":               ApplicationResponse,
	"Application-Uuid":                   ApplicationUUID,
	"Call-Direction":                     CallDirection,
	"Caller-Ani":                         CallerAni,
	"Caller-Orig-Caller-Id-Number":       CallerOrigCallerIDNumber,
	"Caller-Destination-Number":          CallerDestinationNumber,
	"Caller-Channel-Answered-Time":       CallerChannelAnsweredTime,
	"Caller-Channel-Bridged-Time":        CallerChannelBridgedTime,
	"Caller-Channel-Created-Time":        CallerChannelCreatedTime,
	"Caller-Channel-Hangup-Time":         CallerChannelHangupTime,
	"Caller-Network-Addr":                CallerNetworkAddr,
	"Caller-Channel-Progress-Time":       CallerChannelProgressTime,
	"Caller-Unique-Id":                   CallerUniqueID,
	"Channel-Call-State":                 ChannelCallState,
	"Channel-Call-Uuid":                  ChannelCallUUID,
	"Channel-Name":                       ChannelName,
	"Channel-State":                      ChannelState,
	"Channel-State-Number":               ChannelStateNumber,
	"Core-Uuid":                          CoreUUID,
	"Dtmf-Digit":                         DtmfDigit,
	"Event-Name":                         EventName,
	"Event-Date-Gmt":                     EventDateGmt,
	"Event-Date-Timestamp":               EventDateTimeStamp,
	"Freeswitch-Ipv4":                    EventSourceIpv4,
	"Event-Msg-Body":                     EventMsgBody,
	"Event-Subclass":                     EventSubClass,
	"Hangup-Cause":                       HangupCause,
	"Other-Leg-Unique-Id":                OtherLegUniqueID,
	"Other-Leg-Destination-Number":       OtherLegDestinationNumber,
	"Other-Leg-Channel-Answered-Time":    OtherLegChannelAnsweredTime,
	"Other-Leg-Channel-Name":             OtherLegChannelName,
	"Record-File-Path":                   RecordFilePath,
	"Unique-Id":                          UniqueID,
	"Variable_current_application":       VariableCurrentApplication,
	"Variable_sofia_profile_name":        VariableSofiaProfileName,
	"Variable_domain_name":               VariableDomainName,
	"Variable_sip_call_id":               VariableSipCallID,
	"Variable_sip_full_from":             VariableSipFullFrom,
	"Variable_sip_full_to":               VariableSipFullTo,
	"Variable_detect_speech_result":      VariableDetectSpeechResult,
	"Variable_record_seconds":            VariableRecordSeconds,
	"Variable_record_stereo":             VariableRecordStereo,
	"Variable_sip_invite_failure_status": VariableSipInviteFailureStatus,
	"Variable_duration":                  VariableDuration,
	"Variable_dtmfresultvar":             VariableDTMFResult,
	"Variable_dtmfresultvar_invalid":     VariableDTMFResultInvalid,
	"Variable_billsec":                   VariableBillSec,
	"Variable_playback_terminator_used":  VariablePlaybackTerminatorUsed,
	"Variable_parent_verb":               VariableParentVerb,
	"Variable_parent_id":                 VariableParentID,
	"Variable_queue_id":                  VariableQueueID,
	"Custom-Headers":                     CustomHeaders,
	"Variable_sip_h_user-to-user":        UserToUser,
	"Variable_originate_signal_bond":     VariableOriginateSignalBond,
	"VariableInstanceHash":               InstanceHashID,

	// Conference
	"Action":               Action,
	"Conference-Name":      ConferenceName,
	"Conference-Size":      ConferenceSize,
	"Conference-Unique-Id": ConferenceUniqueId,
	"Floor":                Floor,
	"Hear":                 Hear,
	"Hold":                 Hold,
	"Member-Id":            MemberId,
	"MemberType":           MemberType,
	"MuteDetect":           MuteDetect,
	"Talking":              Talking,
	"Speak":                Speak,
	"Path":                 ConfRecPath,
	"Milliseconds-Elapsed": MillisecondsElapsed,
}

const FsEventMapSize = 100
