package datastore

type Datastore interface {
	Start() error
	Stop() error
	Prices() PriceStore
}
