// This package is responsible for reading, writing, encoding, decoding, and 
// validating OWS ledgers.
//
// Validating requires maintaining a snapshot of the ledger state, from which 
// other convenient information can easily be derived (e.g. listing all the 
// current nodes).
//
// The ledger uses CBOR for its binary encoding, and consists of an initial 
// version number followed by a list of change sets. Each change set has a 
// unique id, derived from its blake2b-128 hash, and contains a reference to
// its preceding change set, a list of actions, and a list of signatures.
//
// The first change set is called the *initial configuration*, and its id is
// used as the project id.
//
// The ledger can modify its encoding format based on ledger version. This way
// ledgers can be upgraded without losing backward compatibility.
package ledger

import (
	"fmt"
	"os"
	"path"

	"github.com/google/uuid"
)

// When starting a new project, the ledger initial config is passed to the nodes
// and clients via an environemnt variable with the following name:
const InitialConfigEnvName = "OWS_INITIAL_CONFIG"

// The Ledger is a journal of all the OWS infrastructure configuration changes.
//
// The InitialVersion field is needed to be able to perpetually encode the
// ledger in the same way. The current ledger version can differ from the
// InitialVersion.
//
// The first entry in the Changes list is the initial configuration of a
// project. The latter entries are actual configuration changes.
//
// The snapshot field isn't exported because it might be nil. The snapshot is
// generated on-demand by calling the `Ledger.Snapshot()` method.
type Ledger struct {
	InitialVersion LedgerVersion
	Changes  []ChangeSet
	snapshot *Snapshot
}

// Every new LedgerVersion introduces breaking changes, so a single major
// version number is sufficient to describe it.
//
// The LedgerVersion starts at 1.
type LedgerVersion uint

const LatestLedgerVersion = 1

// For convenience, the first change set (i.e. the initial configuration) and
// latter change sets use the same structure. The `Prev`` ChangeSetID of the 
// first change set is the empty string.
//
// ChangeSet signatures are based on the CBOR encoded ChangeSet without the
// signatures field itself. Signatures are validated against previously defined
// user permissions. The signers of the first change set are the root users,
// and can submit any action without restriction perpetually.
type ChangeSet struct {
	Prev       ChangeSetID
	Actions    []Action
	Signatures []Signature
}

// The ChangeSetID is formed by hashing a CBOR endoded ChangeSet using
// blake2b-128, and then encoding the hash using Bech32 with the "changes"
// prefix.
//
// The "project" prefix is used for the first change set (i.e. the initial
// configuration), and the "changes" prefix is used for other change sets.
type ChangeSetID string

const ChangeSetIDPrefix = "changes"

// The ProjectID uses the underlying bytes of the ChangeSetID of the first
// change set (i.e. the initial configuration), with the "project" prefix.
type ProjectID string

const ProjectIDPrefix = "project"

// There are a large number action types. Each actions is responsible for their
// application to the ledger and must implement the Action interface.
//
// Each action must have a unique name per category.
//
// Valid action categories are:
//    - functions
//    - gateways
//    - nodes
//    - permissions
//    - ...
//
// An action operates on resources, adding/removing or changing them. The 
// `Resources` getter makes it easier to check resource-specific permissions
// during ledger validation.
//
// The Apply() method makes it more manageable to update the project
// configuration without inferring specific action types. The 
// `ResourceIDGenerator` creates unique resource ids.
//
// TODO: should an Unapply() method be added? (for easy rollbacks)
type Action interface {
	Category() string
	Name() string
	Resources() []ResourceID

	Apply(s *Snapshot, genID ResourceIDGenerator) error
}

// Every newly created resource is given a deterministic id.
//
// The resource id is calculated as follows:
//    1. Take the hash of the previous change set (empty bytestring if the
//       resource is defined in the initial configuration).
//    2. Take the little endian encoding of the action index.
//    3. Hash the concatenation of the bytestrings from step 1 and 2 using
//       [Blake2b](https://en.wikipedia.org/wiki/BLAKE_(hash_function))-128).
//    4. Generate a string with human-readable prefix using [Bech32](https://en.bitcoin.it/wiki/Bech32).
//       A different prefix is used for each resource type.
type ResourceID string

type ResourceIDGenerator = func(prefix string) ResourceID

