package sd

import (
  "errors"
  "github.com/valyala/fastrand"
  "runtime"
  "sync/atomic"
)

// Balancer apply a balancing strategy in order to select the backend host to be used
type Balancer interface {
  Host() (string, error)
}

// ErrNoHosts is the error the balancer must return when there are 0 hosts ready
var ErrNoHosts = errors.New("no hosts available")

// NewBalancer returns the best perfomant balancer depending on the number of available processors.
// If GOMAXPROCS = 1, it returns a round robin LB due there is no contention over the atomic counter.
// If GOMAXPROCS > 1, it returns a pseudo random LB optimized for scaling over the number of CPUs.
func NewBalancer(subscriber Subscriber) Balancer {
  if p := runtime.GOMAXPROCS(-1); p == 1 {
    return NewRoundRobinLB(subscriber)
  }
  return NewRandomLB(subscriber)
}

// NewRoundRobinLB returns a new balancer using a round robin strategy
func NewRoundRobinLB(subscriber Subscriber) Balancer {
  if s, ok := subscriber.(FixedSubscriber); ok && len(s) == 1 {
    return nopBalancer(s[0])
  }
  return &roundRobinLB{
    balancer: balancer{subscriber: subscriber},
    counter:  0,
  }
}

type roundRobinLB struct {
  balancer
  counter uint64
}

// Host implements the balancer interface
func (r *roundRobinLB) Host() (string, error) {
  hosts, err := r.hosts()
  if err != nil {
    return "", err
  }
  offset := (atomic.AddUint64(&r.counter, 1) - 1) % uint64(len(hosts))
  return hosts[offset], nil
}

// NewRandomLB returns a new balancer using a fastrand pseudorandom generator
func NewRandomLB(subscriber Subscriber) Balancer {
  if s, ok := subscriber.(FixedSubscriber); ok && len(s) == 1 {
    return nopBalancer(s[0])
  }
  return &randomLB{
    balancer: balancer{subscriber: subscriber},
    rand:     fastrand.Uint32n,
  }
}

type randomLB struct {
  balancer
  rand func(uint32) uint32
}

// Host implements the balancer interface
func (r *randomLB) Host() (string, error) {
  hosts, err := r.hosts()
  if err != nil {
    return "", err
  }
  return hosts[int(r.rand(uint32(len(hosts))))], nil
}

type balancer struct {
  subscriber Subscriber
}

func (b *balancer) hosts() ([]string, error) {
  hs, err := b.subscriber.Hosts()
  if err != nil {
    return hs, err
  }
  if len(hs) <= 0 {
    return hs, ErrNoHosts
  }
  return hs, nil
}

type nopBalancer string

func (b nopBalancer) Host() (string, error) { return string(b), nil }

//// NewRoundRobinLB returns a new balancer using a round robin strategy
//func NewRoundRobinLB(subscriber Subscriber) Balancer {
//  return &roundRobinLB{
//    subscriber: subscriber,
//    counter:    0,
//  }
//}
//
//type roundRobinLB struct {
//  subscriber Subscriber
//  counter    uint64
//}
//
//// Host implements the balancer interface
//func (rr *roundRobinLB) Host() (string, error) {
//  hosts, err := rr.subscriber.Hosts()
//  if err != nil {
//    return "", err
//  }
//  if len(hosts) <= 0 {
//    return "", ErrNoHosts
//  }
//  offset := (atomic.AddUint64(&rr.counter, 1) - 1) % uint64(len(hosts))
//  return hosts[offset], nil
//}
//
//// NewRandomLB returns a new balancer using a pseudo-random strategy
//func NewRandomLB(subscriber Subscriber, seed int64) Balancer {
//  return &randomLB{
//    subscriber: subscriber,
//    rnd:        rand.New(rand.NewSource(seed)),
//  }
//}
//
//type randomLB struct {
//  subscriber Subscriber
//  rnd        *rand.Rand
//}
//
//// Host implements the balancer interface
//func (r *randomLB) Host() (string, error) {
//  hosts, err := r.subscriber.Hosts()
//  if err != nil {
//    return "", err
//  }
//  if len(hosts) <= 0 {
//    return "", ErrNoHosts
//  }
//  return hosts[r.rnd.Intn(len(hosts))], nil
//}
