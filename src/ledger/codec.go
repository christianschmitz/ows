package ledger

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/bech32"
	"github.com/fxamacker/cbor/v2"
)

// `encodeableChangeSet` is an intermediate representation of ChangeSet.
//
// The Prev ChangeSetID has been converted into its underlying hash bytes, and
// the actions have been converted into a CBOR encodeable format.
//
// Empty signatures are omitted to create the change set body that is used for
// signing.
type encodeableChangeSet struct {
	Prev       []byte             `cbor:"0,keyasint"`
	Actions    []encodeableAction `cbor:"1,keyasint"`
	Signatures []Signature        `cbor:"2,keyasint,omitempty"`
}

// `encodeableAction` is an intermediate representation of Action, with
// encoding of the attributes already applied.
//
// If this format changes from one ledger version to another, change name to
// `encodeableActionV1`, and add `encodeableActionV2` etc, and add appropriate
// switch logic to `decodeAction()` and `actionEncoder.encodeable()`.
type encodeableAction struct {
	Category   string `cbor:"0,keyasint"`
	Name       string `cbor:"1,keyasint"`
	Attributes []byte `cbor:"2,keyasint"`
}

// A function that unmarshals CBOR bytes into a specific Action type.
type actionDecoder = func(attrBytes []byte) (Action, error)

// Decoding and validates a ledger.
//
// Validation must be done at the same time because decoding depends on the ledger version,
// which can change from one change set to another.
func DecodeLedger(bs []byte) (*Ledger, error) {
	var entries [][]byte

	err := cbor.Unmarshal(bs, &entries)
	if err != nil {
		return nil, fmt.Errorf("invalid top-level ledger format, expected list of bytestrings (%v)", err)
	}

	if len(entries) < 2 {
		return nil, fmt.Errorf("invalid top-level ledger format, bytestring list contains less than two entries (expected ledger version and initial configuration, got %d entry) (%v)", len(entries), err)
	}

	initialVersion, err := decodeLedgerVersion(entries[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode initial ledger version (%v)", err)
	}

	initialConfig, err := DecodeChangeSet(entries[1], initialVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to decode first ledger change set (%v)", err)
	}

	snapshot := newSnapshot(initialVersion)

	// Use a generator for the remaining change sets, so the validation
	// function can be used both here and to revalidate the ledger from the
	// initial config.
	genChanges := func(yield func(cs *ChangeSet, err error) bool) {
		i := 2

		for {
			if i < len(entries) {
				cs, err := DecodeChangeSet(entries[i], snapshot.Version)

				i++

				if err != nil {
					err = fmt.Errorf("failed to decode ledger change set %d (prev=%s) (%v)", i, snapshot.Head, err)
				}

				if !yield(cs, err) {
					return
				}
			} else {
				return
			}
		}
	}

	changes, err := validateAllChangeSets(initialConfig, genChanges, snapshot)
	if err != nil {
		return nil, err
	}

	return &Ledger{
		InitialVersion: initialVersion,
		Changes:        changes,
		Snapshot:       snapshot,
	}, nil
}

// The Initial config is the Ledger with only the version entry and the first
// change set, encoded using base64.
func ParseLedger(s string) (*Ledger, error) {
	bs, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(s)
	if err != nil {
		return nil, err
	}

	return DecodeLedger(bs)
}

// Encodes the ledger into its binary CBOR representation.
func (l *Ledger) Encode() []byte {
	entries := make([][]byte, len(l.Changes)+1)

	// encode the initial version bytes
	entries[0] = l.InitialVersion.encode()

	for i, c := range l.Changes {
		entries[i+1] = c.Encode()
	}

	bs, err := cbor.Marshal(entries)
	if err != nil {
		panic(fmt.Sprintf("unable to encode Ledger (%v)", err))
	}

	return bs
}

func (l *Ledger) String() string {
	bs := l.Encode()

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bs)
}

