package sd

import (
	"lollipop/pkg/config"
	"lollipop/pkg/registry"
)

// RegisterSubscriberFactory registers the received factory
// Deprecated: RegisterSubscriberFactory. Use the GetRegister function
func RegisterSubscriberFactory(name string, sf SubscriberFactory) error {
	return subscriberFactories.Register(name, sf)
}

// GetSubscriber returns a subscriber from package registry
// Deprecated: GetSubscriber. Use the GetRegister function
func GetSubscriber(cfg *config.Backend) Subscriber {
	return subscriberFactories.Get(cfg.SD)(cfg)
}

// GetRegister returns the package registry
func GetRegister() *Register {
	return subscriberFactories
}

// Register is a SD registry
type Register struct {
	data registry.Untyped
}

func initRegister() *Register {
	return &Register{registry.NewUntyped()}
}

// Register implements the RegisterSetter interface
func (r *Register) Register(name string, sf SubscriberFactory) error {
	r.data.Register(name, sf)
	return nil
}

// Get implements the RegisterGetter interface
func (r *Register) Get(name string) SubscriberFactory {
	tmp, ok := r.data.Get(name)
	if !ok {
		return FixedSubscriberFactory
	}
	sf, ok := tmp.(SubscriberFactory)
	if !ok {
		return FixedSubscriberFactory
	}
	return sf
}

var subscriberFactories = initRegister()
