package capabilities

import (
	"sync"
)

type EventType uint32

// Mointor event struct
type Event struct {
	// Event type
	Type EventType
	// Max RMID of this event can supported
	MaxRmid uint32
	// Scale Factor the montor event being used
	ScaleFactor uint32
}

// Monitor capability
type MonitorCapability struct {
	MaxRmid uint32
	// Event number can be monitor
	EventNum uint32
	// Event list
	Events []*Event
}

// L3CAT capability
type L3CAT struct {
	// Number of CLOS
	NumCLOS uint32
	// Number of cache way
	NumWays uint32
	// Size of cache way
	WaySize uint32
	// Way contention
	WayContention uint32
	// CDP
	CDP bool
}

// L2CAT capability
type L2CAT struct {
	// Number of CLOS
	NumCLOS uint32
	// Number of cache way
	NumWays uint32
	// Size of cache way
	WaySize uint32
	// Way contention
	WayContention uint32
}

// RDT capabilites on host
type Capabilities struct {
	Mon   *MonitorCapability
	L3Cat *L3CAT
	L2Cat *L2CAT
}

var once sync.Once
var lock sync.Mutex
var capabilities *Capabilities

func Initialize(c Capabilities) {
	once.Do(func() {
		capabilities = &c
	})
}

func Setup(mon *MonitorCapability, l3cat *L3CAT, l2cat *L2CAT) {
	Initialize(Capabilities{
		Mon:   mon,
		L3Cat: l3cat,
		L2Cat: l2cat,
	})
}

func Get() Capabilities {
	lock.Lock()
	defer lock.Unlock()
	if capabilities == nil {
		Initialize(Capabilities{
			Mon:   &MonitorCapability{},
			L3Cat: &L3CAT{},
			L2Cat: &L2CAT{},
		})
	}
	return *capabilities
}