func decodeLedgerVersion(bs []byte) (LedgerVersion, error) {
	var v LedgerVersion

	err := cbor.Unmarshal(bs, &v)
	if err != nil {
		return v, fmt.Errorf("invalid LedgerVersion encoding (%v)", err)
	}

	return v, nil
}

func (v LedgerVersion) encode() []byte {
	bs, err := cbor.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("unable to encode LedgerVersion %d (%v)", v, err))
	}

	return bs
}

func DecodeChangeSet(bs []byte, v LedgerVersion) (*ChangeSet, error) {
	ecs := new(encodeableChangeSet)

	if err := cbor.Unmarshal(bs, ecs); err != nil {
		return nil, fmt.Errorf("unable to decode change set (%v)", err)
	}

	return ecs.changeSet(v)
}

func (cs *ChangeSet) Encode() []byte {
	return newEncodeableChangeSet(cs).encode()
}

func newEncodeableChangeSet(cs *ChangeSet) encodeableChangeSet {
	prev, err := cs.Prev.encode()
	if err != nil {
		panic(fmt.Sprintf("invalid Prev ChangeSetID %s (%v)", cs.Prev, err))
	}

	actions := make([]encodeableAction, len(cs.Actions))

	for i, a := range cs.Actions {
		actions[i] = newEncodeableAction(a)
	}

	return encodeableChangeSet{
		Prev:       prev,
		Actions:    actions,
		Signatures: cs.Signatures,
	}
}

// Converts an encodeableChangeSet back into a regular ChangeSet.
func (ecs encodeableChangeSet) changeSet(v LedgerVersion) (*ChangeSet, error) {
	prev := decodeChangeSetID(ecs.Prev)
	actions := make([]Action, len(ecs.Actions))

	for i, ea := range ecs.Actions {
		a, err := ea.action(v)
		if err != nil {
			return nil, fmt.Errorf("unable to decode action %d of change set with prev=%s (%v)", i, prev, err)
		}

		actions[i] = a
	}

	return &ChangeSet{
		Prev:       prev,
		Actions:    actions,
		Signatures: ecs.Signatures,
	}, nil
}

func (ecs encodeableChangeSet) encode() []byte {
	bs, err := cbor.Marshal(ecs)
	if err != nil {
		panic(fmt.Sprintf("unable to encode change set (%v)", err))
	}

	return bs
}

// `decodeChangeSetID()` converts a bytestring into a human readable string.
func decodeChangeSetID(id []byte) ChangeSetID {
	if len(id) == 0 {
		return ChangeSetID("")
	} else {
		// Confusingly, `encodeBech32()` is used internally.
		return ChangeSetID(encodeBech32(ChangeSetIDPrefix, id))
	}
}

func (id ChangeSetID) encode() ([]byte, error) {
	// special case for first ChangeSet.Prev
	if id == "" {
		return []byte{}, nil
	}

	// Confusingly, `decodeBech32()` is used internally.
	prefix, bs, err := decodeBech32(string(id))
	if err != nil {
		return nil, fmt.Errorf("invalid bech32 ChangeSetID %s (%v)", id, err)
	}

	if prefix != ChangeSetIDPrefix {
		return nil, fmt.Errorf("invalid bech32 ChangeSetID prefix in %s (expected %q)", id, ChangeSetIDPrefix)
	}

	return bs, nil
}

func decodeBech32(s string) (string, []byte, error) {
	s = strings.TrimSpace(s)

	prefix, bs5, err := bech32.Decode(s)
	if err != nil {
		return "", nil, fmt.Errorf("invalid bech32 string %s (%v)", s, err)
	}

	bs, err := bech32.ConvertBits(bs5, 5, 8, false)
	if err != nil {
		panic(fmt.Sprintf("unable to convert bech32 bits (%v)", err))
	}

	return prefix, bs, nil
}