type FunctionID = ResourceID
type GatewayID = ResourceID
type NodeID = ResourceID
type PolicyID = ResourceID
type UserID = ResourceID

const (
	FunctionIDPrefix = "fn"
	GatewayIDPrefix = "gateway"
	NodeIDPrefix = "node"
	PolicyIDPrefix = "policy"
	UserIDPrefix = "user"
)

// Some resources, like serverless functions, require files to operate. In OWS,
// these files are referred to as *assets*.
//
// Although assets are referenced in the ledger by some actions, they aren't
// actually defined in the ledger. Hence, assets are treated differently from
// resources, and have a differnt id type: `AssetID`.
//
// An AssetID is calculated by taking the blake2b-128 hash of the asset, and
// encoding that hash using Bech32 with the "asset" prefix.
type AssetID string

const AssetIDPrefix = "asset"

type Port uint16

type FunctionConfig struct {
	Runtime string
	HandlerID AssetID
}

type GatewayConfig struct {
	Port Port
	Endpoints []GatewayEndpointConfig
}

type GatewayEndpointConfig struct {
	Method string
	Path string
	FunctionID FunctionID
}

type NodeConfig struct {
	Key     PublicKey
	Address string 
	GossipPort Port
	SyncPort   Port
}

// Each user attached policy is independent and doesn't impact other policies
// in the list. The final user permissions are formed by the union of all
// attached policies.
type UserConfig struct {
	Key      PublicKey
	IsRoot   bool
	Policies []PolicyID
}

func NewLedger(v LedgerVersion, initialConfig *ChangeSet) (*Ledger, error) {
	l := &Ledger{v, []ChangeSet{*initialConfig}, nil}

	if err := l.Validate(); err != nil {
		return nil, err
	}

	return l, nil
}

// Reads, decodes and validates the ledger located at the path.
// If there is no file is found at `path`, an os.ErrNotExist error is returned,
// which can be used to write the ledger with only the initial config to disk.
func ReadLedger(path string) (*Ledger, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return DecodeLedger(bs)
}

// Validates and appends a change set
func (l *Ledger) Append(cs *ChangeSet) error {
	snapshot := l.Snapshot()

	if err := validateChangeSet(cs, snapshot); err != nil {
		return err
	}

	l.Changes = append(l.Changes, *cs)

	return nil
}

func (l *Ledger) Head() ChangeSetID {
	return l.Snapshot().Head
}

// Creates an unsigned change set.
func (l *Ledger) NewChangeSet(actions ...Action) *ChangeSet {
	cs := &ChangeSet{
		Prev:     l.Head(),
		Actions:    actions,
		Signatures: []Signature{},
	}

	return cs
}

// Returns the snapshot.
// If the snapshot is nil, the snapshot is recreated by revalidating the ledger.
// Validation errors are however not expected at this point, so they result in
// a panic.
func (l *Ledger) Snapshot() *Snapshot {
	if l.snapshot != nil {
		return l.snapshot
	}

	if err := l.Validate(); err != nil {
		panic(fmt.Sprintf("revalidation failed (%v)", err))
	}

	return l.snapshot
}

// Encodes and writes the ledger to disk.
func (l *Ledger) Write(path string) error {
	bs := l.Encode()

	return WriteSafe(path, bs)
}

// Creates any necessary parent directories, then writes a temporary file, and
// finally move the temporary file to the given path location.
func WriteSafe(p string, bs []byte) error {
	d := path.Dir(p)

	if err := os.MkdirAll(d, 0755); err != nil {
		return err
	}

	tmpFileName := uuid.NewString()
	tmpPath := path.Join(d, tmpFileName)

	if err := os.WriteFile(tmpPath, bs, 0644); err != nil {
		return fmt.Errorf("unable to write %s, tmp file creation failed (%v)", p, err)
	}

	if err := os.Rename(tmpPath, p); err != nil {
		return fmt.Errorf("unable to write %s, tmp file movement failed (%v)", p, err)
	}

	return nil
}

func (cs *ChangeSet) apply(s *Snapshot) error {
	for i, a := range cs.Actions {
		if err := a.Apply(s, newResourceIDGenerator(cs.Prev, uint(i))); err != nil {
			return fmt.Errorf("failed to validate action %d of change set with prev=%s (%v)", i, cs.Prev, err)
		}
	}

	return nil
}