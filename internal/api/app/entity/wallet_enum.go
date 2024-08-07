// Code generated by "enumer -type=WalletEventType,TransferStatus -trimprefix=EventType,TransferStatus -transform=snake -output=wallet_enum.go -json -sql -text"; DO NOT EDIT.

package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

const _WalletEventTypeName = "invaliddebit_transfercredit_transferupdate_transfer_status"

var _WalletEventTypeIndex = [...]uint8{0, 7, 21, 36, 58}

const _WalletEventTypeLowerName = "invaliddebit_transfercredit_transferupdate_transfer_status"

func (i WalletEventType) String() string {
	if i >= WalletEventType(len(_WalletEventTypeIndex)-1) {
		return fmt.Sprintf("WalletEventType(%d)", i)
	}
	return _WalletEventTypeName[_WalletEventTypeIndex[i]:_WalletEventTypeIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _WalletEventTypeNoOp() {
	var x [1]struct{}
	_ = x[EventTypeInvalid-(0)]
	_ = x[EventTypeDebitTransfer-(1)]
	_ = x[EventTypeCreditTransfer-(2)]
	_ = x[EventTypeUpdateTransferStatus-(3)]
}

var _WalletEventTypeValues = []WalletEventType{EventTypeInvalid, EventTypeDebitTransfer, EventTypeCreditTransfer, EventTypeUpdateTransferStatus}

var _WalletEventTypeNameToValueMap = map[string]WalletEventType{
	_WalletEventTypeName[0:7]:        EventTypeInvalid,
	_WalletEventTypeLowerName[0:7]:   EventTypeInvalid,
	_WalletEventTypeName[7:21]:       EventTypeDebitTransfer,
	_WalletEventTypeLowerName[7:21]:  EventTypeDebitTransfer,
	_WalletEventTypeName[21:36]:      EventTypeCreditTransfer,
	_WalletEventTypeLowerName[21:36]: EventTypeCreditTransfer,
	_WalletEventTypeName[36:58]:      EventTypeUpdateTransferStatus,
	_WalletEventTypeLowerName[36:58]: EventTypeUpdateTransferStatus,
}

var _WalletEventTypeNames = []string{
	_WalletEventTypeName[0:7],
	_WalletEventTypeName[7:21],
	_WalletEventTypeName[21:36],
	_WalletEventTypeName[36:58],
}

// WalletEventTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func WalletEventTypeString(s string) (WalletEventType, error) {
	if val, ok := _WalletEventTypeNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _WalletEventTypeNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to WalletEventType values", s)
}

// WalletEventTypeValues returns all values of the enum
func WalletEventTypeValues() []WalletEventType {
	return _WalletEventTypeValues
}

// WalletEventTypeStrings returns a slice of all String values of the enum
func WalletEventTypeStrings() []string {
	strs := make([]string, len(_WalletEventTypeNames))
	copy(strs, _WalletEventTypeNames)
	return strs
}

// IsAWalletEventType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i WalletEventType) IsAWalletEventType() bool {
	for _, v := range _WalletEventTypeValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for WalletEventType
func (i WalletEventType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for WalletEventType
func (i *WalletEventType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("WalletEventType should be a string, got %s", data)
	}

	var err error
	*i, err = WalletEventTypeString(s)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface for WalletEventType
func (i WalletEventType) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for WalletEventType
func (i *WalletEventType) UnmarshalText(text []byte) error {
	var err error
	*i, err = WalletEventTypeString(string(text))
	return err
}

func (i WalletEventType) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *WalletEventType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	case fmt.Stringer:
		str = v.String()
	default:
		return fmt.Errorf("invalid value of WalletEventType: %[1]T(%[1]v)", value)
	}

	val, err := WalletEventTypeString(str)
	if err != nil {
		return err
	}

	*i = val
	return nil
}

const _TransferStatusName = "invalidpendingcompletedfailed"

var _TransferStatusIndex = [...]uint8{0, 7, 14, 23, 29}

const _TransferStatusLowerName = "invalidpendingcompletedfailed"

func (i TransferStatus) String() string {
	if i >= TransferStatus(len(_TransferStatusIndex)-1) {
		return fmt.Sprintf("TransferStatus(%d)", i)
	}
	return _TransferStatusName[_TransferStatusIndex[i]:_TransferStatusIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _TransferStatusNoOp() {
	var x [1]struct{}
	_ = x[TransferStatusInvalid-(0)]
	_ = x[TransferStatusPending-(1)]
	_ = x[TransferStatusCompleted-(2)]
	_ = x[TransferStatusFailed-(3)]
}

var _TransferStatusValues = []TransferStatus{TransferStatusInvalid, TransferStatusPending, TransferStatusCompleted, TransferStatusFailed}

var _TransferStatusNameToValueMap = map[string]TransferStatus{
	_TransferStatusName[0:7]:        TransferStatusInvalid,
	_TransferStatusLowerName[0:7]:   TransferStatusInvalid,
	_TransferStatusName[7:14]:       TransferStatusPending,
	_TransferStatusLowerName[7:14]:  TransferStatusPending,
	_TransferStatusName[14:23]:      TransferStatusCompleted,
	_TransferStatusLowerName[14:23]: TransferStatusCompleted,
	_TransferStatusName[23:29]:      TransferStatusFailed,
	_TransferStatusLowerName[23:29]: TransferStatusFailed,
}

var _TransferStatusNames = []string{
	_TransferStatusName[0:7],
	_TransferStatusName[7:14],
	_TransferStatusName[14:23],
	_TransferStatusName[23:29],
}

// TransferStatusString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func TransferStatusString(s string) (TransferStatus, error) {
	if val, ok := _TransferStatusNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _TransferStatusNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to TransferStatus values", s)
}

// TransferStatusValues returns all values of the enum
func TransferStatusValues() []TransferStatus {
	return _TransferStatusValues
}

// TransferStatusStrings returns a slice of all String values of the enum
func TransferStatusStrings() []string {
	strs := make([]string, len(_TransferStatusNames))
	copy(strs, _TransferStatusNames)
	return strs
}

// IsATransferStatus returns "true" if the value is listed in the enum definition. "false" otherwise
func (i TransferStatus) IsATransferStatus() bool {
	for _, v := range _TransferStatusValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for TransferStatus
func (i TransferStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for TransferStatus
func (i *TransferStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("TransferStatus should be a string, got %s", data)
	}

	var err error
	*i, err = TransferStatusString(s)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface for TransferStatus
func (i TransferStatus) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for TransferStatus
func (i *TransferStatus) UnmarshalText(text []byte) error {
	var err error
	*i, err = TransferStatusString(string(text))
	return err
}

func (i TransferStatus) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *TransferStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	case fmt.Stringer:
		str = v.String()
	default:
		return fmt.Errorf("invalid value of TransferStatus: %[1]T(%[1]v)", value)
	}

	val, err := TransferStatusString(str)
	if err != nil {
		return err
	}

	*i = val
	return nil
}