func encodeBech32(prefix string, bs []byte) string {
	bs5, err := bech32.ConvertBits(bs, 8, 5, true)
	if err != nil {
		panic(fmt.Sprintf("unable to convert bits of %x into bech32 form (%v)", bs, err))
	}

	s, err := bech32.Encode(prefix, bs5)
	if err != nil {
		panic(fmt.Sprintf("unable to encode %x into bech32 string with prefix %s (%v)", bs, prefix, err))
	}

	return s
}

// The raw CBOR encoding of an Action actually only represents its attributes.
// Those CBOR bytes are then wrapped in a map including the action category,
// name (see `encodeableAction`). Decoding that map in turn gives the action
// category and name needed to find the action attribute decoder function. The
// action attribute decoder function can then decode those attributes directly
// into the original concrete Action type.
func newActionDecoder[A Action]() actionDecoder {
	return func(attrBytes []byte) (Action, error) {
		var a A
		err := cbor.Unmarshal(attrBytes, &a)
		return a, err
	}
}

// A collection of all actionDecoders, as a triple nested map:
//
//  1. action category
//  2. action name
//  3. ledger version
//
// If a decoder for the current ledger version isn't available, the previous
// available version is used (e.g. if the current ledger version is 3, but the
// only available version of the given action is 1, then version 1 is used).
var actionDecoders = map[string]map[string]map[LedgerVersion]actionDecoder{
	FunctionsCategory: {
		AddFunctionName: {
			1: newActionDecoder[AddFunction](),
		},
		RemoveFunctionName: {
			1: newActionDecoder[RemoveFunction](),
		},
	},
	GatewaysCategory: {
		AddGatewayName: {
			1: newActionDecoder[AddGateway](),
		},
		AddGatewayEndpointName: {
			1: newActionDecoder[AddGatewayEndpoint](),
		},
		RemoveGatewayName: {
			1: newActionDecoder[RemoveGateway](),
		},
	},
	NodesCategory: {
		AddNodeName: {
			1: newActionDecoder[AddNode](),
		},
		RemoveNodeName: {
			1: newActionDecoder[RemoveNode](),
		},
	},
	PermissionsCategory: {
		AddUserName: {
			1: newActionDecoder[AddUser](),
		},
	},
}

func decodeAction(bs []byte, v LedgerVersion) (Action, error) {
	// If the encodeableAction CBOR changes from one ledger version to another,
	// then switch statement must be placed here to select encodeableActionV1,
	// encodeableActionV2 etc.
	ea := new(encodeableAction)

	err := cbor.Unmarshal(bs, ea)
	if err != nil {
		return nil, err
	}

	return ea.action(v)
}

func encodeAction(action Action) []byte {
	return newEncodeableAction(action).encode()
}

func newEncodeableAction(action Action) encodeableAction {
	category := action.Category()
	name := action.Name()

	attr, err := cbor.Marshal(action)
	if err != nil {
		panic(fmt.Sprintf("unable to make the %s:%s action encodeable (%v)", category, name, err))
	}

	return encodeableAction{
		Category:   category,
		Name:       name,
		Attributes: attr,
	}
}

func (ea encodeableAction) action(v LedgerVersion) (Action, error) {
	category := ea.Category
	name := ea.Name
	attrBytes := ea.Attributes

	categoryDecoders, ok := actionDecoders[category]
	if !ok {
		return nil, fmt.Errorf("unable to decode %s:%s, category %s not found", category, name, category)
	}

	versionedDecoders, ok := categoryDecoders[name]
	if !ok {
		return nil, fmt.Errorf("unable to decode %s:%s, action %s not found", category, name, name)
	}

	// Use the decoder version closest to `v`.
	for i := v; i >= 1; i-- {
		if decoder, ok := versionedDecoders[i]; ok {
			return decoder(attrBytes)
		}
	}

	// This situation shouldn't possible because at least one decoder should be
	// defined per action.
	panic(fmt.Sprintf("no decoders defined for %s:%s", category, name))
}

func (ae encodeableAction) encode() []byte {
	bytes, err := cbor.Marshal(ae)
	if err != nil {
		panic(fmt.Sprintf("unable to encode action (%v)", err))
	}

	return bytes
}
